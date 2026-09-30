# Dakeh API Plan

Endpoint list and build order. This is the contract between backend and frontend; it will become `openapi.yaml`.

## Conventions

- Base path: `/api/v1`, JSON only, `snake_case` fields
- Auth: `Authorization: Bearer <access_token>`
- Money: integer cents in USD (`1000` = $10.00)
- IDs: UUID
- Pagination: `?limit=20&cursor=<opaque>` returns `{ "items": [...], "next_cursor": "..." | null }`
- Errors: `{ "error": { "code": "out_of_stock", "message": "...", "details": [] } }` with correct HTTP status codes
- `POST /orders` requires an `Idempotency-Key` header
- Access legend: **P** public, **U** logged-in user, **A** admin

## Endpoints

### Auth and users
| Method | Path | Access | Purpose |
|---|---|---|---|
| POST | `/auth/register` | P | Create account |
| POST | `/auth/login` | P | Returns access + refresh token |
| POST | `/auth/refresh` | P | Rotate refresh token, new access token |
| POST | `/auth/logout` | U | Revoke refresh token |
| GET | `/users/me` | U | Current profile |
| PATCH | `/users/me` | U | Update name, phone |
| PUT | `/users/me/password` | U | Change password |

### Addresses
| Method | Path | Access | Purpose |
|---|---|---|---|
| GET | `/users/me/addresses` | U | List my addresses |
| POST | `/users/me/addresses` | U | Add address (with lat/lng) |
| PATCH | `/users/me/addresses/{id}` | U | Edit, or set as default |
| DELETE | `/users/me/addresses/{id}` | U | Soft delete |

### Catalog (read)
| Method | Path | Access | Purpose |
|---|---|---|---|
| GET | `/categories` | P | Category tree |
| GET | `/tags` | P | All tags |
| GET | `/products` | P | List with filters: `category`, `tag`, `q`, `min_price`, `max_price`, `sort`, cursor |
| GET | `/products/{slug}` | P | Detail with variants, images, rating summary |

### Catalog (admin)
| Method | Path | Access | Purpose |
|---|---|---|---|
| POST / PATCH / DELETE | `/admin/categories[/{id}]` | A | Manage categories |
| POST / PATCH / DELETE | `/admin/tags[/{id}]` | A | Manage tags |
| POST / PATCH / DELETE | `/admin/products[/{id}]` | A | Manage products (including tags) |
| POST / PATCH / DELETE | `/admin/products/{id}/variants[/{variant_id}]` | A | Manage variants |
| POST | `/admin/products/{id}/images` | A | Upload image (to MinIO/S3) |
| DELETE | `/admin/images/{id}` | A | Remove image |
| POST | `/admin/variants/{id}/stock` | A | Adjust stock (creates a stock movement) |
| GET | `/admin/variants/{id}/stock-movements` | A | Stock history |

### Cart (Redis, logged-in only)
| Method | Path | Access | Purpose |
|---|---|---|---|
| GET | `/cart` | U | Cart with current prices and stock status |
| PUT | `/cart/items/{variant_id}` | U | Set quantity |
| DELETE | `/cart/items/{variant_id}` | U | Remove item |
| DELETE | `/cart` | U | Clear cart |
| POST | `/cart/merge` | U | Send guest cart items (from browser) after login |

### Orders and checkout
| Method | Path | Access | Purpose |
|---|---|---|---|
| POST | `/orders` | U | Checkout: body `{ address_id }`, header `Idempotency-Key`. Creates order from cart |
| GET | `/orders` | U | My orders |
| GET | `/orders/{id}` | U | Order detail with items and status history |
| POST | `/orders/{id}/cancel` | U | Cancel while still `pending` |
| GET | `/admin/orders` | A | All orders with filters |
| GET | `/admin/orders/{id}` | A | Any order |
| PATCH | `/admin/orders/{id}/status` | A | Move status (validated by state machine) |

### Payments
| Method | Path | Access | Purpose |
|---|---|---|---|
| POST | `/orders/{id}/pay` | U | Start payment, returns provider info |
| POST | `/payments/webhook` | P (signature verified) | Provider callback, marks order paid |

### Favorites and reviews
| Method | Path | Access | Purpose |
|---|---|---|---|
| GET | `/favorites` | U | My favorite products |
| PUT | `/favorites/{product_id}` | U | Add (idempotent) |
| DELETE | `/favorites/{product_id}` | U | Remove |
| GET | `/products/{id}/reviews` | P | Reviews for a product |
| POST | `/products/{id}/reviews` | U | Add review (one per user per product) |
| PATCH / DELETE | `/reviews/{id}` | U | Edit or delete my review |

### Operations
| Method | Path | Access | Purpose |
|---|---|---|---|
| GET | `/healthz` | P | Liveness |
| GET | `/readyz` | P | Readiness (Postgres + Redis reachable) |
| GET | `/swagger` | P | API docs |

## Checkout flow

1. Frontend calls `POST /orders` with `address_id` and an `Idempotency-Key`.
2. Backend reads the cart from Redis and, in **one DB transaction**: locks inventory rows, checks stock, creates `orders` + `order_items` (snapshotting names and prices), adds the address snapshot and the fixed shipping fee, reserves stock, writes `stock_movements` and the first `order_status_history` row.
3. Backend clears the cart and returns the order (`pending`).
4. `POST /orders/{id}/pay` starts a payment. The webhook confirms it and moves the order to `paid`.
5. A background job cancels `pending` orders older than about 30 minutes and releases the reserved stock.

## Build order

| Phase | What | Why here |
|---|---|---|
| 0 | Foundation: repo layout, docker-compose (Postgres, Redis, MinIO), config, migrations, sqlc setup, slog, error format, health endpoints, CI | Everything else sits on this |
| 1 | Auth and users, role middleware | Almost every endpoint needs it |
| 2 | Addresses | Small, and checkout depends on it |
| 3 | Catalog: categories, tags, products, variants, images, stock, plus a **seed script** with sample data | Frontend can start here with real data |
| 4 | Cart (Redis) and `/cart/merge` | Needs products and variants |
| 5 | Orders, checkout, stock reservation, status history, expiry job | The hardest part: transactions, concurrency, idempotency. Write tests as you go |
| 6 | Payments with a **mock provider** first, real provider later | Keeps checkout finishable before choosing a gateway |
| 7 | Favorites and reviews (with rating aggregates) | Independent, easy wins |
| 8 | Polish: rate limiting, search, integration tests, docs, screenshots | Makes the repo presentable |

**Parallel work:** once the OpenAPI file covers phases 1 to 4, the frontend can be built against a mock server while the backend catches up.
