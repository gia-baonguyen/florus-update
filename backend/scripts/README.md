# Florus Beauty - Data Pipeline Scripts

Scripts để import dữ liệu thật từ Kaggle và xây dựng hệ thống recommendation.

## Kaggle Datasets

Hệ thống sử dụng 3 dataset thật từ Kaggle:

| Dataset | Mô tả | Số lượng |
|---------|-------|----------|
| [Sephora Products & Reviews](https://kaggle.com/datasets/nadyinky/sephora-products-and-skincare-reviews) | Sản phẩm Sephora với đầy đủ thông tin | 8,000+ products |
| [Cosmetics Dataset](https://kaggle.com/datasets/kingabzpro/cosmetics-datasets) | Sản phẩm skincare với thành phần | 1,400+ products |
| [Amazon Beauty Products](https://kaggle.com/datasets/skillsmuggler/amazon-ratings) | Đánh giá sản phẩm làm đẹp | 2,000,000+ ratings |

## Cài đặt

### 1. Cài đặt Python dependencies

```bash
cd backend/scripts
pip install -r requirements.txt
```

### 2. Cấu hình Kaggle API

1. Tạo tài khoản Kaggle: https://www.kaggle.com
2. Vào **Account Settings** → **API** → **Create New API Token**
3. Download file `kaggle.json`
4. Lưu file vào:
   - **Windows**: `C:\Users\<username>\.kaggle\kaggle.json`
   - **Linux/Mac**: `~/.kaggle/kaggle.json`

## Scripts

### 1. Download Kaggle Data (`download_kaggle_data.py`)

Tải tất cả 3 dataset từ Kaggle:

```bash
python download_kaggle_data.py
```

Output:
```
============================================================
FLORUS BEAUTY - KAGGLE DATA DOWNLOADER
============================================================
Target directory: e:\...\backend\data

DATASET 1/3: Sephora Products & Skincare Reviews
[DOWNLOAD] Downloading nadyinky/sephora-products-and-skincare-reviews...
[DOWNLOAD] ✓ Successfully downloaded
  ✓ product_info.csv (5.23 MB)
  ✓ reviews_0-250.csv (180.45 MB)
...
```

### 2. Import Products (`import_sephora_data.py`)

Import sản phẩm từ Kaggle datasets vào database:

```bash
# Import tất cả (yêu cầu Kaggle API)
python import_sephora_data.py

# Giới hạn số sản phẩm mỗi dataset
python import_sephora_data.py --limit 500
```

**Lưu ý**: Script yêu cầu:
- Kaggle API credentials đã cấu hình
- Backend API đang chạy tại `localhost:8081`

### 3. Generate User Events (`generate_user_events.py`)

Tạo user events để test recommendation system:

```bash
# Mặc định: 50 users, 20 events/user
python generate_user_events.py

# Tùy chỉnh số lượng
python generate_user_events.py --users 100 --events 50
```

## Data Flow

```
┌─────────────────────────────────────────────────────────────────────┐
│                         1. DOWNLOAD DATA                            │
│   download_kaggle_data.py → Kaggle API → backend/data/*.csv         │
│                                                                     │
│   Files: product_info.csv, cosmetics.csv, ratings_Beauty.csv       │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         2. IMPORT PRODUCTS                          │
│   import_sephora_data.py → Transform → API → Products DB            │
│                                                                     │
│   - Sephora: 8K+ products (skincare, makeup, hair, fragrance)      │
│   - Cosmetics: 1.4K+ products (skincare với ingredients)           │
│   - Total: ~10K products với real data                              │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                       3. GENERATE USER EVENTS                       │
│   generate_user_events.py → API → User Events DB                    │
│                                                                     │
│   Event types: product_view, product_click, add_to_cart,           │
│                add_to_wishlist, purchase, review                    │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    4. COMPUTE RELATIONSHIPS                         │
│   Analytics Service → Product Relationships → Recommendations        │
│                                                                     │
│   - Co-viewed: "Users who viewed X also viewed Y"                  │
│   - Frequently Bought Together: "Thường mua cùng"                   │
│   - Similar Products: Content-based similarity                      │
└─────────────────────────────────────────────────────────────────────┘
```

## Recommendation Strategies

### Cold-Start (User mới, chưa có history)

| Strategy | Mô tả | Endpoint |
|----------|-------|----------|
| Popular | Sản phẩm được xem nhiều nhất | `/api/recommendations/cold-start` |
| Trending | Sản phẩm hot trong 7 ngày | `/api/recommendations/trending` |
| Top-rated | Sản phẩm đánh giá cao | `/api/recommendations/cold-start` |

### Warm-Start (User có history)

| Strategy | Mô tả | Endpoint |
|----------|-------|----------|
| Personalized | Dựa trên lịch sử user | `/api/recommendations/warm-start` |
| Similar | Sản phẩm tương tự đã xem | `/api/recommendations/similar/:id` |
| Co-viewed | "Users also viewed" | `/api/recommendations/co-viewed/:id` |
| Cross-sell | Sản phẩm hay mua cùng | `/api/recommendations/cross-sell/:id` |

## API Endpoints

```bash
# Cold-start recommendations
GET /api/recommendations/cold-start?limit=10

# Warm-start recommendations (requires auth)
GET /api/recommendations/warm-start?limit=10

# Similar products
GET /api/recommendations/similar/:productId?limit=6

# Co-viewed products
GET /api/recommendations/co-viewed/:productId?limit=6

# Cross-sell products
GET /api/recommendations/cross-sell/:productId?limit=4

# Trending products
GET /api/recommendations/trending?limit=10

# Popular by category
GET /api/recommendations/category/:categoryId?limit=10
```

## Troubleshooting

### Kaggle API Error

```
[ERROR] Kaggle authentication failed
```

**Giải pháp**: Kiểm tra file `kaggle.json` đã được đặt đúng vị trí và có quyền đọc.

### API Connection Error

```
[AUTH] ✗ Login failed
```

**Giải pháp**: Đảm bảo backend API đang chạy:
```bash
cd backend
go run ./cmd/api
```

### Import Duplicates

Script tự động phát hiện và bỏ qua sản phẩm trùng lặp dựa trên slug.

## File Structure

```
backend/
├── data/                    # Downloaded Kaggle datasets
│   ├── product_info.csv     # Sephora products
│   ├── cosmetics.csv        # Cosmetics dataset
│   ├── ratings_Beauty.csv   # Amazon ratings
│   └── reviews_*.csv        # Sephora reviews
│
├── scripts/
│   ├── download_kaggle_data.py   # Download datasets
│   ├── import_sephora_data.py    # Import products
│   ├── generate_user_events.py   # Generate events
│   ├── requirements.txt
│   └── README.md
│
└── internal/
    └── service/
        ├── recommendation_service.go  # Recommendation logic
        └── analytics_service.go       # Analytics computation
```
