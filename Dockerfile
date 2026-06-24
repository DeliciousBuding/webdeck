# webdeck — Production Dockerfile
# Single container: Gateway + Chromium subprocess
# Multi-stage: Node (React frontend) → Go (gateway binary) → Debian (runtime)
#
# Build:
#   docker build -t webdeck .
#
# Run:
#   docker run -p 8090:8090 -v ./cloud_auth.json:/app/cloud_auth.json webdeck
#   docker run -p 8090:8090 -v ./cloud_auth.json:/app/cloud_auth.json webdeck --fps 15 --jpeg-quality 60

# ── Frontend stage ──
FROM node:22-alpine AS frontend
WORKDIR /fe
COPY cmd/gateway/frontend/package.json cmd/gateway/frontend/package-lock.json ./
RUN npm ci
COPY cmd/gateway/frontend/ .
RUN npm run build

# ── Go build stage ──
FROM golang:1.22-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=frontend /fe/dist cmd/gateway/frontend/dist
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/gateway ./cmd/gateway/

# ── Runtime stage ──
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    chromium \
    && rm -rf /var/lib/apt/lists/*

ENV CHROME_BIN=/usr/bin/chromium
ENV PUPPETEER_SKIP_CHROMIUM_DOWNLOAD=true

WORKDIR /app
COPY --from=build /out/gateway .

EXPOSE 8090

ENTRYPOINT ["./gateway"]
CMD ["--auth", "cloud_auth.json", "--port", "8090", "--fps", "30", "--jpeg-quality", "75"]
# For standalone/debug use with auto-navigation:
# CMD ["--auth", "cloud_auth.json", "--start-url", "https://sr.mihoyo.com/cloud/", "--port", "8090"]
