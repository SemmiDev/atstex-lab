# ── Stage 1: Build Go binary ─────────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

WORKDIR /build

# Cache dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a static binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o atstex-lab ./cmd/server/main.go


# ── Stage 2: Runtime — TeX Live full + our binary ────────────────────────────
# texlive/texlive:latest is the official image with TeX Live full (~4 GB).
# It is based on Debian and contains pdflatex, xelatex, lualatex, latexmk,
# plus virtually every CTAN package.
FROM texlive/texlive:latest

# Non-root user for safety
RUN useradd -m -u 1001 atstex-lab

WORKDIR /home/atstex-lab

# Copy compiled binary from builder stage
COPY --from=builder /build/atstex-lab /usr/local/bin/atstex-lab

# Copy .env file from host
COPY .env /home/atstex-lab/.env

# The binary embeds all web assets via go:embed, so nothing else to copy.

USER atstex-lab

EXPOSE 8080

ENV ADDR=:8080

ENTRYPOINT ["/usr/local/bin/atstex-lab"]
