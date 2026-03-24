.PHONY: run build tidy clean css docker-build docker-up docker-down docker-logs docker-run docker-remove-rebuild install-latex migrateup migratedown migrateup1 migratedown1 new_migration

BINARY := atstex-lab
CMD    := ./cmd/server
DB_URL := postgres://atstex:password@localhost:5432/atstex?sslmode=disable

## ── Prerequisites ────────────────────────────────────────────
## Go 1.22+           → https://go.dev/dl/
## Node.js + npm      → needed for Tailwind CSS compilation
## Docker + Compose   → for running with PostgreSQL and TeX Live
## TeX Live           → only if running locally without Docker

# ── CSS ───────────────────────────────────────────────────────
# Compile Tailwind CSS into a single minified stylesheet.
css:
	npx tailwindcss -i ./web/static/css/tailwind.css -o ./web/static/css/style.css --minify

# ── Local Development ─────────────────────────────────────────
# Run the dev server. Compiles CSS first, then starts on :8080.
run: css
	go run $(CMD)/main.go

# Build a production binary. Output: ./atstex-lab
build: css
	go build -o $(BINARY) $(CMD)/main.go

tidy:
	go mod tidy

clean:
	rm -f $(BINARY)

# ── Docker (recommended) ─────────────────────────────────────
# Build the Docker image. Compiles CSS first since it gets embedded in the binary.
docker-build: css
	docker compose build

# Start containers in background (app + PostgreSQL).
docker-up: css
	docker compose up -d

# Build + start in one step.
docker-run: css
	docker compose up -d --build

# Tear down containers and delete the database volume.
docker-remove-rebuild: css
	docker compose down --volumes
	docker compose up -d --build

# Stop containers.
docker-down:
	docker compose down

# Tail app logs.
docker-logs:
	docker compose logs -f atstex-lab

# ── Install TeX Live locally (Ubuntu/Debian) ─────────────────
install-latex:
	sudo apt-get update
	sudo apt-get install -y \
		texlive-latex-base \
		texlive-latex-recommended \
		texlive-latex-extra \
		texlive-fonts-recommended \
		texlive-fonts-extra \
		texlive-science \
		texlive-xetex \
		texlive-luatex \
		latexmk

# ── Database Migrations ───────────────────────────────────────
migrateup:
	migrate -path db/migration -database "$(DB_URL)" -verbose up

migrateup1:
	migrate -path db/migration -database "$(DB_URL)" -verbose up 1

migratedown:
	migrate -path db/migration -database "$(DB_URL)" -verbose down

migratedown1:
	migrate -path db/migration -database "$(DB_URL)" -verbose down 1

new_migration:
	migrate create -ext sql -dir db/migration -seq $(name)
