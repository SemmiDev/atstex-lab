# ── Stage 1: Build Go binary (FAST + CACHED) ────────────────────────────────
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /build

# Better proxy + deterministic builds
ENV GOPROXY=https://proxy.golang.org,direct \
    CGO_ENABLED=0

# Copy deps first (cache friendly)
COPY go.mod go.sum ./

# 🔥 Use BuildKit cache (BIG WIN)
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

# Copy source
COPY . .

# Build binary (also cached)
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o atstex-lab ./cmd/server/main.go


# ── Stage 2: Runtime with Tectonic (LEAN + STABLE) ──────────────────────────
FROM alpine:3.21

# Install only required runtime deps
RUN apk add --no-cache \
    tectonic \
    libstdc++ \
    fontconfig \
    freetype \
    ca-certificates

# Cache directory (persist across container runs if mounted)
ENV XDG_CACHE_HOME=/var/cache/tectonic
RUN mkdir -p ${XDG_CACHE_HOME} && chmod 777 ${XDG_CACHE_HOME}

# 🔥 Prewarm Tectonic cache (but optimized layer)
RUN <<EOF
set -e
cat <<LATEX > /tmp/dummy.tex
\documentclass{article}
\usepackage[T1]{fontenc}
\usepackage[utf8]{inputenc}
\usepackage[english]{babel}
\usepackage{geometry,setspace,fancyhdr,array,tabularx,booktabs,longtable}
\usepackage{enumitem,xcolor,graphicx,hyperref,microtype,titlesec}
\usepackage{multicol,etoolbox,latexsym,marvosym,verbatim,tikz}
\usepackage{helvet,mathptmx,palatino,courier,lmodern}
\usepackage{amsmath,amssymb,amsfonts,amsthm}
\begin{document}
Hello World
\end{document}
LATEX

tectonic /tmp/dummy.tex
rm -rf /tmp/*
EOF

# Non-root user
RUN adduser -D -u 1001 appuser && \
    chown -R appuser:appuser ${XDG_CACHE_HOME}

WORKDIR /home/appuser

# Copy binary only (NO extra files)
COPY --from=builder /build/atstex-lab /usr/local/bin/atstex-lab

USER appuser

EXPOSE 8080
ENV ADDR=:8080

ENTRYPOINT ["/usr/local/bin/atstex-lab"]
