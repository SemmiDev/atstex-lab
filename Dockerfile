# ── Dockerfile.tectonic ──────────────────────────────────────────────────────
# Alternative image using Tectonic instead of pdflatex/xelatex/lualatex.
# Tectonic auto-downloads only the TeX packages each document actually uses,
# producing an image of ~150–200 MB (vs ~8 GB with full TeX Live).
#
# Trade-offs:
#   ✓  Much smaller image
#   ✓  No manual package management (auto-fetch from CTAN bundle)
#   ✓  Single, modern engine with PDF output
#   ✗  No xelatex / lualatex engine selection
#   ✗  First compile of a new document is slower (package download)
#
# Usage:
#   docker compose -f compose.yml -f compose.tectonic.yml up -d --build
# ─────────────────────────────────────────────────────────────────────────────

# ── Stage 1: Build Go binary ────────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o atstex-lab ./cmd/server/main.go


# ── Stage 2: Runtime with Tectonic ──────────────────────────────────────────
FROM alpine:3.21

# Tectonic is available in Alpine community repos
RUN apk add --no-cache tectonic libstdc++ fontconfig freetype

# Persistent package cache directory
RUN mkdir -p /var/cache/tectonic && chmod 777 /var/cache/tectonic
ENV XDG_CACHE_HOME=/var/cache/tectonic

# Pre-download the massive Tectonic package bundle during image build
# so the first user request doesn't timeout after 30s. We include all packages
# and fonts used by the CV templates here so they get cached inside the image.
RUN echo '\documentclass{article}' > /tmp/dummy.tex && \
    echo '\usepackage[T1]{fontenc}' >> /tmp/dummy.tex && \
    echo '\usepackage[utf8]{inputenc}' >> /tmp/dummy.tex && \
    echo '\usepackage[english]{babel}' >> /tmp/dummy.tex && \
    echo '\usepackage{geometry,setspace,fancyhdr,array,tabularx,booktabs,longtable}' >> /tmp/dummy.tex && \
    echo '\usepackage{enumitem,xcolor,graphicx,hyperref,microtype,titlesec}' >> /tmp/dummy.tex && \
    echo '\usepackage{multicol,etoolbox,latexsym,marvosym,verbatim}' >> /tmp/dummy.tex && \
    echo '\usepackage{helvet,mathptmx,palatino,courier,lmodern}' >> /tmp/dummy.tex && \
    echo '\usepackage{amsmath,amssymb,amsfonts,amsthm}' >> /tmp/dummy.tex && \
    echo '\usepackage{fontspec}' >> /tmp/dummy.tex && \
    echo '\begin{document}Hello — World\end{document}' >> /tmp/dummy.tex && \
    tectonic /tmp/dummy.tex && \
    rm -f /tmp/dummy.*

# Non-root user
RUN adduser -D -u 1001 atstex-lab && \
    chown -R atstex-lab:atstex-lab /var/cache/tectonic

WORKDIR /home/atstex-lab

COPY --from=builder /build/atstex-lab /usr/local/bin/atstex-lab
COPY .env /home/atstex-lab/.env

USER atstex-lab

EXPOSE 8080

ENV ADDR=:8080

ENTRYPOINT ["/usr/local/bin/atstex-lab"]
