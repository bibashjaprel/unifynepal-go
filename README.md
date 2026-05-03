# Unify Nepal API

Production-ready Go backend API for **Unify Nepal**, a multi-tenant SaaS platform for Nepali shops to manage billing, inventory, customers, udharo, subscriptions, notifications, and admin operations.

## Tech Stack

- Go
- Gin
- GORM
- PostgreSQL / Neon DB
- JWT authentication
- Session tracking
- Multi-tenant shop isolation
- Docker-ready structure

## Current Status

Core SaaS backend foundation is complete.

- Auth + JWT + sessions
- Shop multi-tenancy
- Products
- Customers
- Billing transaction
- Payments
- Udharo ledger
- Inventory stock movement
- Dashboard stats
- Subscriptions + trial
- Upgrade request
- Notifications
- Audit logs
- Admin protection
- Request ID
- Request logging
- Recovery middleware
- CORS

## Project Structure

```txt
unifynepal-go/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── config/
│   ├── database/
│   ├── middleware/
│   ├── models/
│   ├── modules/
│   │   ├── admin/
│   │   ├── auth/
│   │   ├── audit/
│   │   ├── billing/
│   │   ├── customers/
│   │   ├── dashboard/
│   │   ├── inventory/
│   │   ├── notifications/
│   │   ├── products/
│   │   ├── subscriptions/
│   │   └── udharo/
│   ├── routes/
│   └── utils/
├── ENDPOINTS.MD
├── go.mod
├── go.sum
└── README.md
```

## Environment Variables

Create `.env` in the project root:

```env
APP_NAME=Unify Nepal API
APP_ENV=development
APP_PORT=8080

DATABASE_URL=postgresql://USER:PASSWORD@HOST.neon.tech/DBNAME?sslmode=require

JWT_SECRET=change-this-secret
JWT_EXPIRES_IN_HOURS=24

FRONTEND_URL=http://localhost:3000
```

## Run Locally

```bash
go mod tidy
go run ./cmd/api
```

API will run on:

```txt
http://localhost:8080
```

## Build

```bash
go build ./cmd/api
```

## Health Check

```bash
curl http://localhost:8080/health
```

## Authentication

### Signup

```bash
curl -X POST http://localhost:8080/api/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Bibas",
    "email": "bibas@example.com",
    "password": "password123",
    "shop_name": "Bibas Store"
  }'
```

Signup creates:

- user
- shop
- owner shop membership
- default 30-day Starter trial subscription
- user session
- JWT token

### Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "bibas@example.com",
    "password": "password123"
  }'
```

### Logout

```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer $TOKEN"
```

## Required Headers for Protected Shop APIs

Most shop APIs require:

```http
Authorization: Bearer <token>
X-Shop-ID: <shop_id>
Content-Type: application/json
```

Example:

```bash
export TOKEN='your_jwt_token'
export SHOP_ID='your_shop_id'
```

## Core API Modules

### Products

```http
GET    /api/v1/products
POST   /api/v1/products
GET    /api/v1/products/:id
PUT    /api/v1/products/:id
DELETE /api/v1/products/:id
```

Create product:

```bash
curl -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Shop-ID: $SHOP_ID" \
  -d '{
    "name": "Coca Cola 500ml",
    "sku": "COCA-500",
    "category": "Drinks",
    "unit": "piece",
    "selling_price": 80,
    "cost_price": 65,
    "stock_qty": 50,
    "min_stock_qty": 10
  }'
```

### Customers

```http
GET    /api/v1/customers
POST   /api/v1/customers
GET    /api/v1/customers/:id
PUT    /api/v1/customers/:id
DELETE /api/v1/customers/:id
```

Create customer:

```bash
curl -X POST http://localhost:8080/api/v1/customers \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Shop-ID: $SHOP_ID" \
  -d '{
    "name": "Ram Bahadur",
    "phone": "9800000000",
    "address": "Kathmandu"
  }'
```

### Billing

```http
GET  /api/v1/bills
POST /api/v1/bills
GET  /api/v1/bills/:id
```

Create bill:

```bash
curl -X POST http://localhost:8080/api/v1/bills \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Shop-ID: $SHOP_ID" \
  -d '{
    "customer_id": "CUSTOMER_ID",
    "discount": 0,
    "paid_amount": 50,
    "payment_method": "cash",
    "items": [
      {
        "product_id": "PRODUCT_ID",
        "quantity": 2,
        "unit_price": 80
      }
    ]
  }'
```

Billing transaction performs:

- creates bill
- creates bill items
- reduces product stock
- creates stock movement
- creates payment row if paid amount exists
- creates udharo ledger entry if due amount exists

### Udharo

```http
GET  /api/v1/udharo/summary
GET  /api/v1/udharo/customers
GET  /api/v1/customers/:id/udharo/ledger
POST /api/v1/customers/:id/udharo/payments
```

Collect udharo payment:

```bash
curl -X POST http://localhost:8080/api/v1/customers/CUSTOMER_ID/udharo/payments \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Shop-ID: $SHOP_ID" \
  -d '{
    "amount": 60,
    "method": "cash",
    "note": "Partial udharo collection"
  }'
