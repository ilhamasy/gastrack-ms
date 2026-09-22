# gastrack-ms

GasTrack backend microservice — Golang REST API for the GasTrack motorcycle maintenance tracker.

## Technology Stack

- **Language:** Go 1.26+
- **Architecture:** Layered (Handler → Service → Repository → PostgreSQL)
- **Database:** PostgreSQL
- **Module:** `github.com/ilhamasy/gastrack-ms`

## Prerequisites

- Go 1.22 or later
- PostgreSQL 15 or later
- `make` (optional, for convenience commands)

## Project Structure

```
cmd/
  api/              # Application entry point
internal/
  config/           # Configuration loading
  handler/          # HTTP handlers (thin, request/response only)
  middleware/       # Auth, logging, rate-limit middleware
  service/          # Business logic and domain rules
  repository/       # Database access (SQL queries, transactions)
  model/            # Data models and DTOs
  validator/        # Input validation
  notification/     # Notification scheduling and delivery
migrations/         # Versioned PostgreSQL migrations
docs/               # API documentation
```

## Local Setup

1. **Clone the repo**
   ```bash
   git clone https://github.com/ilhamasy/gastrack-ms.git
   cd gastrack-ms
   ```

2. **Copy environment config**
   ```bash
   cp .env.example .env
   # Edit .env with your local database credentials
   ```

3. **Install dependencies**
   ```bash
   go mod tidy
   ```

4. **Run the server**
   ```bash
   go run ./cmd/api
   ```

5. **Health check**
   ```bash
   curl http://localhost:8080/health
   ```

## Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `PORT` | HTTP server port | `8080` |
| `DATABASE_URL` | PostgreSQL connection string | `postgres://user:pass@localhost:5432/gastrack` |
| `JWT_SECRET` | JWT signing secret | _(generate securely)_ |
| `ENV` | Environment name | `development` |

> ⚠️ Never commit real credentials. Use `.env` locally (excluded by `.gitignore`).

## Architecture Principles

- HTTP handlers are thin — request parsing and response serialization only
- Business rules live in the service layer
- Database access is isolated in the repository layer
- Maintenance calculations are authoritative on the backend (not duplicated in clients)
- All protected resources enforce server-side ownership checks

## API Base URL

```
/api
```

Response format:
```json
{ "data": {}, "meta": {} }
```

Error format:
```json
{ "error": { "code": "VEHICLE_NOT_FOUND", "message": "Vehicle not found" } }
```

## Related Repositories

- **Mobile App:** [gastrack-app](https://github.com/ilhamasy/gastrack-app) — Flutter (Android + iOS)
