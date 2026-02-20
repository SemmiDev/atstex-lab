.PHONY: run build tidy clean docker-build docker-up docker-down docker-logs docker-run

BINARY := latexpad
CMD    := ./cmd/server

# ── Local development ─────────────────────────────────────────
run:
	go run $(CMD)/main.go

build:
	go build -o $(BINARY) $(CMD)/main.go

tidy:
	go mod tidy

clean:
	rm -f $(BINARY)

# ── Docker ────────────────────────────────────────────────────
# Build image (first time: ~4 GB pull for texlive/texlive:latest)
docker-build:
	docker compose build

# Start container in background
docker-up:
	docker compose up -d

# Build + start in one step
docker-run:
	docker compose up -d --build

# Stop container
docker-down:
	docker compose down

# Tail logs
docker-logs:
	docker compose logs -f latexpad

# ── Install LaTeX locally (Ubuntu/Debian) ─────────────────────
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
