# Dakeh frontend

The web storefront for the Dakeh API, built with React, TypeScript, Vite, Tailwind CSS, TanStack Query, React Router, and React Leaflet. Version one is designed for an English-language desktop browser.

## Run locally

Prerequisites: Node.js 20+ and npm. Start the backend and its Docker services first; see the root [README](../README.md).

From PowerShell:

```powershell
cd frontend
Copy-Item .env.example .env
npm install
npm run dev
```

From macOS/Linux, replace `Copy-Item .env.example .env` with `cp .env.example .env`.

Open `http://localhost:5173`. The default API URL is `http://localhost:8080/api/v1`; set `VITE_API_URL` in `frontend/.env` if your API uses a different address.

## Frontend features

- **Shop:** browse products and use category/tag filters, search, sorting, and pagination.
- **Product page:** inspect images, variants, prices, availability, and product reviews; add selected variants to the cart or save the product as a favorite.
- **Cart:** guest cart in browser storage; authenticated cart fetched from Redis-backed API; merge guest items after login.
- **Authentication:** registration, login, refresh handling, logout, and logout everywhere.
- **Account:** profile details, address management with a map picker, Favorites, and My Reviews.
- **Favorites:** list saved products, open product pages, and remove saved items.
- **Reviews:** public product-level review list; eligible users can create, edit, and delete their own review. The account tab lists the signed-in user's reviews with pagination.
- **API inspector:** use the inspector at the bottom-right of the application to inspect requests/responses and copy a cURL example.

**Checkout status:** persisted orders and payment are not part of the original source snapshot. The cart's checkout UI is a placeholder until the order-submission implementation is merged. Submitting a cart does not charge a payment method in this version.

## Images and API URL

Set `VITE_API_URL` to the full API base URL including `/api/v1`:

```env
VITE_API_URL=http://localhost:8080/api/v1
```

Local image assets belong in `frontend/public/images/`. The app uses the configured logo and a placeholder when a product has no image. The API's MinIO/S3 configuration is managed by the backend `.env`, not `frontend/.env`.

## Validation

```bash
npm run typecheck
npm run build
```

`npm run build` runs the TypeScript check before Vite builds the production assets.

## Code organization

```text
src/
  config.ts             API URL, shop name, image paths, frontend limits
  lib/
    api.ts              HTTP wrapper, auth token handling, API inspector integration
    auth.tsx            current user and authentication actions
    cart.tsx            guest/server cart state and guest-cart merge
    types.ts            API response/request types
    variants.ts         product variant option helpers
    messages.ts         API errors mapped to user-friendly messages
  components/           layout, cards, filters, dialogs, address map, forms, reviews
  pages/                shop, product, cart, account, login, register, not-found screens
```

## Implementation notes

- API prices are integer cents and the frontend formats them as USD.
- The API client attaches the access token for authenticated requests and attempts refresh handling on `401`.
- The browser cart retains guest items locally; after login, the API revalidates the actual product/variant and stock state.
- Fonts load from Google Fonts, so the first page load needs network access for the external font files.
