# Unify Nepal API Endpoints

Base URL:

```txt
http://localhost:8080
```

Version prefix:

```txt
/api/v1
```

## Required Headers

For protected APIs:

```http
Authorization: Bearer <JWT_TOKEN>
```

For shop-scoped APIs:

```http
X-Shop-ID: <SHOP_ID>
```

For JSON requests:

```http
Content-Type: application/json
```

---

# System

## Root

```http
GET /
```

Response:

```json
{
  "name": "Unify Nepal API",
  "status": "running",
  "version": "1.0.0"
}
```

## Health Check

```http
GET /health
```

## Readiness Check

```http
GET /api/v1/ready
```

---

# Auth

## Signup

```http
POST /api/v1/auth/signup
```

Request:

```json
{
  "name": "Bibas",
  "email": "bibas@example.com",
  "password": "password123",
  "shop_name": "Bibas Store"
}
```

Creates:

- user
- shop
- shop member with owner role
- 30-day Starter trial subscription
- user session
- JWT token

Response:

```json
{
  "message": "account created",
  "token": "jwt_token",
  "user": {},
  "shop": {},
  "session": {}
}
```

## Login

```http
POST /api/v1/auth/login
```

Request:

```json
{
  "email": "bibas@example.com",
  "password": "password123"
}
```

Response:

```json
{
  "message": "login successful",
  "token": "jwt_token",
  "user": {},
  "session": {}
}
```

## Logout

```http
POST /api/v1/auth/logout
```

Headers:

```http
Authorization: Bearer <JWT_TOKEN>
```

Response:

```json
{
  "message": "logged out successfully"
}
```

---

# Products

All product endpoints require:

```http
Authorization: Bearer <JWT_TOKEN>
X-Shop-ID: <SHOP_ID>
```

## List Products

```http
GET /api/v1/products
```

## Create Product

```http
POST /api/v1/products
```

Request:

```json
{
  "name": "Coca Cola 500ml",
  "sku": "COCA-500",
  "category": "Drinks",
  "unit": "piece",
  "selling_price": 80,
  "cost_price": 65,
  "stock_qty": 50,
  "min_stock_qty": 10
}
```

## Get Product

```http
GET /api/v1/products/:id
```

## Update Product

```http
PUT /api/v1/products/:id
```

## Delete Product

```http
DELETE /api/v1/products/:id
```

---

# Customers

All customer endpoints require:

```http
Authorization: Bearer <JWT_TOKEN>
X-Shop-ID: <SHOP_ID>
```

## List Customers

```http
GET /api/v1/customers
```

Optional query:

```txt
?search=ram
```

## Create Customer

```http
POST /api/v1/customers
```

Request:

```json
{
  "name": "Ram Bahadur",
  "phone": "9800000000",
  "address": "Kathmandu"
}
```

## Get Customer

```http
GET /api/v1/customers/:id
```

## Update Customer

```http
PUT /api/v1/customers/:id
```

Request:

```json
{
  "name": "Ram Bahadur",
  "phone": "9800000001",
  "address": "Lalitpur",
  "is_active": true
}
```

## Delete Customer

```http
DELETE /api/v1/customers/:id
```

This performs soft delete using `is_active = false`.

---

# Billing

All billing endpoints require:

```http
Authorization: Bearer <JWT_TOKEN>
X-Shop-ID: <SHOP_ID>
```

## List Bills

```http
GET /api/v1/bills
```

## Create Bill

```http
POST /api/v1/bills
```

Request:

```json
{
  "customer_id": "CUSTOMER_UUID",
  "discount": 0,
  "paid_amount": 50,
  "payment_method": "cash",
  "items": [
    {
      "product_id": "PRODUCT_UUID",
      "quantity": 2,
      "unit_price": 80
    }
  ]
}
```

Business logic:

```txt
subtotal = sum(item.quantity * item.unit_price)
total_amount = subtotal - discount
due_amount = total_amount - paid_amount
payment_status = paid / partial / unpaid
```

