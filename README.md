## Forge Template: Svelte + Go (Bun + latest Go)

This repository is a minimal example of how to structure a Forge-compatible application:

- Svelte frontend (built with Bun + Vite)
- Go backend serving the built SPA, plus a simple JSON API
- Dockerfile and docker-compose.yml compatible with Forge connectors and Traefik

### What is Forge?

Forge is a small platform to deploy GitHub repos via Docker Compose, fronted by Traefik and Cloudflare Tunnels/Access. In this template’s context:

- You connect a repo+branch in Forge
- Add one or more “connectors” mapping subdomains to a service:port in your Compose
- Forge autodeploys branch pushes to a dev domain (e.g., `ddls.dev`) and lets you manually publish to prod (e.g., `ddls.app`)
- Cloudflare Access protects your apps; the user’s email is passed to the backend via `Cf-Access-Authenticated-User-Email`

### What’s in this template

- `web/` – Svelte app using Vite
  - Fetches `/api/info` and renders the Cloudflare Access email and environment variables
  - Dark-only UI following `design_guidelines.md`
- `backend/` – Go server
  - Embeds the built `web/` into the binary
  - Serves `/api/info` with `{ message, email, env }`
  - Listens on `PORT` (default 8080)
- `Dockerfile` – Multi-stage build: Bun (Svelte) → Go (embed) → alpine runtime
- `docker-compose.yml` – Defines a single service `web`; Forge will inject Traefik labels via an override
- `design_guidelines.md` – Dark-only tokens and usage patterns

### Requirements

- For local builds:
  - Bun (or Docker to build with the multi-stage file)
  - Go (latest)
  - Node tooling is not required if you rely on Bun or Docker
- For Forge deployments:
  - Root-level `docker-compose.yml`
  - A named service (here: `web`) listening on an internal port (here: 8080)
  - Forge connectors configured to map subdomains → service:port
  - Cloudflare Tunnels/Access set up by Forge

### Compose + connectors

This template intentionally leaves Traefik labels off in `docker-compose.yml`. Forge generates a `forge.override.yml` per app/connector to set:

- Traefik router rules for each subdomain
- Traefik service load balancer port for the internal container port
- Attachment to the shared `traefik_proxy` network

Your compose service should:

- Be reachable on the internal port you want exposed (8080 here)
- Join `traefik_proxy` (created as an external network by Forge stack)

### Cloudflare Access email

Your Go backend receives the user email in header `Cf-Access-Authenticated-User-Email`. The sample `/api/info` endpoint returns it so the Svelte app can display it.

### Env vars (dev/prod)

Forge supports env-specific variables. At deploy time the platform writes `.env.dev` or `.env.prod` and mounts them through an override file. This app simply prints all environment variables in `/api/info` for demonstration.

### Local development

Option 1: Docker

```
docker build -t forge-template .
docker run --rm -p 8080:8080 forge-template
```

Option 2: Bun + Go (outside Docker)

```
cd web && bun install && bun run build
cd ../backend && go build ./cmd/app && ./app
```

Visit http://localhost:8080

### Notes

- Keep Compose at repo root for Forge discovery.
- Avoid renaming the `web` service unless you also update your Forge connectors.

