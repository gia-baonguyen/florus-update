#!/usr/bin/env python3
"""
Florus Beauty - Kaggle Data Importer
=====================================
Import real beauty product data from multiple Kaggle datasets:

1. Sephora Products & Reviews (8K+ products, 1M+ reviews)
   - Dataset: nadyinky/sephora-products-and-skincare-reviews
   - Files: product_info.csv, reviews_*.csv

2. Cosmetics Dataset (1.4K+ products)
   - Dataset: kingabzpro/cosmetics-datasets
   - Files: cosmetics.csv

3. Amazon Beauty Products (2M+ ratings)
   - Dataset: skillsmuggler/amazon-ratings
   - Files: ratings_Beauty.csv

Usage:
    1. Install: pip install pandas requests kaggle
    2. Set up Kaggle API credentials (~/.kaggle/kaggle.json)
    3. Run: python import_sephora_data.py

Requirements:
    - Kaggle API credentials (kaggle.json)
    - Backend API running at localhost:8081
"""

import os
import sys
import json
import time
import hashlib
import requests
import pandas as pd
from pathlib import Path
from typing import Dict, List, Optional, Tuple
from datetime import datetime

# Configuration
API_BASE_URL = "http://localhost:8081/api"
ADMIN_EMAIL = "admin@florus.com"
ADMIN_PASSWORD = "admin123"
DATA_DIR = Path(__file__).parent.parent / "data"

# Kaggle datasets configuration
KAGGLE_DATASETS = {
    "sephora": {
        "name": "Sephora Products & Reviews",
        "kaggle_id": "nadyinky/sephora-products-and-skincare-reviews",
        "product_file": "product_info.csv",
        "review_file": "reviews_0-250.csv",
        "priority": 1  # Primary dataset
    },
    "cosmetics": {
        "name": "Cosmetics Dataset",
        "kaggle_id": "kingabzpro/cosmetics-datasets",
        "product_file": "cosmetics.csv",
        "review_file": None,
        "priority": 2
    },
    "amazon": {
        "name": "Amazon Beauty Products",
        "kaggle_id": "skillsmuggler/amazon-ratings",
        "product_file": None,
        "review_file": "ratings_Beauty.csv",
        "priority": 3  # Ratings only - used for user events
    }
}

# Category mapping from various sources to Florus
CATEGORY_MAPPING = {
    # Primary categories (Sephora)
    "Skincare": "skincare",
    "Makeup": "makeup",
    "Hair": "hair",
    "Fragrance": "fragrance",
    "Bath & Body": "bath-body",
    "Tools & Brushes": "tools",
    "Mini Size": "mini",
    "Gifts": "gifts",
    "Men": "men",
    "Wellness": "wellness",

    # Secondary categories (Sephora)
    "Hair Styling & Treatments": "hair-styling",
    "Hair Styling": "hair-styling",
    "Eye": "eye",
    "Face": "face",
    "Moisturizers": "moisturizer",
    "Treatments": "treatments",
    "Shampoo & Conditioner": "shampoo-conditioner",
    "Lip": "lip",
    "Cleansers": "cleanser",
    "Candles & Home Scents": "candles-home-scents",
    "Brushes & Applicators": "brushes-applicators",
    "Body Moisturizers": "body-moisturizers",
    "Eye Care": "eye",
    "Masks": "mask",
    "Cheek": "cheek",
    "Sunscreen": "sunscreen",
    "Bath & Shower": "bath-shower",
    "High Tech Tools": "high-tech-tools",
    "Body Care": "body-care",
    "Self Tanners": "self-tanners",
    "Lip Balms & Treatments": "lip-balms",
    "Nail": "nail",
    "Accessories": "accessories",
    "Value & Gift Sets": "gift-sets",

    # Cosmetics dataset categories
    "Face Masks": "mask",
    "Lip Care": "lip",
    "Sun Protect": "sunscreen",

    # General mappings
    "Body": "bath-body",
    "Hair Care": "hair",
}