Transaction actions:

- creates bill
- creates bill items
- reduces product stock
- creates stock movement with type `sale`
- creates payment if `paid_amount > 0`
- creates customer ledger entry if `due_amount > 0`

## Get Bill

```http
GET /api/v1/bills/:id
```

---

# Udharo

All udharo endpoints require:

```http
Authorization: Bearer <JWT_TOKEN>
X-Shop-ID: <SHOP_ID>
```

## Udharo Summary

```http
GET /api/v1/udharo/summary
```

Response:

```json
{
  "data": {
    "total_due": 50,
    "total_credit": 110,
    "total_payment": 60,
    "customers_with_due": 1,
    "total_ledger_customers": 1
  }
}
```

## Customers With Due

```http
GET /api/v1/udharo/customers
```

## Customer Ledger

```http
GET /api/v1/customers/:id/udharo/ledger
```

Response includes:

- customer
- ledger entries
- total credit
- total payment
- current due

## Apply Udharo Payment

```http
POST /api/v1/customers/:id/udharo/payments
```

Request without bill link:

```json
{
  "amount": 60,
  "method": "cash",
  "note": "Partial udharo collection"
}
```

Request with bill link:

```json
{
  "amount": 50,
  "method": "cash",
  "bill_id": "BILL_UUID",
  "note": "Final udharo collection for bill"
}
```

If `bill_id` is provided:

- updates bill paid amount
- updates bill due amount
- updates payment status
- creates payment row
- creates ledger payment entry

Ledger formula:

```txt
Current Due = SUM(credit) - SUM(payment)
```

---

# Inventory

All inventory endpoints require:

```http
Authorization: Bearer <JWT_TOKEN>
X-Shop-ID: <SHOP_ID>
```

## List Stock Movements

```http
GET /api/v1/stock-movements
```

## Create Stock Movement

```http
POST /api/v1/stock-movements
```

Supported types:

```txt
stock_in
stock_out
adjustment
```

Request:

```json
{
  "product_id": "PRODUCT_UUID",
  "type": "stock_in",
  "quantity": 10,
  "note": "Added 10 bottles from supplier"
}
```

Logic:

```txt
stock_in    => current_stock + quantity
stock_out   => current_stock - quantity
adjustment  => stock_qty becomes quantity
```

## Low Stock Products

```http
GET /api/v1/inventory/low-stock
```

Returns products where:

```txt
stock_qty <= min_stock_qty
```

---

# Dashboard

Requires:

```http
Authorization: Bearer <JWT_TOKEN>
X-Shop-ID: <SHOP_ID>
```

## Dashboard Stats

```http
GET /api/v1/dashboard/stats
```

Response:

```json
{
  "data": {
    "today_sales": 160,
    "month_sales": 160,
    "total_due": 0,
    "total_products": 1,
    "low_stock_count": 0,
    "total_customers": 1,
    "total_bills": 1
  }
}
```

---

# Subscriptions

All subscription endpoints require:

```http
Authorization: Bearer <JWT_TOKEN>
X-Shop-ID: <SHOP_ID>
```

## List Plans

```http
GET /api/v1/plans
```

Seeded plans:

| Plan | Price | Max Users | Max Products | Max Bills / Month |
|---|---:|---:|---:|---:|
| Starter | Rs 499 | 1 | 100 | 300 |
| Business | Rs 999 | 3 | 1000 | 3000 |
| Pro | Rs 1999 | 10 | 10000 | 10000 |

## Current Subscription

```http
GET /api/v1/subscription
```

Response:

```json
{
  "data": {
    "shop_id": "SHOP_UUID",
    "plan_id": "PLAN_UUID",
    "status": "trialing",
    "trial_ends_at": "2026-06-03T02:42:35+05:45",
    "current_period_start": "2026-05-04T02:42:35+05:45",
    "current_period_end": "2026-06-03T02:42:35+05:45"
  }
}
```

