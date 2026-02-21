# Florus E-Commerce Backend API

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version">
  <img src="https://img.shields.io/badge/Gin-1.9+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Gin">
  <img src="https://img.shields.io/badge/GORM-1.25+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="GORM">
  <img src="https://img.shields.io/badge/SQLite-3-003B57?style=for-the-badge&logo=sqlite&logoColor=white" alt="SQLite">
  <img src="https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker&logoColor=white" alt="Docker">
</p>

Backend API cho hệ thống E-Commerce mỹ phẩm Florus với hệ thống gợi ý sản phẩm (Recommendation System).

## 🚀 Tech Stack

| Component | Technology |
|-----------|------------|
| **Language** | Go 1.21+ |
| **Framework** | Gin-Gonic |
| **ORM** | GORM |
| **Database** | SQLite (Dev) / Oracle (Prod) |
| **Authentication** | JWT |
| **Configuration** | Viper |
| **Container** | Docker + Docker Compose |

## 📁 Project Structure

```
backend/
├── cmd/
│   └── api/
│       └── main.go                 # Entry point
├── internal/
│   ├── config/                     # Configuration
│   ├── dto/                        # Data Transfer Objects
│   ├── handler/                    # HTTP Handlers
│   ├── middleware/                 # Middleware (CORS, Auth, Logger)
│   ├── models/                     # GORM Models
│   ├── repository/                 # Data Access Layer
│   ├── router/                     # Route Definitions
│   └── service/                    # Business Logic
├── migrations/                     # Database Migrations & Seeds
├── pkg/
│   ├── database/                   # Database Connection
│   └── utils/                      # Utilities (JWT, Pagination, Response)
├── docker-compose.yaml
├── Dockerfile
├── Makefile
└── README.md
```

## 🛠 Quick Start

### Prerequisites
- Go 1.21+
- Docker & Docker Compose (optional)

### 1. Clone & Setup

```bash
cd florus_datn/backend

# Copy environment file
cp .env.example .env

# Download dependencies
go mod tidy
```

### 2. Run Application

```bash
# Development mode (with SQLite)
go run ./cmd/api

# Or using Make
make run
```

### 3. Run with Docker

```bash
# Start with SQLite (simple)
docker-compose --profile sqlite up -d api-sqlite

# Start with Oracle (full)
docker-compose up -d
```

## 📡 API Endpoints

### Health Check
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/health` | Health check |

### Authentication
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/auth/register` | Register new user |
| POST | `/api/auth/login` | Login |
| GET | `/api/auth/me` | Get current user (Auth required) |

### Categories
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/categories` | List all categories |
| GET | `/api/categories/:id` | Get category by ID |
| POST | `/api/categories` | Create category (Admin) |
| PUT | `/api/categories/:id` | Update category (Admin) |
| DELETE | `/api/categories/:id` | Delete category (Admin) |

### Products
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/products` | List products (with pagination) |
| GET | `/api/products/:id` | Get product by ID |
| GET | `/api/products/slug/:slug` | Get product by slug |
| GET | `/api/products/category/:categoryId` | Products by category |
| POST | `/api/products` | Create product (Admin) |
| PUT | `/api/products/:id` | Update product (Admin) |
| DELETE | `/api/products/:id` | Delete product (Admin) |

**Query Parameters:**
- `page` - Page number (default: 1)
- `limit` - Items per page (default: 10, max: 100)
- `category_id` - Filter by category
- `search` - Search by name, brand, description

### Cart
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/cart` | Get current cart |
| POST | `/api/cart` | Add item to cart |
| PUT | `/api/cart/:itemId` | Update item quantity |
| DELETE | `/api/cart/:itemId` | Remove item |
| DELETE | `/api/cart` | Clear cart |

### Orders
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/orders` | List user orders |
| GET | `/api/orders/:id` | Get order details |
| POST | `/api/orders` | Create order (Checkout) |
| PUT | `/api/orders/:id/cancel` | Cancel order |

