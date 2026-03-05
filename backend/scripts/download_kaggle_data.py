#!/usr/bin/env python3
"""
Download All Kaggle Datasets for Florus Beauty
===============================================
Downloads real beauty product data from 3 Kaggle datasets:

1. Sephora Products & Reviews (8K+ products)
   - nadyinky/sephora-products-and-skincare-reviews

2. Cosmetics Dataset (1.4K+ products)
   - kingabzpro/cosmetics-datasets

3. Amazon Beauty Products (2M+ ratings)
   - skillsmuggler/amazon-ratings

Usage:
    pip install kaggle pandas
    # Set up kaggle.json credentials
    python download_kaggle_data.py
"""

import os
import sys
from pathlib import Path

# Data directory
DATA_DIR = Path(__file__).parent.parent / "data"
DATA_DIR.mkdir(parents=True, exist_ok=True)

# All 3 Kaggle datasets
DATASETS = [
    {
        "name": "Sephora Products & Skincare Reviews",
        "kaggle_id": "nadyinky/sephora-products-and-skincare-reviews",
        "description": "8K+ Sephora products with ratings, reviews, and product details",
        "expected_files": ["product_info.csv", "reviews_0-250.csv"]
    },
    {
        "name": "Cosmetics Dataset",
        "kaggle_id": "kingabzpro/cosmetics-datasets",
        "description": "1.4K+ skincare products with ingredients and skin type compatibility",
        "expected_files": ["cosmetics.csv"]
    },
    {
        "name": "Amazon Beauty Products",
        "kaggle_id": "skillsmuggler/amazon-ratings",
        "description": "2M+ product ratings for recommendation system training",
        "expected_files": ["ratings_Beauty.csv"]
    }
]


def check_kaggle_credentials():
    """Check if Kaggle credentials are configured"""
    kaggle_json = Path.home() / ".kaggle" / "kaggle.json"
    if not kaggle_json.exists():
        # Try Windows path
        kaggle_json = Path(os.environ.get("USERPROFILE", "")) / ".kaggle" / "kaggle.json"

    if not kaggle_json.exists():
        print("=" * 60)
        print("KAGGLE CREDENTIALS NOT FOUND")
        print("=" * 60)
        print("""
To download Kaggle datasets, you need to set up API credentials:

1. Go to https://www.kaggle.com and sign in (or create account)
2. Click on your profile picture → Settings
3. Scroll to "API" section → Click "Create New API Token"
4. This downloads kaggle.json file
5. Move kaggle.json to:
   - Windows: C:\\Users\\<username>\\.kaggle\\kaggle.json
   - Linux/Mac: ~/.kaggle/kaggle.json

Then run this script again.
""")
        return False
    return True


def download_dataset(kaggle_id: str, target_dir: Path):
    """Download a Kaggle dataset"""
    try:
        import kaggle
        kaggle.api.authenticate()

        print(f"[DOWNLOAD] Downloading {kaggle_id}...")
        kaggle.api.dataset_download_files(
            kaggle_id,
            path=str(target_dir),
            unzip=True
        )
        print(f"[DOWNLOAD] OK - Successfully downloaded to {target_dir}")
        return True
    except Exception as e:
        print(f"[ERROR] FAILED - {kaggle_id}: {e}")
        return False


def main():
    print("=" * 60)
    print("FLORUS BEAUTY - KAGGLE DATA DOWNLOADER")
    print("=" * 60)
    print(f"\nTarget directory: {DATA_DIR}\n")

    if not check_kaggle_credentials():
        sys.exit(1)

    try:
        import kaggle
    except ImportError:
        print("[ERROR] kaggle package not installed. Run: pip install kaggle")
        sys.exit(1)

    print(f"[INFO] Downloading {len(DATASETS)} datasets...\n")

    success_count = 0
    for i, dataset in enumerate(DATASETS, 1):
        print(f"\n{'=' * 60}")
        print(f"DATASET {i}/{len(DATASETS)}: {dataset['name']}")
        print(f"{'=' * 60}")
        print(f"Description: {dataset['description']}")
        print(f"Kaggle ID: {dataset['kaggle_id']}")
        print(f"Expected files: {', '.join(dataset['expected_files'])}")
        print()

        if download_dataset(dataset["kaggle_id"], DATA_DIR):
            success_count += 1
            # List downloaded files
            for expected in dataset['expected_files']:
                file_path = DATA_DIR / expected
                if file_path.exists():
                    size_mb = file_path.stat().st_size / (1024 * 1024)
                    print(f"  [OK] {expected} ({size_mb:.2f} MB)")
                else:
                    # Search for similar file
                    found = False
                    for f in DATA_DIR.glob("*.csv"):
                        if expected.split('.')[0].lower() in f.name.lower():
                            size_mb = f.stat().st_size / (1024 * 1024)
                            print(f"  [OK] {f.name} ({size_mb:.2f} MB)")
                            found = True
                            break
                    if not found:
                        print(f"  [?] {expected} (not found)")

    # Summary
    print("\n" + "=" * 60)
    print("DOWNLOAD SUMMARY")
    print("=" * 60)
    print(f"Successfully downloaded: {success_count}/{len(DATASETS)} datasets")

    print("\n[FILES] All CSV files in data directory:")
    total_size = 0
    for f in sorted(DATA_DIR.glob("*.csv")):
        size_mb = f.stat().st_size / (1024 * 1024)
        total_size += size_mb
        print(f"  - {f.name} ({size_mb:.2f} MB)")
    print(f"\nTotal size: {total_size:.2f} MB")

    print("\n" + "=" * 60)
    print("NEXT STEPS")
    print("=" * 60)
    print("1. Make sure backend API is running:")
    print("   cd backend && go run ./cmd/api")
    print("\n2. Import products into database:")
    print("   python import_sephora_data.py")
    print("\n3. Generate user events for recommendations:")
    print("   python generate_user_events.py")
    print("=" * 60)


if __name__ == "__main__":
    main()