## Upgrade Request

```http
POST /api/v1/subscription/upgrade-request
```

Request:

```json
{
  "plan_id": "PLAN_UUID",
  "note": "I want to upgrade to Business plan"
}
```

Creates:

- notification
- audit log

---

# Notifications

Requires:

```http
Authorization: Bearer <JWT_TOKEN>
```

`X-Shop-ID` is optional but recommended for shop-specific filtering.

## List Notifications

```http
GET /api/v1/notifications
```

Response:

```json
{
  "data": [],
  "meta": {
    "unread_count": 1
  }
}
```

## Mark Notification Read

```http
POST /api/v1/notifications/:id/read
```

## Mark All Notifications Read

```http
POST /api/v1/notifications/read-all
```

---

# Audit Logs

Requires:

```http
Authorization: Bearer <JWT_TOKEN>
X-Shop-ID: <SHOP_ID>
```

## List Audit Logs

```http
GET /api/v1/audit/logs
```

Response:

```json
{
  "data": [
    {
      "action": "subscription_upgrade_requested",
      "entity_type": "subscription_plan",
      "metadata": {
        "plan_name": "Business",
        "note": "I want to upgrade to Business plan"
      }
    }
  ]
}
```

---

# Admin

Requires:

```http
Authorization: Bearer <JWT_TOKEN>
```

User must have:

```txt
role = platform_admin
```

## List Shops

```http
GET /api/v1/admin/shops
```

## List Users

```http
GET /api/v1/admin/users
```

## Update Shop Status

```http
PATCH /api/v1/admin/shops/:id/status
```

Request:

```json
{
  "status": "active"
}
```

Supported status values:

```txt
active
paused
suspended
```

---

# Common Response Patterns

## Success

```json
{
  "data": {}
}
```

Some endpoints also return:

```json
{
  "message": "operation successful",
  "data": {}
}
```

## Error

```json
{
  "message": "error message"
}
```

## Error With Request ID

Recovery middleware returns:

```json
{
  "success": false,
  "message": "internal server error",
  "request_id": "uuid"
}
```

---

# Request Tracking

Every response includes:

```http
X-Request-ID: uuid
```

Example:

```bash
curl -i -X GET http://localhost:8080/api/v1/dashboard/stats \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Shop-ID: $SHOP_ID"
```

---

# cURL Testing Flow

## 1. Signup

```bash
curl -X POST http://localhost:8080/api/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Trial Test",
    "email": "trialtest@example.com",
    "password": "password123",
    "shop_name": "Trial Store"
  }'
```

## 2. Export Token and Shop

```bash
export TOKEN='PASTE_TOKEN'
export SHOP_ID='PASTE_SHOP_ID'
```

## 3. Create Product

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

## 4. Create Customer

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

## 5. Create Bill

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

## 6. Check Udharo

```bash
curl -X GET http://localhost:8080/api/v1/udharo/summary \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Shop-ID: $SHOP_ID"
```

## 7. Collect Udharo Payment

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

## 8. Check Dashboard

```bash
curl -X GET http://localhost:8080/api/v1/dashboard/stats \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Shop-ID: $SHOP_ID"
```

---

# Notes

## Multi-Tenant Isolation

Every shop-scoped endpoint validates:

```txt
user_id + shop_id
```

against:

```txt
shop_members
```

## JWT Session

JWT contains:

```json
{
  "user_id": "uuid",
  "email": "user@example.com",
  "token_id": "uuid",
  "jti": "uuid",
  "exp": 123456789,
  "iat": 123456789
}
```

`token_id` is checked against `user_sessions`.

## Logout

Logout sets:

```txt
user_sessions.is_active = false
user_sessions.revoked_at = now()
```

## Billing Safety

Billing runs in a transaction. If any step fails, everything rolls back.

## Stock Safety

Stock movement runs in a transaction.

## Udharo Safety

Udharo payment runs in a transaction.
