#!/usr/bin/env python3
"""
Update Product Images from Cosmetic Brand Products Dataset
"""

import requests
import pandas as pd
import time
from pathlib import Path

# Configuration
API_BASE_URL = "http://localhost:8081/api"
ADMIN_EMAIL = "admin@florus.com"
ADMIN_PASSWORD = "admin123"
DATA_DIR = Path(__file__).parent.parent / "data"

def main():
    print("=" * 60)
    print("UPDATE PRODUCT IMAGES")
    print("=" * 60)

    session = requests.Session()

    # Login
    print("\n[AUTH] Logging in...")
    response = session.post(
        f"{API_BASE_URL}/auth/login",
        json={"email": ADMIN_EMAIL, "password": ADMIN_PASSWORD}
    )
    if response.status_code != 200:
        print("[AUTH] [X] Failed to login")
        return

    data = response.json().get("data", response.json())
    token = data.get("token")
    session.headers["Authorization"] = f"Bearer {token}"
    print("[AUTH] [OK] Logged in")

    # Load image URLs from output.csv
    print("\n[IMAGES] Loading image URLs from output.csv...")
    images_file = DATA_DIR / "output.csv"
    if not images_file.exists():
        print("[IMAGES] [X] output.csv not found!")
        return

    df = pd.read_csv(images_file)
    print(f"[IMAGES] Found {len(df)} products with images")

    # Filter valid image URLs
    valid_images = []
    for idx, row in df.iterrows():
        img_url = row.get('image_link', '')
        if pd.notna(img_url) and str(img_url).startswith('http'):
            valid_images.append(str(img_url))

    print(f"[IMAGES] Valid image URLs: {len(valid_images)}")

    if not valid_images:
        print("[IMAGES] [X] No valid images found!")
        return

    # Get products from database
    print("\n[PRODUCTS] Fetching products from database...")
    page = 1
    all_products = []

    while True:
        response = session.get(f"{API_BASE_URL}/products?page={page}&limit=100")
        if response.status_code != 200:
            break

        data = response.json()
        products = data.get("data", [])
        if isinstance(products, dict):
            products = products.get("products", [])

        if not products:
            break

        all_products.extend(products)
        page += 1

        if page > 100:  # Safety limit
            break

    print(f"[PRODUCTS] Found {len(all_products)} products in database")

    # Update products with images
    print("\n[UPDATE] Updating products with images...")
    updated = 0
    errors = 0

    for i, product in enumerate(all_products):
        product_id = product.get("id")
        current_image = product.get("image_url", "")

        # Skip if already has image
        if current_image and current_image.startswith("http"):
            continue

        # Get image from pool (cycle through)
        image_url = valid_images[i % len(valid_images)]

        # Update product
        try:
            response = session.put(
                f"{API_BASE_URL}/products/{product_id}",
                json={"image_url": image_url}
            )

            if response.status_code in [200, 201]:
                updated += 1
                if updated % 100 == 0:
                    print(f"[UPDATE] Progress: {updated} products updated...")
            else:
                errors += 1
        except Exception as e:
            errors += 1

        # Rate limiting
        if i % 50 == 0:
            time.sleep(0.1)

    # Summary
    print("\n" + "=" * 60)
    print("SUMMARY")
    print("=" * 60)
    print(f"Products updated: {updated}")
    print(f"Errors: {errors}")
    print(f"Image pool size: {len(valid_images)}")

    # Verify
    print("\n[VERIFY] Checking a sample product...")
    response = session.get(f"{API_BASE_URL}/products?limit=1")
    if response.status_code == 200:
        data = response.json()
        products = data.get("data", [])
        if isinstance(products, dict):
            products = products.get("products", [])
        if products:
            print(f"Sample product image_url: {products[0].get('image_url', 'N/A')}")

if __name__ == "__main__":
    main()
