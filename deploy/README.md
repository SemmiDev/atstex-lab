# ATSTEX-LAB — Ansible Deployment 🚀

This directory contains the complete Infrastructure-as-Code (IaC) setup for deploying **ATSTEX-LAB** to a production Ubuntu server using Ansible.

It automates the installation of Docker, Nginx (with Let's Encrypt SSL), PostgreSQL, and handles the building, deploying, and rollback of the Go application.

---

## 🏗 Architecture

- **OS:** Ubuntu Linux (Hardened with UFW firewall, Fail2Ban, Swap file, unattended-upgrades)
- **Proxy:** Nginx with Let's Encrypt SSL auto-renewal
- **Database:** PostgreSQL containerized, mapped only to `localhost`
- **Application:** Multi-stage Go Docker container mapping to port `8080`

---

## 📁 Directory Structure

```text
deploy/
├── ansible.cfg              # Ansible configuration tweaks
├── group_vars/
│   ├── all.yml              # Global settings (app name, ports)
│   └── production.yml       # Production secrets, domains, CPU limits
├── inventory/
│   └── production.ini       # Server IP and SSH connection details
├── playbooks/
│   ├── site.yml             # Main playbook: Executes full server setup
│   ├── setup.yml            # Infrastructure only (no app deploy)
│   └── deploy.yml           # App deployment & updates only
├── roles/                   # Ansible roles (common, docker, nginx, postgres, app)
└── scripts/                 # Utility shell scripts installed to /opt/atstex-lab/scripts
```

---

## 🔐 Managing Secrets Safely for GitHub (Ansible Vault)

Your `group_vars/production.yml` contains sensitive information (database passwords, Google OAuth secrets, AI API keys). **Never push plaintext secrets to a public GitHub repository.**

Instead, use **Ansible Vault** to encrypt the file before committing:

### 1. Encrypt the file
```bash
cd deploy/
ansible-vault encrypt group_vars/production.yml
```
*You will be prompted to create a vault password. Remember this password!*

### 2. Edit the encrypted file
If you need to change a variable later, do not decrypt it completely. Use the edit command:
```bash
ansible-vault edit group_vars/production.yml
```

### 3. Run playbooks with Vault
When running deployment commands manually, you must tell Ansible to ask for the vault password:
```bash
ansible-playbook -i inventory/production.ini playbooks/site.yml --ask-vault-pass
```
*Note: The `./scripts/deploy.sh` wrapper script handles this automatically if it detects an encrypted file.*

---

## 🚀 Deployment Guide

### Prerequisites
1. You must have Ansible installed on your local control machine:
   ```bash
   brew install ansible         # macOS
   sudo apt install ansible     # Ubuntu/Debian
   ```
2. Your local SSH public key (`~/.ssh/id_rsa.pub` or `~/.ssh/id_ed25519.pub`) must be added to the remote server's `root` user initially.

### Step 1: Configuration
1. Edit `inventory/production.ini` and set your server's public IP address.
2. Edit `group_vars/production.yml` (using `ansible-vault edit` if encrypted) and ensure your domain name, email, and API keys are correct.

### Step 2: Initial Server Setup & Deploy
To provision a brand new server from scratch and deploy the application for the first time, use the wrapper script from the **project root**:

```bash
cd deploy/
./scripts/deploy.sh --env production
```
This will run the `site.yml` playbook, configuring the OS, installing Docker/Nginx/PostgreSQL, getting SSL certificates, and building the Go app.

### Step 3: Pushing App Updates
When you modify your Go code or HTML templates and want to push the update to the server, you do **not** need to run the full setup again. Just run the application deployment playbook:

```bash
cd deploy/
./scripts/deploy.sh --env production --app-only
```
This skips the infrastructure setup, builds the new Docker image, tags the old one as a rollback, and restarts the container with zero-downtime Nginx proxying.

---

## 🛠 Server Utility Scripts

During setup, Ansible installs several handy helper scripts on the remote server located at `/opt/atstex-lab/scripts/`.

You can access these by SSHing into the server:
```bash
ssh deploy@<YOUR_SERVER_IP>
cd /opt/atstex-lab/scripts/
```

### 1. Instant Rollbacks
If a new deployment breaks the application, you can instantly revert to the previous Docker container image:
```bash
sudo ./rollback.sh
```

### 2. Database Backups
PostgreSQL backups are automatically run daily via cron. However, you can trigger a manual backup anytime:
```bash
sudo ./backup-db.sh
```
*Backups are stored as `.sql.gz` files in `/opt/atstex-lab/backups/`.*

### 3. Database Restoration
To restore the database from a specific backup file (it will automatically create a pre-restore safety backup first):
```bash
sudo ./restore-db.sh /opt/atstex-lab/backups/backup_YYYYMMDD_HHMMSS.sql.gz
```

### 4. Health Checks
Check the status of Docker, Nginx, PostgreSQL, memory, and disk space:
```bash
sudo ./health-check.sh
```

---

## 🛡 Security Notes
- The `common` Ansible role automatically hardens the server by disabling SSH password authentication and root login (forcing the use of the `deploy` user with SSH keys).
- UFW Firewall drops all traffic except ports 22 (SSH), 80 (HTTP), and 443 (HTTPS).
- Fail2Ban prevents brute-force SSH attacks.
- PostgreSQL port `5432` is intentionally NOT exposed to the public internet; it only binds to Docker's internal networking (`127.0.0.1`).
