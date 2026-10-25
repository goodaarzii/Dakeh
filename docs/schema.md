# Dakeh Database Schema

This page keeps the schema diagram as Mermaid so GitHub can render it directly. The first diagram describes tables present in the current migrations (`db/migrations/00001` through `00010`). The cart is stored in Redis and therefore does not appear in the relational diagram.

All product/variant prices are integer minor units. The frontend currently presents them as USD cents.

## Current database schema

```mermaid
erDiagram
    USERS ||--o{ ADDRESSES : owns
    USERS ||--o{ REFRESH_TOKENS : has
    USERS ||--o{ FAVORITES : saves
    USERS ||--o{ REVIEWS : writes

    CATEGORIES o|--o{ CATEGORIES : parent_of
    CATEGORIES ||--o{ PRODUCTS : contains
    PRODUCTS ||--o{ PRODUCT_VARIANTS : offers
    PRODUCTS ||--o{ PRODUCT_IMAGES : has
    PRODUCT_VARIANTS o|--o{ PRODUCT_IMAGES : optionally_shown_in
    PRODUCTS ||--o{ PRODUCT_TAGS : tagged_with
    TAGS ||--o{ PRODUCT_TAGS : assigned_to
    PRODUCT_VARIANTS ||--o| INVENTORY : tracks
    PRODUCTS ||--o{ FAVORITES : saved_by
    PRODUCTS ||--o{ REVIEWS : receives

    USERS {
        uuid id PK
        string email UK
        string phone UK
        string password_hash
        string full_name
        string role "customer | admin"
        boolean is_active
        timestamp last_login_at
        timestamp created_at
        timestamp updated_at
    }

    REFRESH_TOKENS {
        uuid id PK
        uuid user_id FK
        string token_hash UK
        string user_agent
        timestamp expires_at
        timestamp revoked_at
        timestamp created_at
    }

    ADDRESSES {
        uuid id PK
        uuid user_id FK
        string label
        string recipient_name
        string phone
        string country
        string province
        string city
        string street
        string postal_code
        float lat
        float lng
        boolean is_default
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    CATEGORIES {
        uuid id PK
        uuid parent_id FK
        string name
        string slug UK
        string description
        string image_url
        int sort_order
        boolean is_active
        timestamp created_at
        timestamp updated_at
    }

    PRODUCTS {
        uuid id PK
        uuid category_id FK
        string name
        string slug "unique while not deleted"
        string description
        string brand
        string status "draft | active | archived"
        float avg_rating
        int rating_count
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    TAGS {
        uuid id PK
        string name UK
        string slug UK
        timestamp created_at
        timestamp updated_at
    }

    PRODUCT_TAGS {
        uuid product_id PK, FK
        uuid tag_id PK, FK
    }

    PRODUCT_VARIANTS {
        uuid id PK
        uuid product_id FK
        string sku "unique while not deleted"
        string name
        bigint price "integer cents"
        bigint compare_at_price "0 means none"
        jsonb attributes
        int weight_grams
        boolean is_active
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    INVENTORY {
        uuid variant_id PK, FK
        int quantity_on_hand
        int quantity_reserved
        int low_stock_threshold
        timestamp updated_at
    }

    PRODUCT_IMAGES {
        uuid id PK
        uuid product_id FK
        uuid variant_id FK "optional"
        string storage_key UK
        string url
        string alt_text
        int sort_order
        boolean is_primary
        timestamp created_at
    }

    FAVORITES {
        uuid user_id PK, FK
        uuid product_id PK, FK
        timestamp created_at
    }

    REVIEWS {
        uuid id PK
        uuid product_id FK
        uuid user_id FK
        smallint rating "1 to 5"
        string title
        string body
        timestamp created_at
        timestamp updated_at
    }
```

### Notes on the current schema

- `favorites` has a composite primary key on `(user_id, product_id)`, preventing duplicate favorites.
- `reviews` has a unique constraint on `(product_id, user_id)`: one review per customer per product. A trigger maintains the product's average rating and review count.
- A product or variant may be soft-deleted; its `deleted_at` column preserves the row when historical references may matter.
- `inventory.quantity_reserved` cannot exceed `quantity_on_hand`. Available stock is `quantity_on_hand - quantity_reserved`.
- Product image bytes are not stored in PostgreSQL. The database stores each object's storage key and public URL; MinIO/S3 stores the object itself.
- The current migration set does **not** create `orders`, `order_items`, `payments`, or checkout stock-movement tables.

## Checkout/order extension (planned separately)

The following is a design view for the order-submission patch, not a description of migrations `00001`–`00010`. It becomes part of the active schema only after the matching order migration is merged and applied.

```mermaid
erDiagram
    USERS ||--o{ ORDERS : submits
    ADDRESSES o|--o{ ORDERS : source_address
    ORDERS ||--o{ ORDER_ITEMS : contains
    ORDERS ||--o{ ORDER_STATUS_HISTORY : records

    ORDERS {
        uuid id PK
        uuid user_id FK
        uuid address_id FK "nullable after address removal"
        string status "submitted | cancelled"
        string idempotency_key "unique per user"
        string currency "USD"
        bigint subtotal
        bigint shipping_fee
        bigint total
        string shipping_recipient_name "snapshot"
        string shipping_phone "snapshot"
        string shipping_country "snapshot"
        string shipping_province "snapshot"
        string shipping_city "snapshot"
        string shipping_street "snapshot"
        string shipping_postal_code "snapshot"
        timestamp created_at
        timestamp updated_at
        timestamp cancelled_at
    }

    ORDER_ITEMS {
        uuid id PK
        uuid order_id FK
        uuid product_id "snapshot reference"
        uuid variant_id "snapshot reference"
        string product_name "snapshot"
        string product_slug "snapshot"
        string variant_name "snapshot"
        string sku "snapshot"
        jsonb attributes "snapshot"
        string image_url "snapshot"
        bigint unit_price "snapshot"
        int quantity
        bigint line_total
        timestamp created_at
    }

    ORDER_STATUS_HISTORY {
        uuid id PK
        uuid order_id FK
        string from_status
        string to_status
        string note
        timestamp created_at
    }

    STOCK_MOVEMENTS {
        uuid id PK
        uuid variant_id "snapshot reference"
        int delta_on_hand
        int delta_reserved
        string reason
        string reference_type
        uuid reference_id
        timestamp created_at
    }
```

In this first checkout phase, an order is submitted without charging a payment method. Product/variant identifiers and display details are snapshots in `order_items` so later catalog edits/deletions do not rewrite the customer's order history. `stock_movements.reference_type` and `reference_id` can refer to the order, but this is a polymorphic reference rather than a PostgreSQL foreign key, so the diagram does not draw it as a relational edge. Payment tables/provider callbacks should be added in a later phase rather than implied by this diagram.
