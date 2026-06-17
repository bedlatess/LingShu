# LingShu

LingShu is a private AI API aggregation gateway. It issues platform API keys,
forwards OpenAI-compatible requests to upstream providers, and charges users by
`base_cost * rate_multiplier`.

## Scope

- Private operation with two roles: `admin` and `user`.
- OpenAI-compatible gateway endpoints: `/v1/models` and `/v1/chat/completions`.
- Billing is always `charge = base_cost * rate_multiplier`; gateway logs and
  balance ledger both store `base_cost`, `rate_multiplier`, and `charge`.
- Recharge is manual: admin grants or redeem codes. There is no payment gateway,
  distribution tree, ticket system, or multi-level permission model.

## Production Docker Compose

1. Copy `.env.example` to `.env` and change all secrets:

```powershell
Copy-Item .env.example .env
notepad .env
```

Minimum required changes:

- `APP_ENV=production`
- `JWT_SECRET`
- `KEY_ENCRYPTION_SECRET`
- `POSTGRES_PASSWORD`
- `ADMIN_USER`
- `ADMIN_PASS`

2. Start the stack:

```bash
docker compose up --build -d
```

The production compose file starts the full stack:

- `web` — Nginx serving both SPAs and reverse-proxying the backend:
  - **User console** (public) on port `80`
  - **Admin console** (internal — restrict to LAN/VPN or an IP allowlist) on port `8081`
- `app` — Go gateway, internal only (exposed to other containers on `8080`, not published)
- `postgres` with the `pgdata` persistent volume
- `redis` with AOF enabled and the `redisdata` persistent volume

The frontend talks to the backend same-origin: Nginx proxies `/api`, `/v1`
(streaming SSE, buffering off), and `/healthz` to `app:8080`. The backend runs
migrations and seeds the initial administrator on startup.

Smoke check after `up`:

```bash
curl -fsS http://localhost/healthz        # via web → app
curl -fsS http://localhost:8081/healthz   # admin host → app
```

3. Put Caddy or Nginx with TLS in front for HTTPS in production. If you use
   `X-Session-Id` for upstream stickiness later, allow underscore headers in your
   reverse proxy.

## Local Development

Expose Postgres and Redis locally with the dev override:

```powershell
docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build
```

Or run PostgreSQL and Redis yourself, then:

```powershell
make migrate
make seed
```

On Windows without `make`, use:

```powershell
.\scripts\dev.ps1 migrate
.\scripts\dev.ps1 seed
.\scripts\dev.ps1 sqlc
.\scripts\dev.ps1 test
```

Start the backend outside Docker:

```powershell
cd backend
go run ./cmd/server
```

Start frontend apps:

```powershell
npm --prefix frontend/user install
npm --prefix frontend/admin install
.\scripts\dev.ps1 user-dev    # http://localhost:5173
.\scripts\dev.ps1 admin-dev   # http://localhost:5174
```

Build and test:

```powershell
cd backend
go test ./...

cd ..\frontend
npm run build
```

## Operational Checks

- Single API key over RPM or concurrency limit returns `429`.
- Insufficient balance returns `402`.
- Upstream `401/403/429/5xx` is retried on the next healthy bound channel; repeated
  failures increase `fail_count` and mark the channel `unhealthy` at its threshold.
- Balance changes are durable in Postgres; Redis only stores transient frozen
  reservations, rate windows, and concurrency counters.
