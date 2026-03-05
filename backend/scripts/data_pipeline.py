#!/usr/bin/env python3
"""
Florus Beauty - Complete Data Pipeline
========================================
Comprehensive ETL pipeline for both Web Database and Data Lake.

Architecture:
┌─────────────────────────────────────────────────────────────────────┐
│                        DATA SOURCES (Kaggle)                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────────┐  │
│  │   Sephora    │  │  Cosmetics   │  │     Amazon Ratings       │  │
│  │  Products &  │  │   Dataset    │  │    (2M+ ratings)         │  │
│  │   Reviews    │  │  (1.4K+)     │  │                          │  │
│  └──────┬───────┘  └──────┬───────┘  └───────────┬──────────────┘  │
└─────────┼─────────────────┼──────────────────────┼──────────────────┘
          │                 │                      │
          ▼                 ▼                      ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      EXTRACTION LAYER                               │
│                   (download_kaggle_data)                            │
└─────────────────────────────┬───────────────────────────────────────┘
                              │
          ┌───────────────────┴───────────────────┐
          │                                       │
          ▼                                       ▼
┌───────────────────────┐             ┌───────────────────────────────┐
│    WEB DATABASE       │             │         DATA LAKE             │
│    (Oracle/SQLite)    │             │                               │
│  ┌─────────────────┐  │             │  ┌─────────────────────────┐  │
│  │   Products      │  │             │  │  RAW LAYER              │  │
│  │   Categories    │  │             │  │  - products_raw.parquet │  │
│  │   Users         │  │             │  │  - reviews_raw.parquet  │  │
│  │   Events        │  │             │  │  - ratings_raw.parquet  │  │
│  │   Orders        │  │             │  └───────────┬─────────────┘  │
│  └─────────────────┘  │             │              │                │
│                       │             │              ▼                │
│                       │             │  ┌─────────────────────────┐  │
│                       │             │  │  PROCESSED LAYER        │  │
│                       │             │  │  - user_product_matrix  │  │
│                       │             │  │  - co_occurrence_matrix │  │
│                       │             │  │  - product_features     │  │
│                       │             │  └───────────┬─────────────┘  │
│                       │             │              │                │
│                       │             │              ▼                │
│                       │             │  ┌─────────────────────────┐  │
│                       │             │  │  FEATURES LAYER         │  │
│                       │             │  │  - user_embeddings      │  │
│                       │             │  │  - product_embeddings   │  │
│                       │             │  │  - similarity_scores    │  │
│                       │             │  └─────────────────────────┘  │
└───────────────────────┘             └───────────────────────────────┘

Usage:
    python data_pipeline.py --mode full          # Run complete pipeline
    python data_pipeline.py --mode web-only      # Only load to web database
    python data_pipeline.py --mode lake-only     # Only build data lake
    python data_pipeline.py --mode features      # Only generate features
    python data_pipeline.py --mode minio-sync    # Sync data lake to MinIO
    python data_pipeline.py --mode full --sync-minio  # Full pipeline + MinIO sync

MinIO Storage:
    Data lake files are synced to MinIO for distributed ML training.
    Bucket: florus-datalake
    Structure:
      - datalake/raw/         # Raw parquet files
      - datalake/processed/   # Processed matrices
      - datalake/features/    # ML features
"""

import os
import sys
import json
import time
import hashlib
import requests
import numpy as np
import pandas as pd
from pathlib import Path
from typing import Dict, List, Optional, Tuple
from datetime import datetime
from collections import defaultdict
from io import BytesIO
import warnings
warnings.filterwarnings('ignore')

# MinIO client
try:
    from minio import Minio
    from minio.error import S3Error
    MINIO_AVAILABLE = True
except ImportError:
    MINIO_AVAILABLE = False
    print("[WARNING] minio package not installed. Run: pip install minio")

# Configuration
API_BASE_URL = "http://localhost:8081/api"
ADMIN_EMAIL = "admin@florus.com"
ADMIN_PASSWORD = "admin123"

# MinIO Configuration
MINIO_ENDPOINT = "localhost:9000"
MINIO_ACCESS_KEY = "minioadmin"
MINIO_SECRET_KEY = "minioadmin123"
MINIO_BUCKET = "florus-datalake"
MINIO_USE_SSL = False

