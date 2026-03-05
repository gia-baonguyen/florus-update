## `user_events` schema (analytics & recommendation)

This table is the **stable event log** for all analytics / recommendation
pipelines. Future changes should avoid renaming/dropping columns – prefer
adding new columns or using the `METADATA` JSON field.

### Table structure (Oracle)

- `ID` (INTEGER, identity, PK)
- `USER_ID` (INTEGER, nullable) – logged-in user id, or NULL for anonymous
- `SESSION_ID` (VARCHAR2(100)) – stable session key from FE (`X-Session-ID`)
- `EVENT_TYPE` (VARCHAR2(50), NOT NULL) – high‑level type, see below
- `PRODUCT_ID` (INTEGER, nullable)
- `CATEGORY_ID` (INTEGER, nullable)
- `ORDER_ID` (INTEGER, nullable)
- `SEARCH_QUERY` (VARCHAR2(255), nullable)
- `QUANTITY` (INTEGER, nullable)
- `PRICE` (NUMBER(15,2), nullable) – unit price at time of event
- `RATING` (INTEGER, nullable)
- `METADATA` (CLOB, nullable) – JSON blob for extra data
- `USER_AGENT` (VARCHAR2(500), nullable)
- `IP_ADDRESS` (VARCHAR2(45), nullable)
- `REFERRER` (VARCHAR2(500), nullable)
- `CREATED_AT` (TIMESTAMP WITH TIME ZONE)

### Event types currently in use

All event types are stored as lowercase strings in `EVENT_TYPE` and mapped in
`internal/models/user_event.go` (`EventType`):

- `product_view` – product detail viewed
- `product_click` – product clicked from a list/card
- `add_to_cart` – item added to cart
- `remove_from_cart` – item removed from cart
- `add_to_wishlist` – product added to wishlist
- `purchase` – order created (one event per order item)
- `search` – search submitted (header, products page, category page)
- `category_view` – category page viewed
- `review` – review created
- `rating` – rating given (even if no text review)

### Guaranteed invariants for pipelines

- Table name and column names are **stable**; do not rename/drop.
- `EVENT_TYPE` is the primary classifier for behavior.
- When present:
  - `PRODUCT_ID` points to `products.ID`
  - `CATEGORY_ID` points to `categories.ID`
  - `ORDER_ID` points to `orders.ID`
- Missing values are represented as `NULL`, **not** sentinel values.
- Any extra per‑event attributes should go into `METADATA` as JSON, e.g.:
  - `{ "source": "banner_home", "position": 3 }`
  - `{ "experiment": "rec_v2", "variant": "B" }`

### Typical queries (for data pipelines)

- User sequence:
  - `SELECT * FROM user_events WHERE user_id = ? ORDER BY created_at`
- Product engagement:
  - `SELECT * FROM user_events WHERE product_id = ?`
- Co‑view / co‑purchase:
  - Use `EVENT_TYPE IN ('product_view','purchase')` grouped by `USER_ID`

