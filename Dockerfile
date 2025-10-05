## Multi-stage build: Svelte (bun) → Go (latest) → runtime

# 1) Build Svelte app with bun
FROM oven/bun:1.2-alpine AS web
WORKDIR /app
COPY web/package.json ./
RUN bun install
COPY web/ ./
RUN bun run build && \
    mkdir -p /out && \
    if [ -d dist ]; then cp -r dist/* /out/; else cp -r build/* /out/; fi

# 2) Build Go server with embedded static
FROM golang:1.25.1-alpine AS backend
WORKDIR /src
COPY backend/ ./backend/
# Copy compiled static assets, supporting Vite (dist) or SvelteKit (build)
COPY --from=web /out ./backend/static
WORKDIR /src/backend
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/app ./cmd/app

# 3) Runtime
FROM alpine:3.20
WORKDIR /app
COPY --from=backend /out/app ./app
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/app/app"]


