# ── Stage 1: Build React frontend ─────────────────────────────────────────────
FROM node:20-alpine AS web-builder
WORKDIR /web
COPY web/package*.json ./
RUN npm install --no-audit --no-fund
COPY web/ .
RUN npm run build

# ── Stage 2: Build Go binary ───────────────────────────────────────────────────
FROM golang:1.23-alpine AS go-builder
WORKDIR /app

# Download dependencies first (cached layer)
COPY go.mod ./
RUN go mod download 2>/dev/null || true

COPY . .

# Copy built frontend into the embed path
COPY --from=web-builder /web/dist ./web/dist

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GONOSUMDB=*
ENV GOFLAGS=-mod=mod

RUN go build -o feedfarmer .

# ── Stage 3: Minimal runtime image ────────────────────────────────────────────
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=go-builder /app/feedfarmer ./feedfarmer

EXPOSE 8080
VOLUME ["/data"]
ENV DB_PATH=/data/feedfarmer.db

ENTRYPOINT ["./feedfarmer"]
