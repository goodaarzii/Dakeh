# Dakeh front end

The web shop for the Dakeh API. React, TypeScript, Vite and Tailwind CSS.
Version one is for Chrome on a normal desktop screen, in English, and talks to the Go API in this repo.

## Run it

You need Node.js 20 or newer.

```bash
cd frontend
cp .env.example .env        # on Windows: copy .env.example .env
npm install
npm run typecheck           # optional: checks the TypeScript
npm run dev                 # opens on http://localhost:5173
```

The API must be running first (`docker compose up -d`, then `go run ./cmd/api`).
`.env` has one setting: `VITE_API_URL` (default `http://localhost:8080/api/v1`).

The dev server uses port 5173 on purpose. The API allows that origin through `CORS_ORIGINS`, and `middleware.CORS(...)` must be
registered first in `routes.go`. If the page loads but the inspector shows `ERR` for every call, that is CORS.

## Your own images

Put your files in `frontend/public/images/` with these exact names. Nothing else needs to change.

| File | Used for |
|---|---|
| `logo.svg` | The logo in the header. A PNG also works if you change `LOGO_IMAGE` in `src/config.ts` |
| `placeholder-product.png` | Shown for any product without a picture. A 4:5 portrait works best (for example 800 x 1000) |
| `../favicon.svg` (in `frontend/public/`) | The tab icon |

The logo, favicon and placeholder in the project are plain stand-ins.

## What is in version one

- Shop: product list with visible filters (categories, tags, search, sort) and previous/next paging. The address bar holds the filters.
- Product page: pictures, variant options, price, stock, quantity, add to cart, tags, reviews summary.
- Cart: a guest cart in the browser, and the real cart on the server after login. The guest cart is merged on login.
- Accounts: register, log in, log out, log out everywhere, profile.
- Addresses: add, edit and delete, with a map to drop the pin (OpenStreetMap, no key needed).
- Under construction dialogs: checkout and payment, favorites, and writing a review.
- API inspector: the button at the bottom right shows every request and response. "Copy as curl" uses `{{baseUrl}}` and `{{token}}`.

Not in this version: orders, payment, favorites, reviews, password reset, profile editing, the admin panel and the home page banners.

## How it is organized

```
src/
  config.ts            API address, image paths, limits
  lib/
    api.ts             fetch wrapper: token header, refresh on 401, records every call for the inspector
    auth.tsx           who is logged in
    cart.tsx           one cart, two sources (guest in the browser, or the server)
    inspector.ts       the store behind the API inspector
    types.ts           the shapes of the API's JSON
    variants.ts        turns variant attributes into option buttons
    messages.ts        API error codes -> friendly sentences
  components/          header and footer, product card, filters, dialogs, map, forms...
  pages/               one file per screen
```

## Good to know

- Prices come from the API in cents and are shown as dollars.
- Tokens are kept in localStorage. That is simple and common for a single-page app. A script injected into the page could read them.
- A guest cart keeps a copy of the product name and price so it can be shown. Prices and stock are checked again after login.
- Fonts (Bricolage Grotesque and Instrument Sans) load from Google Fonts, so the first run needs internet.
- The address API accepts "default address" only when an address is created, so the form hides that box when you already have one.
