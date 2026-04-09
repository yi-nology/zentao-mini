# ========================
# Stage 1: Build Frontend
# ========================
FROM node:20-bookworm-slim AS frontend-builder

WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install --registry=https://registry.npmmirror.com
COPY frontend/ ./
RUN cp .env.docker .env.production 2>/dev/null || true
RUN npm run build

# ========================
# Stage 2: Build Backend (app mode with embedded frontend)
# ========================
FROM golang:1.24-bookworm AS backend-builder

WORKDIR /app
COPY go.mod go.sum ./
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download
COPY . .

# Copy frontend dist into app static dir (go:embed will pick it up)
COPY --from=frontend-builder /app/frontend/dist ./backend/cmd/app/static/

# Build the app binary (single binary serves both frontend + API)
RUN cd backend && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /app/zentao-mini ./cmd/app/main.go

# ========================
# Stage 3: Production
# ========================
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates tzdata curl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=backend-builder /app/zentao-mini /app/zentao-mini

RUN mkdir -p /app/data

ENV TZ=Asia/Shanghai
ENV GIN_MODE=release
ENV PORT=12345

EXPOSE 12345

HEALTHCHECK --interval=30s --timeout=5s --retries=3 --start-period=10s \
    CMD curl -f http://localhost:12345/api/health || exit 1

CMD ["/app/zentao-mini"]