# Default categories to create
DEFAULT_CATEGORIES = [
    # Primary categories
    {"name": "Skincare", "slug": "skincare", "description": "Face creams, serums, cleansers and more"},
    {"name": "Makeup", "slug": "makeup", "description": "Lipstick, foundation, eyeshadow and more"},
    {"name": "Hair", "slug": "hair", "description": "Shampoo, conditioner, styling products"},
    {"name": "Fragrance", "slug": "fragrance", "description": "Perfumes and body sprays"},
    {"name": "Bath & Body", "slug": "bath-body", "description": "Body lotions, shower gels"},
    {"name": "Tools & Brushes", "slug": "tools", "description": "Makeup brushes and beauty tools"},
    {"name": "Mini Size", "slug": "mini", "description": "Travel size products"},
    {"name": "Gifts", "slug": "gifts", "description": "Gift sets and bundles"},
    {"name": "Men", "slug": "men", "description": "Men's grooming products"},
    {"name": "Wellness", "slug": "wellness", "description": "Supplements and wellness products"},
    # Secondary categories (Sephora)
    {"name": "Hair Styling", "slug": "hair-styling", "description": "Hair styling products and treatments"},
    {"name": "Face", "slug": "face", "description": "Face care and makeup products"},
    {"name": "Treatments", "slug": "treatments", "description": "Skincare treatments and serums"},
    {"name": "Shampoo & Conditioner", "slug": "shampoo-conditioner", "description": "Hair cleansing and conditioning"},
    {"name": "Candles & Home Scents", "slug": "candles-home-scents", "description": "Candles and home fragrance"},
    {"name": "Brushes & Applicators", "slug": "brushes-applicators", "description": "Makeup brushes and applicators"},
    {"name": "Lip", "slug": "lip", "description": "Lip care and lip makeup products"},
    {"name": "Eye", "slug": "eye", "description": "Eye makeup and eye care products"},
    {"name": "Cheek", "slug": "cheek", "description": "Cheek makeup - blush, bronzer, highlighter"},
    {"name": "Self Tanners", "slug": "self-tanners", "description": "Self-tanning and bronzing products"},
    {"name": "Body Care", "slug": "body-care", "description": "Body lotions, oils and treatments"},
    {"name": "Gift Sets", "slug": "gift-sets", "description": "Value and gift sets"},
    {"name": "Accessories", "slug": "accessories", "description": "Beauty accessories and tools"},
    {"name": "High Tech Tools", "slug": "high-tech-tools", "description": "High-tech beauty devices"},
    {"name": "Bath & Shower", "slug": "bath-shower", "description": "Bath and shower products"},
    {"name": "Nail", "slug": "nail", "description": "Nail care and nail polish"},
    {"name": "Body Moisturizers", "slug": "body-moisturizers", "description": "Body lotions and moisturizers"},
    {"name": "Lip Balms", "slug": "lip-balms", "description": "Lip balms and lip treatments"},
    {"name": "Cleanser", "slug": "cleanser", "description": "Face cleansers and makeup removers"},
    {"name": "Moisturizer", "slug": "moisturizer", "description": "Face and body moisturizers"},
    {"name": "Sunscreen", "slug": "sunscreen", "description": "Sun protection products"},
    {"name": "Mask", "slug": "mask", "description": "Face masks and treatments"},
]