# Directories (local cache)
BASE_DIR = Path(__file__).parent.parent
DATA_DIR = BASE_DIR / "data"
LAKE_DIR = DATA_DIR / "lake"
RAW_LAYER = LAKE_DIR / "raw"
PROCESSED_LAYER = LAKE_DIR / "processed"
FEATURES_LAYER = LAKE_DIR / "features"

# Create directories
for d in [DATA_DIR, LAKE_DIR, RAW_LAYER, PROCESSED_LAYER, FEATURES_LAYER]:
    d.mkdir(parents=True, exist_ok=True)


class DataPipeline:
    """Complete data pipeline for Florus Beauty recommendation system"""

    def __init__(self):
        self.session = requests.Session()
        self.token = None
        self.categories = {}
        self.products_df = None
        self.reviews_df = None
        self.ratings_df = None
        self.minio_client = None
        self.stats = {
            "products_loaded": 0,
            "events_loaded": 0,
            "features_generated": 0,
            "files_uploaded": 0,
            "errors": []
        }

    # ============================================================
    # MINIO INTEGRATION
    # ============================================================

    def init_minio(self) -> bool:
        """Initialize MinIO client and create bucket"""
        if not MINIO_AVAILABLE:
            print("[MINIO] [X] minio package not installed")
            return False

        try:
            self.minio_client = Minio(
                MINIO_ENDPOINT,
                access_key=MINIO_ACCESS_KEY,
                secret_key=MINIO_SECRET_KEY,
                secure=MINIO_USE_SSL
            )

            # Create bucket if not exists
            if not self.minio_client.bucket_exists(MINIO_BUCKET):
                self.minio_client.make_bucket(MINIO_BUCKET)
                print(f"[MINIO] [OK] Created bucket: {MINIO_BUCKET}")
            else:
                print(f"[MINIO] [OK] Bucket exists: {MINIO_BUCKET}")

            return True
        except Exception as e:
            print(f"[MINIO] [X] Failed to connect: {e}")
            return False

    def upload_to_minio(self, local_path: Path, minio_path: str) -> bool:
        """Upload a file to MinIO"""
        if not self.minio_client:
            return False

        try:
            self.minio_client.fput_object(
                MINIO_BUCKET,
                minio_path,
                str(local_path),
                content_type="application/octet-stream"
            )
            return True
        except Exception as e:
            print(f"[MINIO] [X] Upload failed {local_path.name}: {e}")
            return False

    def sync_to_minio(self) -> int:
        """Sync all data lake files to MinIO"""
        print("\n" + "=" * 60)
        print("MINIO - Syncing Data Lake")
        print("=" * 60)

        if not self.init_minio():
            print("[MINIO] [X] Skipping MinIO sync - not available")
            return 0

        uploaded = 0
        layers = [
            ("raw", RAW_LAYER),
            ("processed", PROCESSED_LAYER),
            ("features", FEATURES_LAYER)
        ]

        for layer_name, layer_dir in layers:
            print(f"\n[MINIO] Uploading {layer_name} layer...")
            for f in layer_dir.glob("*.parquet"):
                minio_path = f"datalake/{layer_name}/{f.name}"
                if self.upload_to_minio(f, minio_path):
                    size_mb = f.stat().st_size / (1024 * 1024)
                    print(f"[MINIO] [OK] {minio_path} ({size_mb:.2f} MB)")
                    uploaded += 1

        self.stats["files_uploaded"] = uploaded
        print(f"\n[MINIO] [OK] Uploaded {uploaded} files to MinIO")

        # Print usage instructions
        print(f"\n[MINIO] Access files via:")
        print(f"  - S3 URL: s3://{MINIO_BUCKET}/datalake/")
        print(f"  - HTTP: http://{MINIO_ENDPOINT}/{MINIO_BUCKET}/datalake/")

        return uploaded

    def list_minio_files(self) -> List[str]:
        """List all files in MinIO bucket"""
        if not self.init_minio():
            return []

        files = []
        try:
            objects = self.minio_client.list_objects(MINIO_BUCKET, prefix="datalake/", recursive=True)
            for obj in objects:
                files.append(obj.object_name)
        except Exception as e:
            print(f"[MINIO] [X] List failed: {e}")

        return files

    # ============================================================
    # EXTRACTION LAYER
    # ============================================================

    def extract_kaggle_data(self) -> Dict[str, Path]:
        """Download and extract all Kaggle datasets"""
        print("\n" + "=" * 60)
        print("EXTRACTION LAYER - Downloading Kaggle Datasets")
        print("=" * 60)

        try:
            import kaggle
            kaggle.api.authenticate()
        except Exception as e:
            print(f"[ERROR] Kaggle not configured: {e}")
            return {}

        datasets = {
            "sephora": "nadyinky/sephora-products-and-skincare-reviews",
            "cosmetics": "kingabzpro/cosmetics-datasets",
            "amazon": "skillsmuggler/amazon-ratings"
        }

        downloaded = {}
        for name, kaggle_id in datasets.items():
            print(f"\n[EXTRACT] Downloading {name}...")
            try:
                kaggle.api.dataset_download_files(
                    kaggle_id,
                    path=str(DATA_DIR),
                    unzip=True
                )
                print(f"[EXTRACT] [OK] {name} downloaded")
                downloaded[name] = DATA_DIR
            except Exception as e:
                print(f"[EXTRACT] [X] {name} failed: {e}")

        return downloaded

    # ============================================================
    # DATA LAKE - RAW LAYER
    # ============================================================

    def load_raw_layer(self) -> bool:
        """Load raw data into Data Lake raw layer as Parquet files"""
        print("\n" + "=" * 60)
        print("DATA LAKE - RAW LAYER")
        print("=" * 60)

        # Load Sephora Products
        sephora_products = DATA_DIR / "product_info.csv"
        if sephora_products.exists():
            print(f"[RAW] Loading Sephora products...")
            df = pd.read_csv(sephora_products, on_bad_lines='skip')
            df['source'] = 'sephora'
            df['loaded_at'] = datetime.now().isoformat()
            df.to_parquet(RAW_LAYER / "sephora_products_raw.parquet", index=False)
            print(f"[RAW] [OK] Saved {len(df)} Sephora products")
            self.products_df = df

        # Load Cosmetics Products
        cosmetics_file = DATA_DIR / "cosmetics.csv"
        if cosmetics_file.exists():
            print(f"[RAW] Loading Cosmetics products...")
            df = pd.read_csv(cosmetics_file, on_bad_lines='skip')
            df['source'] = 'cosmetics'
            df['loaded_at'] = datetime.now().isoformat()
            df.to_parquet(RAW_LAYER / "cosmetics_products_raw.parquet", index=False)
            print(f"[RAW] [OK] Saved {len(df)} Cosmetics products")

        # Load Sephora Reviews
        review_files = list(DATA_DIR.glob("reviews*.csv"))
        if review_files:
            print(f"[RAW] Loading Sephora reviews...")
            dfs = []
            for f in review_files[:3]:  # Limit to first 3 files
                df = pd.read_csv(f, on_bad_lines='skip', nrows=100000)
                dfs.append(df)
            reviews_df = pd.concat(dfs, ignore_index=True)
            reviews_df['source'] = 'sephora'
            reviews_df['loaded_at'] = datetime.now().isoformat()
            # Convert problematic columns to string to avoid type conflicts
            for col in reviews_df.columns:
                if reviews_df[col].dtype == 'object':
                    reviews_df[col] = reviews_df[col].astype(str)
            reviews_df.to_parquet(RAW_LAYER / "sephora_reviews_raw.parquet", index=False)
            print(f"[RAW] [OK] Saved {len(reviews_df)} Sephora reviews")
            self.reviews_df = reviews_df

        # Load Amazon Ratings
        amazon_file = DATA_DIR / "ratings_Beauty.csv"
        if amazon_file.exists():
            print(f"[RAW] Loading Amazon ratings...")
            df = pd.read_csv(
                amazon_file,
                names=['user_id', 'product_id', 'rating', 'timestamp'],
                on_bad_lines='skip',
                nrows=500000  # Limit for memory
            )
            df = df[df['rating'].apply(lambda x: str(x).replace('.', '').isdigit())]
            df['rating'] = df['rating'].astype(float)
            df['user_id'] = df['user_id'].astype(str)
            df['product_id'] = df['product_id'].astype(str)
            df['timestamp'] = df['timestamp'].astype(str)
            df['source'] = 'amazon'
            df['loaded_at'] = datetime.now().isoformat()
            df.to_parquet(RAW_LAYER / "amazon_ratings_raw.parquet", index=False)
            print(f"[RAW] [OK] Saved {len(df)} Amazon ratings")
            self.ratings_df = df

        print(f"\n[RAW] Raw layer files:")
        for f in RAW_LAYER.glob("*.parquet"):
            size_mb = f.stat().st_size / (1024 * 1024)
            print(f"  - {f.name} ({size_mb:.2f} MB)")

        return True

    # ============================================================
    # DATA LAKE - PROCESSED LAYER
    # ============================================================

    def build_processed_layer(self) -> bool:
        """Build processed data for recommendation algorithms"""
        print("\n" + "=" * 60)
        print("DATA LAKE - PROCESSED LAYER")
        print("=" * 60)

        # Load raw data if not in memory
        if self.ratings_df is None:
            ratings_file = RAW_LAYER / "amazon_ratings_raw.parquet"
            if ratings_file.exists():
                self.ratings_df = pd.read_parquet(ratings_file)

        if self.reviews_df is None:
            reviews_file = RAW_LAYER / "sephora_reviews_raw.parquet"
            if reviews_file.exists():
                self.reviews_df = pd.read_parquet(reviews_file)

        # 1. Build User-Product Interaction Matrix
        print("\n[PROCESSED] Building user-product interaction matrix...")
        if self.ratings_df is not None:
            interactions = self._build_interaction_matrix(self.ratings_df)
            interactions.to_parquet(PROCESSED_LAYER / "user_product_interactions.parquet", index=False)
            print(f"[PROCESSED] [OK] User-product matrix: {len(interactions)} interactions")

        # 2. Build Product Co-occurrence Matrix
        print("\n[PROCESSED] Building product co-occurrence matrix...")
        if self.ratings_df is not None:
            co_occurrence = self._build_co_occurrence_matrix(self.ratings_df)
            co_occurrence.to_parquet(PROCESSED_LAYER / "product_co_occurrence.parquet", index=False)
            print(f"[PROCESSED] [OK] Co-occurrence matrix: {len(co_occurrence)} pairs")

        # 3. Build Product Features
        print("\n[PROCESSED] Building product features...")
        if self.products_df is not None:
            product_features = self._build_product_features(self.products_df)
            product_features.to_parquet(PROCESSED_LAYER / "product_features.parquet", index=False)
            print(f"[PROCESSED] [OK] Product features: {len(product_features)} products")

        # 4. Build User Behavior Profiles
        print("\n[PROCESSED] Building user behavior profiles...")
        if self.ratings_df is not None:
            user_profiles = self._build_user_profiles(self.ratings_df)
            user_profiles.to_parquet(PROCESSED_LAYER / "user_profiles.parquet", index=False)
            print(f"[PROCESSED] [OK] User profiles: {len(user_profiles)} users")

        print(f"\n[PROCESSED] Processed layer files:")
        for f in PROCESSED_LAYER.glob("*.parquet"):
            size_mb = f.stat().st_size / (1024 * 1024)
            print(f"  - {f.name} ({size_mb:.2f} MB)")

        return True

    def _build_interaction_matrix(self, df: pd.DataFrame) -> pd.DataFrame:
        """Build user-product interaction matrix"""
        interactions = df.groupby(['user_id', 'product_id']).agg({
            'rating': ['mean', 'count']
        }).reset_index()
        interactions.columns = ['user_id', 'product_id', 'avg_rating', 'interaction_count']
        interactions['interaction_score'] = (
            interactions['avg_rating'] * 0.7 +
            np.log1p(interactions['interaction_count']) * 0.3
        )
        return interactions

    def _build_co_occurrence_matrix(self, df: pd.DataFrame, min_support: int = 2) -> pd.DataFrame:
        """Build product co-occurrence matrix (products viewed/rated by same users)"""
        # Get products per user
        user_products = df.groupby('user_id')['product_id'].apply(set).to_dict()

        # Count co-occurrences
        co_occurrence = defaultdict(int)
        for user_id, products in user_products.items():
            products_list = list(products)
            for i, p1 in enumerate(products_list):
                for p2 in products_list[i+1:]:
                    pair = tuple(sorted([str(p1), str(p2)]))
                    co_occurrence[pair] += 1

        # Convert to DataFrame
        records = [
            {'product_a': pair[0], 'product_b': pair[1], 'co_count': count}
            for pair, count in co_occurrence.items()
            if count >= min_support
        ]

        return pd.DataFrame(records)

    def _build_product_features(self, df: pd.DataFrame) -> pd.DataFrame:
        """Build product feature vectors for content-based filtering"""
        features = []

        for idx, row in df.iterrows():
            # Handle NaN values safely
            reviews_val = row.get('reviews', 0)
            if pd.isna(reviews_val):
                reviews_val = 0
            rating_val = row.get('rating', row.get('Rank', 4))
            if pd.isna(rating_val):
                rating_val = 4
            price_val = row.get('price_usd', row.get('Price', 0))
            if pd.isna(price_val):
                price_val = 0

            feature = {
                'product_id': row.get('product_id', idx),
                'name': str(row.get('product_name', row.get('Name', ''))),
                'brand': str(row.get('brand_name', row.get('Brand', 'Unknown'))),
                'category': str(row.get('primary_category', row.get('Label', 'Other'))),
                'price': float(price_val),
                'rating': float(rating_val),
                'reviews_count': int(reviews_val),
            }
            features.append(feature)

        return pd.DataFrame(features)

    def _build_user_profiles(self, df: pd.DataFrame) -> pd.DataFrame:
        """Build user behavior profiles for collaborative filtering"""
        user_stats = df.groupby('user_id').agg({
            'rating': ['mean', 'std', 'count'],
            'product_id': 'nunique'
        }).reset_index()

        user_stats.columns = [
            'user_id', 'avg_rating', 'rating_std', 'total_ratings', 'unique_products'
        ]
        user_stats['rating_std'] = user_stats['rating_std'].fillna(0)

        return user_stats

    # ============================================================
    # DATA LAKE - FEATURES LAYER (for ML models)
    # ============================================================

    def build_features_layer(self) -> bool:
        """Build feature vectors for recommendation models"""
        print("\n" + "=" * 60)
        print("DATA LAKE - FEATURES LAYER")
        print("=" * 60)

        # Load processed data
        interactions_file = PROCESSED_LAYER / "user_product_interactions.parquet"
        co_occurrence_file = PROCESSED_LAYER / "product_co_occurrence.parquet"
        product_features_file = PROCESSED_LAYER / "product_features.parquet"

        # 1. Build Product Similarity Scores
        print("\n[FEATURES] Computing product similarity scores...")
        if co_occurrence_file.exists():
            co_df = pd.read_parquet(co_occurrence_file)
            similarity_df = self._compute_similarity_scores(co_df)
            similarity_df.to_parquet(FEATURES_LAYER / "product_similarity.parquet", index=False)
            print(f"[FEATURES] [OK] Product similarity: {len(similarity_df)} pairs")

        # 2. Build Collaborative Filtering Features
        print("\n[FEATURES] Building collaborative filtering features...")
        if interactions_file.exists():
            interactions_df = pd.read_parquet(interactions_file)
            cf_features = self._build_cf_features(interactions_df)
            cf_features.to_parquet(FEATURES_LAYER / "cf_features.parquet", index=False)
            print(f"[FEATURES] [OK] CF features: {len(cf_features)} entries")

        # 3. Build Content-Based Features
        print("\n[FEATURES] Building content-based features...")
        if product_features_file.exists():
            products_df = pd.read_parquet(product_features_file)
            cb_features = self._build_cb_features(products_df)
            cb_features.to_parquet(FEATURES_LAYER / "cb_features.parquet", index=False)
            print(f"[FEATURES] [OK] CB features: {len(cb_features)} products")

        # 4. Build Hybrid Recommendation Scores
        print("\n[FEATURES] Computing hybrid recommendation scores...")
        hybrid_scores = self._build_hybrid_scores()
        if hybrid_scores is not None:
            hybrid_scores.to_parquet(FEATURES_LAYER / "hybrid_scores.parquet", index=False)
            print(f"[FEATURES] [OK] Hybrid scores: {len(hybrid_scores)} product pairs")

        print(f"\n[FEATURES] Features layer files:")
        for f in FEATURES_LAYER.glob("*.parquet"):
            size_mb = f.stat().st_size / (1024 * 1024)
            print(f"  - {f.name} ({size_mb:.2f} MB)")

        self.stats["features_generated"] = len(list(FEATURES_LAYER.glob("*.parquet")))
        return True

    def _compute_similarity_scores(self, co_df: pd.DataFrame) -> pd.DataFrame:
        """Compute Jaccard similarity between products"""
        # Normalize co-occurrence counts to similarity scores
        max_count = co_df['co_count'].max()
        co_df['similarity'] = co_df['co_count'] / max_count
        return co_df[['product_a', 'product_b', 'similarity', 'co_count']]

    def _build_cf_features(self, interactions_df: pd.DataFrame) -> pd.DataFrame:
        """Build collaborative filtering features"""
        # Aggregate user preferences
        user_prefs = interactions_df.groupby('user_id').agg({
            'interaction_score': ['mean', 'std', 'sum'],
            'product_id': 'count'
        }).reset_index()
        user_prefs.columns = ['user_id', 'avg_score', 'score_std', 'total_score', 'product_count']
        user_prefs['score_std'] = user_prefs['score_std'].fillna(0)
        return user_prefs

    def _build_cb_features(self, products_df: pd.DataFrame) -> pd.DataFrame:
        """Build content-based features with TF-IDF-like encoding"""
        # One-hot encode categories and brands
        products_df['category_encoded'] = pd.Categorical(products_df['category']).codes
        products_df['brand_encoded'] = pd.Categorical(products_df['brand']).codes

        # Normalize numerical features
        for col in ['price', 'rating', 'reviews_count']:
            if col in products_df.columns:
                max_val = products_df[col].max()
                if max_val > 0:
                    products_df[f'{col}_normalized'] = products_df[col] / max_val

        return products_df

    def _build_hybrid_scores(self) -> Optional[pd.DataFrame]:
        """Combine CF and CB features into hybrid scores"""
        similarity_file = FEATURES_LAYER / "product_similarity.parquet"
        cb_file = FEATURES_LAYER / "cb_features.parquet"

        if not similarity_file.exists():
            return None

        similarity_df = pd.read_parquet(similarity_file)

        # Hybrid score = weighted combination
        # For now, just use collaborative similarity
        similarity_df['hybrid_score'] = similarity_df['similarity']

        return similarity_df[['product_a', 'product_b', 'hybrid_score']]

    # ============================================================
    # WEB DATABASE LOADING
    # ============================================================

    def login_api(self) -> bool:
        """Authenticate with backend API"""
        try:
            response = self.session.post(
                f"{API_BASE_URL}/auth/login",
                json={"email": ADMIN_EMAIL, "password": ADMIN_PASSWORD}
            )
            if response.status_code == 200:
                data = response.json().get("data", response.json())
                self.token = data.get("token")
                if self.token:
                    self.session.headers["Authorization"] = f"Bearer {self.token}"
                    return True
            return False
        except Exception as e:
            print(f"[API] Login error: {e}")
            return False

    def load_to_web_database(self, limit: int = None) -> bool:
        """Load transformed data to web database via API"""
        print("\n" + "=" * 60)
        print("WEB DATABASE - Loading Data")
        print("=" * 60)

        if not self.login_api():
            print("[WEB DB] [X] Failed to authenticate with API")
            return False

        print("[WEB DB] [OK] Authenticated with API")

        # 1. Load Products
        print("\n[WEB DB] Loading products...")
        products_loaded = self._load_products_to_api(limit)
        print(f"[WEB DB] [OK] Loaded {products_loaded} products")

        # 2. Load Events from processed data
        print("\n[WEB DB] Loading user events...")
        events_loaded = self._load_events_to_api(limit)
        print(f"[WEB DB] [OK] Loaded {events_loaded} events")

        self.stats["products_loaded"] = products_loaded
        self.stats["events_loaded"] = events_loaded

        return True

    def _load_products_to_api(self, limit: int = None) -> int:
        """Load products to web database via API"""
        products_file = RAW_LAYER / "sephora_products_raw.parquet"
        if not products_file.exists():
            return 0

        df = pd.read_parquet(products_file)
        if limit:
            df = df.head(limit)

        loaded = 0
        for idx, row in df.iterrows():
            product_data = self._transform_product(row, idx)
            if product_data:
                try:
                    response = self.session.post(
                        f"{API_BASE_URL}/products",
                        json=product_data
                    )
                    if response.status_code in [200, 201]:
                        loaded += 1
                        if loaded % 100 == 0:
                            print(f"[WEB DB] Products progress: {loaded}...")
                except Exception as e:
                    pass

            if idx % 50 == 0:
                time.sleep(0.1)

        return loaded

    def _transform_product(self, row: pd.Series, idx: int) -> Optional[Dict]:
        """Transform raw product to API format"""
        name = str(row.get('product_name', '')).strip()
        brand = str(row.get('brand_name', 'Unknown')).strip()

        if not name or len(name) < 3:
            return None

        price = 0
        for col in ['price_usd', 'sale_price_usd', 'value_price_usd']:
            try:
                p = row.get(col, 0)
                if pd.isna(p):
                    continue
                p = float(p)
                if p > 0:
                    price = p * 24000  # Convert to VND
                    break
            except:
                pass

        if price <= 0:
            return None

        slug = self._generate_slug(name, brand)
        rating_val = row.get('rating', 4)
        if pd.isna(rating_val):
            rating_val = 4
        rating = min(5.0, max(0.0, float(rating_val)))

        reviews_val = row.get('reviews', 0)
        if pd.isna(reviews_val):
            reviews_val = 0
        reviews = int(reviews_val)

        return {
            "name": name[:255],
            "slug": slug,
            "brand": brand[:100],
            "price": price,
            "original_price": price * 1.2 if rating >= 4 else None,
            "category_id": 1,  # Default to first category
            "description": str(row.get('highlights', f'{brand} {name}'))[:2000],
            "image_url": "",
            "stock": 100,
            "rating": rating,
            "review_count": reviews,
            "is_active": True
        }

    def _generate_slug(self, name: str, brand: str) -> str:
        base = f"{brand}-{name}".lower()
        slug = "".join(c if c.isalnum() or c == " " else "" for c in base)
        slug = "-".join(slug.split())
        hash_suffix = hashlib.md5(f"{brand}{name}".encode()).hexdigest()[:6]
        return f"{slug[:80]}-{hash_suffix}"

    def _load_events_to_api(self, limit: int = None) -> int:
        """Load user events to web database"""
        # Get product IDs from database
        try:
            response = self.session.get(f"{API_BASE_URL}/products?limit=10000")
            products = response.json().get("data", {}).get("products", [])
            product_ids = [p["id"] for p in products if "id" in p]
        except:
            product_ids = list(range(1, 101))  # Fallback

        if not product_ids:
            return 0

        # Load ratings from processed layer
        interactions_file = PROCESSED_LAYER / "user_product_interactions.parquet"
        if not interactions_file.exists():
            return 0

        df = pd.read_parquet(interactions_file)
        if limit:
            df = df.head(limit)

        loaded = 0
        for idx, row in df.iterrows():
            # Map to internal product ID
            product_idx = hash(str(row['product_id'])) % len(product_ids)
            product_id = product_ids[product_idx]

            # Map to virtual user ID
            user_id = 1000 + (hash(str(row['user_id'])) % 500)

            # Create events
            rating = int(min(5, max(1, row['avg_rating'])))

            events = [
                {"event_type": "rating", "product_id": product_id, "rating": rating, "user_id": user_id},
                {"event_type": "product_view", "product_id": product_id, "user_id": user_id},
            ]

            if rating >= 4:
                events.append({"event_type": "purchase", "product_id": product_id, "quantity": 1, "user_id": user_id})

            for event in events:
                try:
                    self.session.headers["X-Session-ID"] = f"pipeline-user-{user_id}"
                    response = self.session.post(f"{API_BASE_URL}/events", json=event)
                    if response.status_code in [200, 201]:
                        loaded += 1
                except:
                    pass

            if idx % 100 == 0:
                time.sleep(0.05)
                if idx % 1000 == 0:
                    print(f"[WEB DB] Events progress: {loaded}...")

        return loaded

    # ============================================================
    # MAIN PIPELINE ORCHESTRATION
    # ============================================================

    def run(self, mode: str = "full", limit: int = None, sync_minio: bool = False):
        """Run the data pipeline"""
        print("=" * 70)
        print("FLORUS BEAUTY - DATA PIPELINE")
        print("=" * 70)
        print(f"Time: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        print(f"Mode: {mode}")
        print(f"Limit: {limit if limit else 'None (full data)'}")
        print(f"MinIO Sync: {'Yes' if sync_minio else 'No'}")
        print("=" * 70)

        start_time = time.time()

        if mode in ["full", "extract"]:
            self.extract_kaggle_data()

        if mode in ["full", "lake-only", "raw"]:
            self.load_raw_layer()

        if mode in ["full", "lake-only", "processed"]:
            self.build_processed_layer()

        if mode in ["full", "lake-only", "features"]:
            self.build_features_layer()

        if mode in ["full", "web-only"]:
            self.load_to_web_database(limit)

        # Sync to MinIO if requested or mode is minio-sync
        if sync_minio or mode == "minio-sync":
            self.sync_to_minio()

        elapsed = time.time() - start_time

        # Summary
        print("\n" + "=" * 70)
        print("PIPELINE SUMMARY")
        print("=" * 70)
        print(f"Duration: {elapsed:.2f} seconds")
        print(f"Products loaded to web DB: {self.stats['products_loaded']}")
        print(f"Events loaded to web DB: {self.stats['events_loaded']}")
        print(f"Features files generated: {self.stats['features_generated']}")
        print(f"Files uploaded to MinIO: {self.stats['files_uploaded']}")
        print(f"Errors: {len(self.stats['errors'])}")

        print("\n[DATA LAKE FILES]")
        for layer_name, layer_dir in [("Raw", RAW_LAYER), ("Processed", PROCESSED_LAYER), ("Features", FEATURES_LAYER)]:
            files = list(layer_dir.glob("*.parquet"))
            if files:
                print(f"\n  {layer_name} Layer:")
                for f in files:
                    size_mb = f.stat().st_size / (1024 * 1024)
                    print(f"    - {f.name} ({size_mb:.2f} MB)")

        # MinIO files info
        if self.stats['files_uploaded'] > 0:
            print("\n[MINIO FILES]")
            minio_files = self.list_minio_files()
            for f in minio_files[:20]:  # Show first 20 files
                print(f"    - s3://{MINIO_BUCKET}/{f}")

        print("\n" + "=" * 70)
        print("USAGE INSTRUCTIONS")
        print("=" * 70)
        print("""
To use the data lake for recommendations:

1. Load feature data (Local):
   import pandas as pd
   similarity = pd.read_parquet('data/lake/features/product_similarity.parquet')
   cf_features = pd.read_parquet('data/lake/features/cf_features.parquet')

2. Load from MinIO (for distributed training):
   from minio import Minio
   import pandas as pd
   from io import BytesIO

   client = Minio('localhost:9000', access_key='minioadmin', secret_key='minioadmin123', secure=False)
   obj = client.get_object('florus-datalake', 'datalake/features/product_similarity.parquet')
   df = pd.read_parquet(BytesIO(obj.read()))

3. Query similar products:
   product_id = 'P12345'
   similar = similarity[similarity['product_a'] == product_id].sort_values('similarity', ascending=False)

4. Train ML model:
   from sklearn.neighbors import NearestNeighbors
   # Use cb_features.parquet for content-based
   # Use cf_features.parquet for collaborative filtering
""")


def main():
    global MINIO_ENDPOINT, MINIO_BUCKET
    import argparse

    # Store defaults before argparse
    default_endpoint = MINIO_ENDPOINT
    default_bucket = MINIO_BUCKET

    parser = argparse.ArgumentParser(description="Florus Beauty Data Pipeline")
    parser.add_argument("--mode", choices=["full", "web-only", "lake-only", "features", "extract", "raw", "processed", "minio-sync"],
                       default="full", help="Pipeline mode")
    parser.add_argument("--limit", type=int, help="Limit records per dataset")
    parser.add_argument("--sync-minio", action="store_true", help="Sync data lake to MinIO after pipeline")
    parser.add_argument("--minio-endpoint", type=str, default=default_endpoint, help="MinIO endpoint")
    parser.add_argument("--minio-bucket", type=str, default=default_bucket, help="MinIO bucket name")
    args = parser.parse_args()

    # Update MinIO config from args
    MINIO_ENDPOINT = args.minio_endpoint
    MINIO_BUCKET = args.minio_bucket

    pipeline = DataPipeline()
    pipeline.run(mode=args.mode, limit=args.limit, sync_minio=args.sync_minio)


if __name__ == "__main__":
    main()