```

Ledger formula:

```txt
Current Due = SUM(credit) - SUM(payment)
```

### Inventory

```http
GET  /api/v1/stock-movements
POST /api/v1/stock-movements
GET  /api/v1/inventory/low-stock
```

Manual stock in:

```bash
curl -X POST http://localhost:8080/api/v1/stock-movements \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Shop-ID: $SHOP_ID" \
  -d '{
    "product_id": "PRODUCT_ID",
    "type": "stock_in",
    "quantity": 10,
    "note": "Added 10 bottles from supplier"
  }'
```

Supported stock movement types:

```txt
stock_in
stock_out
adjustment
sale
opening
```

### Dashboard

```http
GET /api/v1/dashboard/stats
```

Returns:

```json
{
  "today_sales": 160,
  "month_sales": 160,
  "total_due": 0,
  "total_products": 1,
  "low_stock_count": 0,
  "total_customers": 1,
  "total_bills": 1
}
```

### Subscriptions

```http
GET  /api/v1/plans
GET  /api/v1/subscription
POST /api/v1/subscription/upgrade-request
```

Seeded plans:

| Plan | Price | Users | Products | Bills / Month |
|---|---:|---:|---:|---:|
| Starter | Rs 499 | 1 | 100 | 300 |
| Business | Rs 999 | 3 | 1000 | 3000 |
| Pro | Rs 1999 | 10 | 10000 | 10000 |

Upgrade request:

```bash
curl -X POST http://localhost:8080/api/v1/subscription/upgrade-request \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Shop-ID: $SHOP_ID" \
  -d '{
    "plan_id": "PLAN_ID",
    "note": "I want to upgrade to Business plan"
  }'
```

This creates:

- notification
- audit log

### Notifications

```http
GET  /api/v1/notifications
POST /api/v1/notifications/:id/read
POST /api/v1/notifications/read-all
```

List notifications:

```bash
curl -X GET http://localhost:8080/api/v1/notifications \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Shop-ID: $SHOP_ID"
```

### Audit Logs

```http
GET /api/v1/audit/logs
```

List audit logs:

```bash
curl -X GET http://localhost:8080/api/v1/audit/logs \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Shop-ID: $SHOP_ID"
```

### Admin

```http
GET   /api/v1/admin/shops
GET   /api/v1/admin/users
PATCH /api/v1/admin/shops/:id/status
```

Admin APIs require user role:

```txt
platform_admin
```

To test locally:

```sql
UPDATE users
SET role = 'platform_admin'
WHERE email = 'your_email@example.com';
```

Then login again and use the new token.

## Database Tables

Main tables:

```txt
users
user_sessions
shops
shop_members
products
customers
bills
bill_items
payments
customer_ledger_entries
stock_movements
subscription_plans
shop_subscriptions
notifications
audit_logs
```

## Multi-Tenant Rule

Every business table uses:

```txt
shop_id
```

Backend validates:

```txt
user_id + shop_id exists in shop_members
```

Never trust `shop_id` from request body.

Use:

```http
X-Shop-ID: <shop_id>
```

## Request Tracking

Every request receives:

```http
X-Request-ID
```

Example:

```bash
curl -i http://localhost:8080/api/v1/dashboard/stats \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Shop-ID: $SHOP_ID"
```

Response header:

```txt
X-Request-Id: uuid
```

## Middleware

Current middleware:

- `AuthRequired`
- `ShopRequired`
- `PlatformAdminRequired`
- `RequestID`
- `RequestLogger`
- `Recovery`
- `CORS`

## Security Notes

Implemented:

- bcrypt password hashing
- JWT auth
- token sessions
- logout/revoke support
- shop membership validation
- soft-delete style using `is_active` where needed
- transaction-safe billing
- transaction-safe stock movement
- transaction-safe udharo payment
- panic recovery middleware
- CORS protection

Do not log:

- passwords
- JWT tokens
- database passwords
- secrets

## Development Flow

Recommended build order already completed:

```txt
Auth
Shop
Product
Customer
Billing
Udharo
Inventory
Dashboard
Subscriptions
Notifications
Audit
Admin
```

## Useful Commands

Run:

```bash
go run ./cmd/api
```

Build:

```bash
go build ./cmd/api
```

Kill port 8080:

```bash
sudo fuser -k 8080/tcp
```

Check port:

```bash
sudo ss -ltnp | grep :8080
```

Format code:

```bash
gofmt -w .
```

Tidy modules:

```bash
go mod tidy
```

## Git Commit

```bash
git add .
git commit -m "Build core SaaS backend foundation"
```

## Next Improvements

Recommended next phase:

- Add Dockerfile
- Add Docker Compose
- Add Swagger/OpenAPI docs
- Add pagination helper
- Add validation helper
- Add refresh token support
- Add role-based permissions
- Add email verification
- Add password reset email
- Add reports module
- Add CSV export
- Add Prometheus metrics
- Add Grafana dashboard
- Add Loki logging
- Add Sentry error tracking
- Add production deployment config
