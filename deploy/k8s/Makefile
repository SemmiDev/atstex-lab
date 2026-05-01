# ==========================================
# atstex-lab  —  Kubernetes helpers
# ==========================================
# Usage:
#   make create-secrets \
#     POSTGRES_PASSWORD=supersecret \
#     GOOGLE_CLIENT_ID=xxx \
#     GOOGLE_CLIENT_SECRET=xxx \
#     AI_API_KEY=xxx
#
#   make apply
#   make delete
# ==========================================

NAMESPACE       := atstex-lab
SECRET_NAME     := atstex-secrets
POSTGRES_USER   := atstex
POSTGRES_DB     := atstex

# --- Secret values (pass via CLI) ---------
POSTGRES_PASSWORD   ?= CHANGE_ME
GOOGLE_CLIENT_ID    ?= CHANGE_ME
GOOGLE_CLIENT_SECRET?= CHANGE_ME
AI_API_KEY          ?= CHANGE_ME

DATABASE_URL := postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@postgres-svc:5432/$(POSTGRES_DB)?sslmode=disable

# ===========================================
# Targets
# ===========================================

.PHONY: create-secrets apply delete status

## Inject secrets into the cluster (idempotent)
create-secrets:
	@echo "🔐 Creating/updating secret '$(SECRET_NAME)' in namespace '$(NAMESPACE)' ..."
	kubectl create secret generic $(SECRET_NAME) \
	  --namespace $(NAMESPACE) \
	  --from-literal=POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
	  --from-literal=DATABASE_URL="$(DATABASE_URL)" \
	  --from-literal=GOOGLE_CLIENT_ID=$(GOOGLE_CLIENT_ID) \
	  --from-literal=GOOGLE_CLIENT_SECRET=$(GOOGLE_CLIENT_SECRET) \
	  --from-literal=AI_API_KEY=$(AI_API_KEY) \
	  --dry-run=client -o yaml | kubectl apply -f -
	@echo "✅ Secret created/updated."

## Apply all manifests in order
apply:
	kubectl apply -f 00-namespace.yaml
	kubectl apply -f 02-configmap.yaml
	kubectl apply -f 03-pv.yaml
	kubectl apply -f 04-pvc.yaml
	kubectl apply -f 05-postgres.yaml
	kubectl apply -f 06-backend.yaml
	kubectl apply -f 07-ingress.yaml
	@echo "✅ All manifests applied."

## Delete everything in the namespace
delete:
	kubectl delete namespace $(NAMESPACE) --ignore-not-found
	kubectl delete pv postgres-pv tectonic-cache-pv --ignore-not-found
	@echo "🗑️  Namespace and PVs deleted."

## Show status
status:
	@echo "=== Pods ==="
	kubectl get pods -n $(NAMESPACE)
	@echo ""
	@echo "=== Services ==="
	kubectl get svc -n $(NAMESPACE)
	@echo ""
	@echo "=== Secrets ==="
	kubectl get secrets -n $(NAMESPACE)
	@echo ""
	@echo "=== PVC ==="
	kubectl get pvc -n $(NAMESPACE)
	@echo ""
	@echo "=== Ingress ==="
	kubectl get ingress -n $(NAMESPACE)