### Recommendations
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/recommendations/cold-start` | Popular products (new users) |
| GET | `/api/recommendations/warm-start` | Personalized (active users) |
| GET | `/api/recommendations/similar/:productId` | Similar products |
| GET | `/api/recommendations/co-viewed/:productId` | Co-viewed products |
| GET | `/api/recommendations/cross-sell/:productId` | Cross-sell products |

## 🔐 Authentication

### Register
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "name": "John Doe"
  }'
```

### Login
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

### Use Token
```bash
curl -X GET http://localhost:8080/api/cart \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 🛒 Order Flow

1. **Add to Cart**
```bash
curl -X POST http://localhost:8080/api/cart \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"product_id": 1, "quantity": 2}'
```

2. **Checkout (Create Order)**
```bash
curl -X POST http://localhost:8080/api/orders \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "shipping_address": "123 Main St, City",
    "note": "Please deliver before 5pm"
  }'
```

## 📊 Database Schema

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│    Users     │     │  Categories  │     │   Products   │
├──────────────┤     ├──────────────┤     ├──────────────┤
│ id           │     │ id           │     │ id           │
│ email        │     │ name         │     │ name         │
│ password_hash│     │ slug         │     │ slug         │
│ name         │     │ description  │     │ brand        │
│ role         │     │ parent_id    │     │ price        │
│ user_status  │     │ is_active    │     │ category_id  │◄─┐
└──────────────┘     └──────────────┘     │ stock        │  │
       │                                   │ rating       │  │
       │                                   └──────────────┘  │
       │                                          │          │
       ▼                                          │          │
┌──────────────┐     ┌──────────────┐            │          │
│  Cart Items  │     │    Orders    │            │          │
├──────────────┤     ├──────────────┤            │          │
│ id           │     │ id           │            │          │
│ user_id      │◄────│ user_id      │            ▼          │
│ product_id   │────►│ order_code   │     ┌──────────────┐  │
│ quantity     │     │ subtotal     │     │ Order Items  │  │
└──────────────┘     │ shipping_fee │     ├──────────────┤  │
                     │ total        │     │ order_id     │◄─┤
                     │ status       │     │ product_id   │──┘
                     └──────────────┘     │ quantity     │
                                          │ unit_price   │
                                          └──────────────┘
```

## 🔧 Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | Server port | 8080 |
| `SERVER_MODE` | Gin mode (debug/release) | debug |
| `JWT_SECRET` | JWT signing key | - |
| `JWT_EXPIRATION_HOURS` | Token expiration | 24 |

## 📝 Default Admin Account

After running the application, a default admin account is created:

- **Email:** `admin@florus.com`
- **Password:** `admin123`

## 🧪 Testing

```bash
# Run all tests
make test

# With coverage
make test-coverage
```

## 📦 Available Make Commands

```bash
make help           # Show all commands
make build          # Build application
make run            # Build and run
make dev            # Run with hot reload
make test           # Run tests
make docker-up      # Start with Docker
make docker-down    # Stop Docker services
```

## 🏗 Business Logic

### Shipping Fee
- Free shipping for orders ≥ 5,000,000 VND
- 50,000 VND shipping fee otherwise

### Order Transaction
1. Validate cart items
2. Check stock availability (with row locking)
3. Calculate totals
4. Create order & order items
5. Deduct stock
6. Clear cart
7. Update user status to "warm"

### User Status
- **Cold**: New user (no orders)
- **Warm**: Active user (has orders)

## 🤝 API Response Format

### Success
```json
{
  "success": true,
  "message": "Products retrieved successfully",
  "data": [...],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 13,
    "total_pages": 2
  }
}
```

### Error
```json
{
  "success": false,
  "message": "Bad Request",
  "error": "Invalid product ID"
}
```

## 📄 License

MIT License - Florus DATN Project