class KaggleDataImporter:
    """Import product data from Kaggle datasets into Florus Beauty database"""

    def __init__(self):
        self.session = requests.Session()
        self.token = None
        self.categories = {}
        self.imported_slugs = set()  # Track imported products to avoid duplicates
        self.stats = {
            "products_imported": 0,
            "products_skipped": 0,
            "products_duplicate": 0,
            "reviews_imported": 0,
            "categories_created": 0,
            "errors": []
        }

    def check_kaggle_setup(self) -> bool:
        """Verify Kaggle API is properly configured"""
        print("\n[KAGGLE] Checking Kaggle API setup...")

        try:
            import kaggle
            kaggle.api.authenticate()
            print("[KAGGLE] Kaggle API authenticated successfully!")
            return True
        except ImportError:
            print("[ERROR] Kaggle package not installed!")
            print("        Run: pip install kaggle")
            return False
        except Exception as e:
            print(f"[ERROR] Kaggle authentication failed: {e}")
            print("\n" + "=" * 60)
            print("KAGGLE SETUP INSTRUCTIONS")
            print("=" * 60)
            print("""
1. Go to https://www.kaggle.com and sign in (or create account)
2. Click on your profile picture → Settings
3. Scroll to "API" section → Click "Create New API Token"
4. This downloads kaggle.json file
5. Move kaggle.json to:
   - Windows: C:\\Users\\<username>\\.kaggle\\kaggle.json
   - Linux/Mac: ~/.kaggle/kaggle.json
6. Run this script again
""")
            return False

    def download_datasets(self) -> Dict[str, Path]:
        """Download all Kaggle datasets"""
        import kaggle

        DATA_DIR.mkdir(parents=True, exist_ok=True)
        downloaded = {}

        for key, config in KAGGLE_DATASETS.items():
            print(f"\n[DOWNLOAD] Downloading {config['name']}...")
            print(f"           Dataset: {config['kaggle_id']}")

            try:
                # Download dataset
                kaggle.api.dataset_download_files(
                    config['kaggle_id'],
                    path=str(DATA_DIR),
                    unzip=True
                )

                # Verify files exist
                if config['product_file']:
                    product_path = DATA_DIR / config['product_file']
                    if product_path.exists():
                        downloaded[f"{key}_products"] = product_path
                        size_mb = product_path.stat().st_size / (1024 * 1024)
                        print(f"[DOWNLOAD] [OK] {config['product_file']} ({size_mb:.2f} MB)")
                    else:
                        # Try to find file with different name
                        for f in DATA_DIR.glob("*.csv"):
                            if key in f.name.lower() or "product" in f.name.lower():
                                downloaded[f"{key}_products"] = f
                                print(f"[DOWNLOAD] [OK] Found: {f.name}")
                                break

                if config['review_file']:
                    review_path = DATA_DIR / config['review_file']
                    if review_path.exists():
                        downloaded[f"{key}_reviews"] = review_path
                        size_mb = review_path.stat().st_size / (1024 * 1024)
                        print(f"[DOWNLOAD] [OK] {config['review_file']} ({size_mb:.2f} MB)")
                    else:
                        # Try to find review files
                        for f in DATA_DIR.glob("*review*.csv"):
                            downloaded[f"{key}_reviews"] = f
                            print(f"[DOWNLOAD] [OK] Found: {f.name}")
                            break
                        for f in DATA_DIR.glob("*rating*.csv"):
                            downloaded[f"{key}_reviews"] = f
                            print(f"[DOWNLOAD] [OK] Found: {f.name}")
                            break

            except Exception as e:
                print(f"[DOWNLOAD] [X] Failed: {e}")
                self.stats["errors"].append(f"Download {key}: {e}")

        return downloaded

    def login(self) -> bool:
        """Authenticate with the Florus API"""
        print(f"\n[AUTH] Logging in as {ADMIN_EMAIL}...")
        try:
            response = self.session.post(
                f"{API_BASE_URL}/auth/login",
                json={"email": ADMIN_EMAIL, "password": ADMIN_PASSWORD}
            )
            if response.status_code == 200:
                result = response.json()
                # Handle wrapped response: {"success": true, "data": {"token": "..."}}
                data = result.get("data", result) if isinstance(result, dict) else result
                self.token = data.get("token") if isinstance(data, dict) else None
                if self.token:
                    self.session.headers["Authorization"] = f"Bearer {self.token}"
                    print("[AUTH] [OK] Login successful!")
                    return True
                else:
                    print(f"[AUTH] [X] No token in response: {result}")
                    return False
            else:
                print(f"[AUTH] [X] Login failed: {response.text}")
                return False
        except Exception as e:
            print(f"[AUTH] [X] Error: {e}")
            print("[AUTH] Make sure the backend API is running at localhost:8081")
            return False

    def ensure_categories(self) -> bool:
        """Create default categories if they don't exist"""
        print("\n[CATEGORIES] Setting up categories...")

        # Get existing categories
        try:
            response = self.session.get(f"{API_BASE_URL}/categories")
            if response.status_code == 200:
                data = response.json()
                # Handle wrapped response: {"success": true, "data": [...]}
                existing = data.get("data", data) if isinstance(data, dict) else data
                if existing:
                    for cat in existing:
                        self.categories[cat["slug"]] = cat["id"]
                    print(f"[CATEGORIES] Found {len(existing)} existing categories")
        except Exception as e:
            print(f"[CATEGORIES] Error fetching: {e}")

        # Create missing categories
        for cat_data in DEFAULT_CATEGORIES:
            if cat_data["slug"] not in self.categories:
                try:
                    response = self.session.post(
                        f"{API_BASE_URL}/categories",
                        json=cat_data
                    )
                    if response.status_code in [200, 201]:
                        result = response.json()
                        # Handle wrapped response
                        created = result.get("data", result) if isinstance(result, dict) else result
                        cat_id = created.get("id") if isinstance(created, dict) else None
                        if cat_id:
                            self.categories[cat_data["slug"]] = cat_id
                        else:
                            # Use a fallback ID
                            self.categories[cat_data["slug"]] = len(self.categories) + 1
                        self.stats["categories_created"] += 1
                        print(f"[CATEGORIES] [OK] Created: {cat_data['name']}")
                except Exception as e:
                    print(f"[CATEGORIES] [X] Error creating {cat_data['name']}: {e}")

        return len(self.categories) > 0

    def map_category(self, category_str: str) -> int:
        """Map source category to Florus category ID"""
        if not category_str or pd.isna(category_str):
            return self.categories.get("skincare", 1)

        category_str = str(category_str)

        # Try direct mapping
        for key, slug in CATEGORY_MAPPING.items():
            if key.lower() in category_str.lower():
                return self.categories.get(slug, self.categories.get("skincare", 1))

        # Default to skincare
        return self.categories.get("skincare", 1)

    def generate_slug(self, name: str, brand: str) -> str:
        """Generate a unique slug for a product"""
        base = f"{brand}-{name}".lower()
        # Remove special characters
        slug = "".join(c if c.isalnum() or c == " " else "" for c in base)
        slug = "-".join(slug.split())
        # Add hash for uniqueness
        hash_suffix = hashlib.md5(f"{brand}{name}".encode()).hexdigest()[:6]
        return f"{slug[:80]}-{hash_suffix}"

    def clean_value(self, value, default=""):
        """Clean a value - handle NaN, None, etc."""
        if value is None or pd.isna(value) or str(value).lower() == "nan":
            return default
        return str(value).strip()

    def transform_sephora_product(self, row: pd.Series) -> Optional[Dict]:
        """Transform a Sephora product row to Florus format"""
        try:
            # Sephora CSV columns: product_id, product_name, brand_id, brand_name,
            # loves_count, rating, reviews, size, variation_type, variation_value,
            # variation_desc, ingredients, price_usd, value_price_usd, sale_price_usd,
            # limited_edition, new, online_only, out_of_stock, sephora_exclusive,
            # highlights, primary_category, secondary_category, tertiary_category,
            # child_count, child_max_price, child_min_price

            name = self.clean_value(row.get("product_name"))
            brand = self.clean_value(row.get("brand_name"), "Unknown")

            # Skip invalid
            if not name or len(name) < 3:
                return None

            # Price handling
            price = 0
            for price_col in ["price_usd", "sale_price_usd", "value_price_usd"]:
                try:
                    p = float(row.get(price_col, 0))
                    if p > 0:
                        price = p
                        break
                except:
                    pass

            if price <= 0:
                return None

            # Convert USD to VND-like (multiply by 24000)
            price_vnd = price * 24000

            # Rating
            rating = 0
            try:
                rating = float(row.get("rating", 0))
                rating = min(5.0, max(0.0, rating))
            except:
                pass

            # Reviews count
            reviews = 0
            try:
                reviews = int(row.get("reviews", 0))
            except:
                pass

            # Category - use secondary for more specific categorization
            secondary_cat = self.clean_value(row.get("secondary_category"))
            primary_cat = self.clean_value(row.get("primary_category"), "Skincare")
            category_id = self.map_category(secondary_cat) if secondary_cat else self.map_category(primary_cat)

            # Description from highlights
            description = self.clean_value(row.get("highlights"))

            # Image URL (Sephora doesn't include images in CSV)
            image_url = ""

            slug = self.generate_slug(name, brand)

            # Check for duplicate
            if slug in self.imported_slugs:
                return None

            return {
                "name": name[:255],
                "slug": slug,
                "brand": brand[:100],
                "price": price_vnd,
                "original_price": price_vnd * 1.2 if rating >= 4 else None,
                "category_id": category_id,
                "description": description[:2000] if description else f"{brand} {name}",
                "image_url": image_url,
                "stock": 100,
                "rating": rating,
                "review_count": reviews,
                "is_active": True
            }
        except Exception as e:
            self.stats["errors"].append(f"Transform Sephora error: {e}")
            return None

    def transform_cosmetics_product(self, row: pd.Series) -> Optional[Dict]:
        """Transform a Cosmetics dataset product row to Florus format"""
        try:
            # Cosmetics CSV columns: Label, Brand, Name, Price, Rank, Ingredients, Combination, Dry, Normal, Oily, Sensitive

            name = self.clean_value(row.get("Name"))
            brand = self.clean_value(row.get("Brand"), "Unknown")
            label = self.clean_value(row.get("Label"), "Skincare")

            if not name or len(name) < 3:
                return None

            # Price handling
            price = 0
            try:
                price_str = self.clean_value(row.get("Price"), "0")
                # Remove $ and other characters
                price_str = "".join(c for c in price_str if c.isdigit() or c == ".")
                price = float(price_str) if price_str else 0
            except:
                pass

            if price <= 0:
                price = 29.99  # Default price if missing

            # Convert to VND
            price_vnd = price * 24000

            # Rating from Rank (1-5)
            rating = 4.0
            try:
                rank = float(row.get("Rank", 4))
                rating = min(5.0, max(1.0, rank))
            except:
                pass

            # Category from Label
            category_id = self.map_category(label)

            # Description from skin type suitability
            skin_types = []
            for skin in ["Combination", "Dry", "Normal", "Oily", "Sensitive"]:
                try:
                    if row.get(skin) == 1 or row.get(skin) == "1":
                        skin_types.append(skin)
                except:
                    pass

            description = f"{label} product by {brand}."
            if skin_types:
                description += f" Suitable for {', '.join(skin_types)} skin types."

            # Ingredients
            ingredients = self.clean_value(row.get("Ingredients"))
            if ingredients:
                description += f" Key ingredients: {ingredients[:200]}"

            slug = self.generate_slug(name, brand)

            if slug in self.imported_slugs:
                return None

            return {
                "name": name[:255],
                "slug": slug,
                "brand": brand[:100],
                "price": price_vnd,
                "original_price": price_vnd * 1.15,
                "category_id": category_id,
                "description": description[:2000],
                "image_url": "",
                "stock": 100,
                "rating": rating,
                "review_count": int(rating * 100),
                "is_active": True
            }
        except Exception as e:
            self.stats["errors"].append(f"Transform Cosmetics error: {e}")
            return None

    def import_product(self, product_data: Dict) -> bool:
        """Import a single product to the database"""
        try:
            response = self.session.post(
                f"{API_BASE_URL}/products",
                json=product_data
            )
            if response.status_code in [200, 201]:
                self.imported_slugs.add(product_data["slug"])
                return True
            else:
                if "duplicate" in response.text.lower() or "exists" in response.text.lower():
                    self.stats["products_duplicate"] += 1
                    return False
                self.stats["errors"].append(f"API error: {response.text[:100]}")
                return False
        except Exception as e:
            self.stats["errors"].append(f"Import error: {e}")
            return False

    def import_sephora_products(self, csv_path: Path, limit: int = None) -> None:
        """Import products from Sephora CSV"""
        print(f"\n[SEPHORA] Importing products from {csv_path.name}...")

        try:
            df = pd.read_csv(csv_path, encoding='utf-8', on_bad_lines='skip')
            print(f"[SEPHORA] Found {len(df)} products in CSV")
            print(f"[SEPHORA] Columns: {list(df.columns)[:10]}...")

            if limit:
                df = df.head(limit)
                print(f"[SEPHORA] Limiting to {limit} products")

            imported = 0
            for idx, row in df.iterrows():
                product_data = self.transform_sephora_product(row)
                if product_data:
                    if self.import_product(product_data):
                        imported += 1
                        self.stats["products_imported"] += 1
                        if imported % 100 == 0:
                            print(f"[SEPHORA] Progress: {imported} products imported...")
                    else:
                        self.stats["products_skipped"] += 1
                else:
                    self.stats["products_skipped"] += 1

                # Rate limiting
                if idx % 50 == 0:
                    time.sleep(0.1)

            print(f"[SEPHORA] [OK] Imported {imported} products")

        except Exception as e:
            print(f"[SEPHORA] [X] Error: {e}")
            self.stats["errors"].append(f"Sephora import: {e}")

    def import_sephora_reviews(self, csv_path: Path, limit: int = None) -> None:
        """Import Sephora reviews as user events"""
        print(f"\n[SEPHORA REVIEWS] Importing reviews from {csv_path.name}...")

        try:
            df = pd.read_csv(csv_path, encoding='utf-8', on_bad_lines='skip',
                           nrows=limit if limit else 50000)

            print(f"[SEPHORA REVIEWS] Found {len(df)} reviews")
            print(f"[SEPHORA REVIEWS] Columns: {list(df.columns)}")

            # Get existing product IDs from database
            product_ids = self.get_product_ids()
            if not product_ids:
                print("[SEPHORA REVIEWS] [X] No products found in database!")
                return

            print(f"[SEPHORA REVIEWS] Found {len(product_ids)} products in database")

            # Map author IDs to virtual user IDs
            if 'author_id' in df.columns:
                unique_authors = df['author_id'].unique()
                author_mapping = {}
                for i, author_id in enumerate(unique_authors[:500]):
                    author_mapping[author_id] = 2000 + (i % 500)  # Use 2000+ range for Sephora users
            else:
                author_mapping = None

            events_imported = 0

            for idx, row in df.iterrows():
                # Get rating from review
                rating = 4  # Default
                if 'rating' in df.columns:
                    try:
                        rating = int(float(row.get('rating', 4)))
                    except:
                        rating = 4

                # Map to a product (use product_id if available, else hash)
                if 'product_id' in df.columns:
                    ext_product_id = str(row.get('product_id', ''))
                    product_idx = hash(ext_product_id) % len(product_ids)
                else:
                    product_idx = idx % len(product_ids)

                product_id = product_ids[product_idx]

                # Get virtual user ID
                if author_mapping and 'author_id' in df.columns:
                    author_id = row.get('author_id')
                    virtual_user_id = author_mapping.get(author_id, 2000)
                else:
                    virtual_user_id = 2000 + (idx % 500)

                # Set session ID
                self.session.headers["X-Session-ID"] = f"sephora-user-{virtual_user_id}"

                # Create review/rating event with user_id
                event_data = {
                    "event_type": "rating",
                    "product_id": product_id,
                    "rating": min(5, max(1, rating)),
                    "user_id": virtual_user_id
                }

                if self.import_event(event_data):
                    events_imported += 1

                    # Create product_view event with user_id
                    view_event = {"event_type": "product_view", "product_id": product_id, "user_id": virtual_user_id}
                    self.import_event(view_event)

                    # High ratings (4-5) -> purchase
                    if rating >= 4:
                        purchase_event = {"event_type": "purchase", "product_id": product_id, "quantity": 1, "user_id": virtual_user_id}
                        self.import_event(purchase_event)

                if events_imported % 1000 == 0 and events_imported > 0:
                    print(f"[SEPHORA REVIEWS] Progress: {events_imported} events imported...")

                if idx % 100 == 0:
                    time.sleep(0.05)

            self.stats["reviews_imported"] += events_imported
            print(f"[SEPHORA REVIEWS] [OK] Imported {events_imported} review events")

        except Exception as e:
            print(f"[SEPHORA REVIEWS] [X] Error: {e}")
            import traceback
            traceback.print_exc()
            self.stats["errors"].append(f"Sephora reviews: {e}")

    def import_cosmetics_products(self, csv_path: Path, limit: int = None) -> None:
        """Import products from Cosmetics CSV"""
        print(f"\n[COSMETICS] Importing products from {csv_path.name}...")

        try:
            df = pd.read_csv(csv_path, encoding='utf-8', on_bad_lines='skip')
            print(f"[COSMETICS] Found {len(df)} products in CSV")
            print(f"[COSMETICS] Columns: {list(df.columns)}")

            if limit:
                df = df.head(limit)
                print(f"[COSMETICS] Limiting to {limit} products")

            imported = 0
            for idx, row in df.iterrows():
                product_data = self.transform_cosmetics_product(row)
                if product_data:
                    if self.import_product(product_data):
                        imported += 1
                        self.stats["products_imported"] += 1
                        if imported % 100 == 0:
                            print(f"[COSMETICS] Progress: {imported} products imported...")
                    else:
                        self.stats["products_skipped"] += 1
                else:
                    self.stats["products_skipped"] += 1

                if idx % 50 == 0:
                    time.sleep(0.1)

            print(f"[COSMETICS] [OK] Imported {imported} products")

        except Exception as e:
            print(f"[COSMETICS] [X] Error: {e}")
            self.stats["errors"].append(f"Cosmetics import: {e}")

    def get_product_ids(self) -> List[int]:
        """Fetch all product IDs from the database"""
        try:
            response = self.session.get(f"{API_BASE_URL}/products?limit=10000")
            if response.status_code == 200:
                data = response.json()
                products = data.get("data", data) if isinstance(data, dict) else data
                if isinstance(products, dict) and "products" in products:
                    products = products["products"]
                return [p["id"] for p in products if isinstance(p, dict) and "id" in p]
        except Exception as e:
            print(f"[PRODUCTS] Error fetching product IDs: {e}")
        return []

    def create_virtual_users(self, count: int = 100) -> List[int]:
        """Create virtual user IDs for event simulation"""
        # For simplicity, we'll use user IDs 1000-1099 as virtual users
        # In production, you'd create actual user accounts
        return list(range(1000, 1000 + count))

    def import_event(self, event_data: Dict) -> bool:
        """Import a single event to the database via API"""
        try:
            response = self.session.post(
                f"{API_BASE_URL}/events",
                json=event_data
            )
            return response.status_code in [200, 201]
        except Exception as e:
            return False

    def import_amazon_reviews(self, csv_path: Path, limit: int = None) -> None:
        """Import Amazon reviews as user events (for recommendation training)"""
        print(f"\n[AMAZON] Processing reviews from {csv_path.name}...")
        print("[AMAZON] Importing ratings as user events for recommendations...")

        try:
            # Amazon ratings format: UserId, ProductId, Rating, Timestamp
            df = pd.read_csv(csv_path, encoding='utf-8', on_bad_lines='skip',
                           names=['user_id', 'product_id', 'rating', 'timestamp'],
                           nrows=limit if limit else 100000)  # Limit for memory

            print(f"[AMAZON] Found {len(df)} ratings")
            print(f"[AMAZON] Unique users: {df['user_id'].nunique()}")
            print(f"[AMAZON] Unique products: {df['product_id'].nunique()}")
            print(f"[AMAZON] Rating distribution:")
            print(df['rating'].value_counts().sort_index())

            # Get existing product IDs from database
            product_ids = self.get_product_ids()
            if not product_ids:
                print("[AMAZON] [X] No products found in database. Import products first!")
                return

            print(f"[AMAZON] Found {len(product_ids)} products in database")

            # Create user ID mapping (hash external user IDs to virtual user IDs)
            unique_users = df['user_id'].unique()
            user_mapping = {}
            for i, ext_user_id in enumerate(unique_users[:500]):  # Limit to 500 users
                # Map external user ID to virtual user ID (1000+)
                user_mapping[ext_user_id] = 1000 + (i % 500)

            print(f"[AMAZON] Mapped {len(user_mapping)} unique users")

            # Import events
            events_imported = 0
            events_limit = min(len(df), 50000)  # Limit events to import

            # Sample events if too many
            if len(df) > events_limit:
                df = df.sample(n=events_limit, random_state=42)
                print(f"[AMAZON] Sampled {events_limit} events for import")

            for idx, row in df.iterrows():
                ext_user_id = row['user_id']
                rating = row['rating']

                # Skip if user not in mapping
                if ext_user_id not in user_mapping:
                    continue

                # Map to internal user ID and random product
                internal_user_id = user_mapping[ext_user_id]
                # Use hash to consistently map to a product
                product_idx = hash(f"{ext_user_id}-{row['product_id']}") % len(product_ids)
                product_id = product_ids[product_idx]

                # Create rating event - skip invalid ratings
                try:
                    rating_int = int(float(rating))
                except (ValueError, TypeError):
                    continue  # Skip rows with invalid rating values

                event_data = {
                    "event_type": "rating",
                    "product_id": product_id,
                    "rating": rating_int,
                    "user_id": internal_user_id
                }

                # Set session ID based on user for consistency
                self.session.headers["X-Session-ID"] = f"amazon-user-{internal_user_id}"

                if self.import_event(event_data):
                    events_imported += 1

                    # Also create a product_view event (users typically view before rating)
                    view_event = {
                        "event_type": "product_view",
                        "product_id": product_id,
                        "user_id": internal_user_id
                    }
                    self.import_event(view_event)

                    # For high ratings (4-5), also create purchase event
                    if rating_int >= 4:
                        purchase_event = {
                            "event_type": "purchase",
                            "product_id": product_id,
                            "quantity": 1,
                            "user_id": internal_user_id
                        }
                        self.import_event(purchase_event)

                if events_imported % 1000 == 0 and events_imported > 0:
                    print(f"[AMAZON] Progress: {events_imported} events imported...")

                # Rate limiting
                if idx % 100 == 0:
                    time.sleep(0.05)

            self.stats["reviews_imported"] = events_imported
            print(f"[AMAZON] [OK] Imported {events_imported} rating events")
            print(f"[AMAZON] [OK] Also created view/purchase events for collaborative filtering")

        except Exception as e:
            print(f"[AMAZON] [X] Error: {e}")
            import traceback
            traceback.print_exc()
            self.stats["errors"].append(f"Amazon reviews: {e}")

    def run(self, limit: int = None, events_only: bool = False):
        """Main execution flow"""
        print("=" * 60)
        print("FLORUS BEAUTY - KAGGLE DATA IMPORTER")
        print("=" * 60)
        print(f"Time: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        print(f"Data directory: {DATA_DIR}")
        if events_only:
            print("[MODE] Events only - skipping product import")

        # Step 1: Check Kaggle setup
        if not self.check_kaggle_setup():
            print("\n[ERROR] Kaggle API not configured. Please set up credentials.")
            sys.exit(1)

        # Step 2: Login to API
        if not self.login():
            print("\n[ERROR] Failed to authenticate with Florus API")
            print("        Make sure backend is running: go run ./cmd/api")
            sys.exit(1)

        # Step 3: Set up categories
        if not self.ensure_categories():
            print("\n[ERROR] Failed to set up categories")
            sys.exit(1)

        # Step 4: Download datasets
        print("\n" + "=" * 60)
        print("DOWNLOADING KAGGLE DATASETS")
        print("=" * 60)
        files = self.download_datasets()

        if not files:
            print("\n[ERROR] No datasets downloaded. Check Kaggle credentials and internet connection.")
            sys.exit(1)

        print(f"\n[INFO] Downloaded files: {list(files.keys())}")

        # Step 5: Import products (skip if events_only)
        if not events_only:
            print("\n" + "=" * 60)
            print("IMPORTING PRODUCTS")
            print("=" * 60)

            # Import Sephora products first (primary dataset)
            if "sephora_products" in files:
                self.import_sephora_products(files["sephora_products"], limit=limit)

            # Import Cosmetics products
            if "cosmetics_products" in files:
                self.import_cosmetics_products(files["cosmetics_products"], limit=limit)

        # Step 6: Import user events from reviews/ratings
        print("\n" + "=" * 60)
        print("IMPORTING USER EVENTS FROM REVIEWS/RATINGS")
        print("=" * 60)

        # Import Sephora reviews as events
        if "sephora_reviews" in files:
            self.import_sephora_reviews(files["sephora_reviews"], limit=limit)
        else:
            # Try to find any review files
            for f in DATA_DIR.glob("*review*.csv"):
                if "sephora" in f.name.lower() or "0-250" in f.name:
                    self.import_sephora_reviews(f, limit=limit)
                    break

        # Import Amazon ratings as events
        if "amazon_reviews" in files:
            self.import_amazon_reviews(files["amazon_reviews"], limit=limit)

        # Summary
        print("\n" + "=" * 60)
        print("IMPORT SUMMARY")
        print("=" * 60)
        print(f"Categories created: {self.stats['categories_created']}")
        print(f"Products imported:  {self.stats['products_imported']}")
        print(f"Products skipped:   {self.stats['products_skipped']}")
        print(f"Duplicates:         {self.stats['products_duplicate']}")
        print(f"Reviews processed:  {self.stats['reviews_imported']}")
        print(f"Errors:             {len(self.stats['errors'])}")

        if self.stats["errors"]:
            print(f"\nFirst 5 errors:")
            for err in self.stats["errors"][:5]:
                print(f"  - {err}")

        print("\n" + "=" * 60)
        print("NEXT STEPS")
        print("=" * 60)
        print("1. Generate user events: python generate_user_events.py")
        print("2. Access API endpoints:")
        print("   - Products: http://localhost:8081/api/products")
        print("   - Recommendations: http://localhost:8081/api/recommendations/cold-start")
        print("=" * 60)


def main():
    import argparse
    parser = argparse.ArgumentParser(description="Import Kaggle beauty data into Florus")
    parser.add_argument("--limit", type=int, help="Limit number of products/events per dataset")
    parser.add_argument("--events-only", action="store_true", help="Only import events, skip products")
    args = parser.parse_args()

    importer = KaggleDataImporter()
    importer.run(limit=args.limit, events_only=args.events_only)


if __name__ == "__main__":
    main()
