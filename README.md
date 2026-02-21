# Florus - Cosmetics E-commerce Platform

A full-stack e-commerce platform for cosmetics and beauty products, featuring a modern React frontend and a robust Go backend with AI-powered product recommendations.

![React](https://img.shields.io/badge/React-18.3-61DAFB?logo=react)
![TypeScript](https://img.shields.io/badge/TypeScript-5.6-3178C6?logo=typescript)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![Tailwind CSS](https://img.shields.io/badge/Tailwind_CSS-3.4-06B6D4?logo=tailwindcss)
![SQLite](https://img.shields.io/badge/SQLite-Database-003B57?logo=sqlite)

## Introduction

Florus is a comprehensive e-commerce solution designed for cosmetics and beauty product retailers. The platform provides a seamless shopping experience for customers while offering powerful management tools for administrators.

### Key Highlights
- Modern, responsive UI with dark/light mode support
- JWT-based authentication system
- AI-powered product recommendation engine (cold-start & warm-start)
- Complete order management workflow
- Real-time inventory tracking

## Features

### Customer Features
- **Authentication** - Register/Login with JWT tokens
- **Product Browsing** - Browse products with advanced filtering (category, price range, brand)
- **Search** - Full-text product search
- **Product Details** - View detailed product information with image gallery
- **Reviews & Ratings** - Rate and review purchased products
- **Smart Recommendations** - AI-powered product suggestions based on user behavior
- **Shopping Cart** - Add/remove items, update quantities
- **Wishlist** - Save favorite products for later
- **Checkout** - Place orders with coupon code support
- **Order History** - Track order status and history
- **Profile Management** - Update personal information and change password

### Admin Features
- **Dashboard** - Real-time statistics and analytics
- **Product Management** - Full CRUD operations for products
- **Category Management** - Organize products into categories
- **Order Management** - View, update, and manage customer orders
- **User Management** - View and manage registered users
- **Coupon Management** - Create and manage discount coupons

## Tech Stack

### Frontend
| Technology | Version | Purpose |
|------------|---------|---------|
| React | 18.3 | UI Library |
| TypeScript | 5.6 | Type Safety |
| Vite | 6.3 | Build Tool |
| Tailwind CSS | 3.4 | Styling |
| Radix UI | - | UI Components |
| React Router DOM | 7.1 | Routing |
| Axios | 1.7 | HTTP Client |
| Recharts | 2.15 | Charts & Analytics |

### Backend
| Technology | Version | Purpose |
|------------|---------|---------|
| Go | 1.21+ | Programming Language |
| Gin-Gonic | 1.10 | Web Framework |
| GORM | 1.25 | ORM |
| SQLite | - | Database |
| JWT-Go | 5.2 | Authentication |
| Bcrypt | - | Password Hashing |

## Project Structure

```
florus_datn/
├── src/                          # Frontend source code
│   ├── components/               # Reusable UI components
│   │   ├── admin/               # Admin-specific components
│   │   ├── auth/                # Authentication components
│   │   ├── cart/                # Shopping cart components
│   │   ├── layout/              # Layout components
│   │   ├── product/             # Product-related components
│   │   └── ui/                  # Base UI components (Radix)
│   ├── contexts/                # React context providers
│   ├── hooks/                   # Custom React hooks
│   ├── lib/                     # Utility libraries
│   ├── pages/                   # Page components
│   │   ├── admin/               # Admin pages
│   │   └── ...                  # Customer pages
│   ├── services/                # API service functions
│   └── types/                   # TypeScript type definitions
│
├── backend/                      # Backend source code
│   ├── cmd/
│   │   └── api/                 # Application entry point
│   ├── internal/
│   │   ├── config/              # Configuration
│   │   ├── dto/                 # Data Transfer Objects
│   │   ├── handlers/            # HTTP handlers
│   │   ├── middleware/          # Middleware (auth, cors)
│   │   ├── models/              # Database models
│   │   ├── repository/          # Data access layer
│   │   ├── routes/              # Route definitions
│   │   └── services/            # Business logic
│   └── pkg/
│       └── utils/               # Utility functions
│
├── public/                       # Static assets
└── package.json                  # Frontend dependencies
```

## Installation & Setup

### Prerequisites
- Node.js 18+
- Go 1.21+
- Git

### Step 1: Clone the Repository
```bash
git clone <repository-url>
cd florus_datn
```

### Step 2: Setup Frontend
```bash
# Install dependencies
npm install

# Start development server
npm run dev
```

### Step 3: Setup Backend
```bash
# Navigate to backend directory
cd backend

# Download Go dependencies
go mod tidy

# Run the server
go run ./cmd/api
```

### Step 4: Access the Application
- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8081

## Default Accounts

| Role | Email | Password |
|------|-------|----------|
| Admin | admin@florus.com | admin123 |
| User | Register a new account | - |

## API Endpoints

### Authentication
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/auth/register` | Register new user |
| POST | `/api/auth/login` | User login |
| GET | `/api/auth/me` | Get current user |

### Products
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/products` | List all products |
| GET | `/api/products/:id` | Get product details |
| GET | `/api/products/search` | Search products |

### Categories
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/categories` | List all categories |
| GET | `/api/categories/:id` | Get category details |

### Cart
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/cart` | Get user's cart |
| POST | `/api/cart` | Add item to cart |
| PUT | `/api/cart/:id` | Update cart item |
| DELETE | `/api/cart/:id` | Remove from cart |

### Orders
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/orders` | List user's orders |
| POST | `/api/orders` | Create new order |
| GET | `/api/orders/:id` | Get order details |

### Wishlist
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/wishlist` | Get user's wishlist |
| POST | `/api/wishlist` | Add to wishlist |
| DELETE | `/api/wishlist/:id` | Remove from wishlist |

### Reviews
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/products/:id/reviews` | Get product reviews |
| POST | `/api/products/:id/reviews` | Add review |

### Recommendations
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/recommendations` | Get personalized recommendations |

### Coupons
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/coupons/validate` | Validate coupon code |

### Admin Endpoints
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/admin/dashboard` | Dashboard statistics |
| GET/POST/PUT/DELETE | `/api/admin/products/*` | Product management |
| GET/POST/PUT/DELETE | `/api/admin/categories/*` | Category management |
| GET/PUT | `/api/admin/orders/*` | Order management |
| GET | `/api/admin/users` | User management |
| GET/POST/PUT/DELETE | `/api/admin/coupons/*` | Coupon management |

## Environment Configuration

The backend uses the following configuration (configurable via environment variables or config file):

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 8081 | Server port |
| `JWT_SECRET` | (auto-generated) | JWT signing secret |
| `DB_PATH` | ./florus.db | SQLite database path |
| `CORS_ORIGINS` | http://localhost:3000 | Allowed CORS origins |

## Database Schema

### Core Tables
- **users** - User accounts and profiles
- **products** - Product catalog
- **categories** - Product categories
- **product_images** - Product image gallery
- **cart_items** - Shopping cart items
- **orders** - Customer orders
- **order_items** - Order line items
- **reviews** - Product reviews and ratings
- **wishlists** - User wishlists
- **coupons** - Discount coupons
- **user_interactions** - User behavior tracking (for recommendations)

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Authors

**nhom baox3**

---

*Built with React, Go, and Tailwind CSS*
