# ATSTEX-LAB

Write your biodata, pick a LaTeX template, compile to PDF.

---

## What It Does

- Fill in your resume data through a guided form (personal info, experience, education, skills, etc.)
- Save multiple CV profiles — one for "Back End Developer", another for "Data Science", etc.
- Pick from a library of ATS-friendly LaTeX templates
- Compile to PDF instantly (pdflatex, xelatex, or lualatex)
- Preview the PDF side-by-side with your biodata form
- Sign in with Google to save your data across devices

---

## Prerequisites

| Tool | Why | Install |
|------|-----|---------|
| **Docker + Docker Compose** | Runs the app, PostgreSQL, and TeX Live in containers | [docker.com](https://docs.docker.com/get-docker/) |
| **Go 1.22+** | Only for local dev (not needed with Docker) | [go.dev](https://go.dev/dl/) |
| **Node.js + npm** | Compiles Tailwind CSS | [nodejs.org](https://nodejs.org/) |

---

## Running with Docker (recommended)

This is the easiest way. The container ships with a full TeX Live installation so you don't need to install anything else.

### 1. Set up environment variables

Copy and edit the `.env` file:

```bash
cp .env.example .env
```

Fill in:

```env
DATABASE_URL=postgres://postgres:postgres@postgres:5432/atstex_lab?sslmode=disable
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_CALLBACK_URL=http://localhost:8080/auth/google/callback

# AI Extraction Config
# Supported providers: openai (default), anthropic, googleai (gemini), mistral, ollama
AI_PROVIDER=openai
AI_MODEL=gpt-4o-mini
AI_API_KEY=your-api-key
# AI_BASE_URL=        # Optional: for OpenAI-compatible APIs (Groq, Together, etc.)
```

> To get Google OAuth credentials, go to [Google Cloud Console](https://console.cloud.google.com/apis/credentials) and create an OAuth 2.0 Client ID.

### 2. Start the app

```bash
make docker-run
```

This builds the image and starts both tthe app and PostgreSQL. First run pulls ~4 GB for the TeX Live image.

Open **http://localhost:8080**.

### Other commands

```bash
make docker-logs            # tail app logs
make docker-down            # stop containers
make docker-remove-rebuild  # nuke DB volume and rebuild from scratch
```

---

## Running Locally (without Docker)

You'll need Go, Node.js, a running PostgreSQL instance, and TeX Live installed on your machine.

### 1. Install TeX Live

```bash
# Ubuntu / Debian
make install-latex

# macOS
brew install --cask mactex
```

### 2. Set up the database

Create a PostgreSQL database and run the schema:

```bash
psql -U postgres -c "CREATE DATABASE atstex_lab;"
psql -U postgres -d atstex_lab -f init.sql
```

### 3. Configure `.env`

```env
DATABASE_URL=postgres://postgres:postgres@localhost:5432/atstex_lab?sslmode=disable
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_CALLBACK_URL=http://localhost:8080/auth/google/callback

# AI Extraction Config
AI_PROVIDER=openai
AI_MODEL=gpt-4o-mini
AI_API_KEY=your-api-key
```

### 4. Run

```bash
npm install        # install Tailwind CSS (first time only)
make run           # compiles CSS + starts dev server on :8080
```

### 5. Build a standalone binary

```bash
make build
./atstex-lab       # run the compiled binary
```

---

## Makefile Reference

| Command | What it does |
|---------|-------------|
| `make run` | Compile CSS + start dev server |
| `make build` | Compile CSS + build Go binary (`./atstex-lab`) |
| `make css` | Compile Tailwind CSS only |
| `make tidy` | Run `go mod tidy` |
| `make clean` | Delete the compiled binary |
| `make docker-run` | Build image + start app and PostgreSQL |
| `make docker-up` | Start containers (assumes image is built) |
| `make docker-down` | Stop containers |
| `make docker-remove-rebuild` | Delete everything and rebuild fresh |
| `make docker-logs` | Tail live logs |
| `make install-latex` | Install TeX Live on Ubuntu/Debian |

---

## Project Structure

```
atstex-lab/
├── cmd/server/main.go              # App entry point, routes, middleware
├── internal/
│   ├── auth/auth.go                # Google OAuth login/logout/callback
│   ├── compiler/compiler.go        # Runs LaTeX engine in a temp directory
│   ├── config/config.go            # Loads .env configuration
│   ├── cvtemplate/                  # LaTeX CV template loader
│   ├── domain/                     # Data models (User, Session, CVProfile)
│   ├── extractor/extractor.go      # AI-powered PDF resume extraction (multi-provider)
│   ├── handler/                    # HTTP handlers (pages + API)
│   └── repository/repository.go    # PostgreSQL queries
├── web/
│   ├── embed.go                    # Embeds templates + static files into binary
│   ├── templates/                  # HTML pages + LaTeX templates
│   └── static/                     # CSS, JS, images
├── init.sql                        # Database schema
├── Dockerfile                      # Multi-stage: Go build → TeX Live runtime
├── compose.yml                     # Docker Compose (app + PostgreSQL)
├── tailwind.config.js              # Tailwind CSS config
└── Makefile
```

---

## API Endpoints

### Pages

| Route | Description |
|-------|-------------|
| `GET /` | Home page |
| `GET /input` | Biodata form (full page) |
| `GET /input/embed` | Biodata form (embedded, no sidebar) |
| `GET /editor` | LaTeX editor + PDF preview |
| `GET /profile` | User profile + session management |

### Auth

| Route | Description |
|-------|-------------|
| `GET /auth/google/login` | Start Google login |
| `GET /auth/google/callback` | OAuth callback |
| `POST /auth/logout` | Logout |
| `POST /auth/sessions/{token}/delete` | Delete a session |

### CV Profiles

| Route | Description |
|-------|-------------|
| `GET /api/cv-profiles` | List all profiles |
| `POST /api/cv-profiles` | Create a profile (`{"title": "..."}`) |
| `GET /api/cv-profiles/{id}` | Get a profile + biodata |
| `PUT /api/cv-profiles/{id}` | Save biodata (`{"biodata": {...}}`) |
| `DELETE /api/cv-profiles/{id}` | Delete a profile |

### Templates & Compile

| Route | Description |
|-------|-------------|
| `GET /api/templates` | List available templates |
| `GET /api/templates/{name}` | Get raw template source |
| `POST /api/templates/{name}/render` | Render template with biodata JSON |
| `POST /api/extract-pdf` | Extract biodata from PDF text via AI |
| `POST /compile` | Compile LaTeX → PDF |
