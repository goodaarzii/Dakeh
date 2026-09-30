# Dakeh

An open-source online shop, built from scratch as a modular monolith in Go.
Dakeh is designed to be a clear, well-documented reference for building a real-world e-commerce backend: clean architecture, a carefully designed relational schema, and production-minded practices.

> **Status:** early development. See the [Roadmap](#roadmap).

## Features

**MVP**
- User registration and login with JWT (short-lived access token + revocable refresh token)
- Roles (customer, admin) and role-protected endpoints
- Addresses with coordinates (lat/lng)
- Categories (tree), tags, products, and product variants (SKU)
- Product images
- Cart stored in Redis
- Orders and order items with price/name snapshots
- Stock management with transactional decrement (no overselling)
- Favorites / wishlist
- Ratings and reviews
- Payment abstraction (provider is pluggable)

**Later**
- Search (Postgres full-text first)
- Admin UI
- Returns and refunds

## Tech Stack

| Area | Choice |
|------|--------|
| Language | Go |
| HTTP framework | [Gin](https://github.com/gin-gonic/gin) |
| Database | PostgreSQL |
| DB access | [pgx](https://github.com/jackc/pgx) + [sqlc](https://sqlc.dev) (type-safe generated code from plain SQL) |
| Migrations | golang-migrate (or goose) |
| Cache / cart / rate limiting | Redis |
| Auth | JWT (golang-jwt) + argon2id/bcrypt password hashing |
| Object storage | S3-compatible (MinIO locally, CDN/S3 in production) |
| API docs | OpenAPI / Swagger |
| Validation | go-playground/validator |
| Logging | `log/slog` (structured, with request ID) |
| Testing | testify + testcontainers-go (integration tests on real Postgres/Redis) |
| CI | GitHub Actions (golangci-lint, tests, Docker build) |
| Containers | Docker (multi-stage) + Docker Compose |

## Architecture

Dakeh follows clean architecture with manual dependency injection:

```
cmd/            entrypoints (main.go wires everything together)
internal/
  domain/       entities and business rules
  usecase/      application services (define the interfaces they need)
  repository/   Postgres / Redis implementations
  transport/    HTTP handlers, middleware, DTOs
  config/       environment-based configuration
db/
  migrations/   versioned SQL migrations
  queries/      SQL used by sqlc
docs/           schema diagram, API spec, design decisions
```

Design decisions:
- Money is stored as integers in minor units, never floats.
- Order items snapshot the product name and price at purchase time.
- Order status follows an explicit state machine.
- Stock is decremented inside a database transaction.
- Checkout uses an idempotency key to prevent duplicate orders.
- Products are soft-deleted so historical orders stay valid.

## Getting Started

```bash
git clone https://github.com/<your-username>/dakeh.git
cd dakeh
cp .env.example .env
docker compose up -d
make migrate-up
make run
```

The API will be available at `http://localhost:8080`, with Swagger UI at `/swagger`.

## Development

```bash
make test        # unit tests
make test-int    # integration tests (needs Docker)
make lint        # golangci-lint
make sqlc        # regenerate DB code
```

## Roadmap

- [ ] Database schema and diagram
- [ ] Auth (register, login, refresh, roles)
- [ ] Catalog (categories, products, variants, images)
- [ ] Cart (Redis)
- [ ] Orders, stock, and checkout
- [ ] Payments
- [ ] Favorites and ratings
- [ ] Frontend
- [ ] Admin UI

## Contributing

Contributions are welcome. Please open an issue to discuss larger changes first.

## License

MIT
