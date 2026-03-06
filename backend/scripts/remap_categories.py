#!/usr/bin/env python3
"""
Re-map products to use new categories (24-41) based on Sephora secondary_category
"""

import requests
import pandas as pd
from pathlib import Path
import time

API_URL = 'http://localhost:8081/api'
DATA_DIR = Path(__file__).parent.parent / 'data'

# Full category mapping including new categories (24-41)
CATEGORY_MAPPING = {
    # Secondary categories -> new category slugs
    'Hair Styling & Treatments': 'hair-styling',
    'Hair Styling': 'hair-styling',
    'Eye': 'eye',
    'Face': 'face',
    'Moisturizers': 'moisturizer',
    'Treatments': 'treatments',
    'Shampoo & Conditioner': 'shampoo-conditioner',
    'Lip': 'lip',
    'Cleansers': 'cleanser',
    'Candles & Home Scents': 'candles-home-scents',
    'Brushes & Applicators': 'brushes-applicators',
    'Body Moisturizers': 'body-moisturizers',
    'Eye Care': 'eye',
    'Masks': 'mask',
    'Cheek': 'cheek',
    'Sunscreen': 'sunscreen',
    'Bath & Shower': 'bath-shower',
    'High Tech Tools': 'high-tech-tools',
    'Body Care': 'body-care',
    'Self Tanners': 'self-tanners',
    'Lip Balms & Treatments': 'lip-balms',
    'Nail': 'nail',
    'Accessories': 'accessories',
    'Value & Gift Sets': 'gift-sets',
    'Women': 'fragrance',
    'Makeup': 'makeup',
    'Mini Size': 'mini',
    # Primary categories
    'Skincare': 'skincare',
    'Hair': 'hair',
    'Fragrance': 'fragrance',
    'Bath & Body': 'bath-body',
    'Tools & Brushes': 'tools',
    'Gifts': 'gifts',
    'Men': 'men',
    'Wellness': 'wellness',
}


def main():
    # Login
    session = requests.Session()
    resp = session.post(f'{API_URL}/auth/login',
                       json={'email': 'admin@florus.com', 'password': 'admin123'})
    token = resp.json().get('data', {}).get('token')
    session.headers['Authorization'] = f'Bearer {token}'
    print('[AUTH] Logged in')

    # Get all categories with IDs
    resp = session.get(f'{API_URL}/categories')
    categories = {c['slug']: c['id'] for c in resp.json().get('data', [])}
    cat_names = {c['id']: c['name'] for c in resp.json().get('data', [])}
    print(f'[CATEGORIES] Loaded {len(categories)} categories')

    # Show new categories (24+)
    new_cats = {k: v for k, v in categories.items() if v >= 24}
    print(f'[NEW CATEGORIES] {len(new_cats)} new categories (ID >= 24):')
    for slug, cat_id in sorted(new_cats.items(), key=lambda x: x[1]):
        print(f'  {cat_id}: {slug}')

    def get_category_id(secondary, primary):
        """Get category ID, preferring new categories"""
        # Try secondary first (maps to new categories 24-41)
        if secondary and str(secondary) in CATEGORY_MAPPING:
            slug = CATEGORY_MAPPING[str(secondary)]
            if slug in categories:
                return categories[slug]
        # Fall back to primary
        if primary and str(primary) in CATEGORY_MAPPING:
            slug = CATEGORY_MAPPING[str(primary)]
            if slug in categories:
                return categories[slug]
        return categories.get('skincare', 14)

    # Load Sephora data
    print('\n[DATA] Loading Sephora product data...')
    df = pd.read_csv(DATA_DIR / 'product_info.csv')
    print(f'[DATA] Loaded {len(df)} Sephora products')

    # Build lookup by brand+name (lowercase, normalized)
    sephora_lookup = {}
    for _, row in df.iterrows():
        brand = str(row.get('brand_name', '')).lower().strip()
        name = str(row.get('product_name', '')).lower().strip()
        # Create multiple keys for matching
        key1 = f'{brand}|{name}'
        key2 = name[:50]  # Just name prefix
        key3 = f'{brand[:20]}|{name[:30]}'  # Shortened

        cat_info = {
            'secondary': row.get('secondary_category'),
            'primary': row.get('primary_category')
        }
        sephora_lookup[key1] = cat_info
        sephora_lookup[key2] = cat_info
        sephora_lookup[key3] = cat_info

    print(f'[LOOKUP] Built {len(sephora_lookup)} lookup entries')

    # Get all products with pagination
    print('\n[PRODUCTS] Fetching all products...')
    all_products = []
    page = 1
    while True:
        resp = session.get(f'{API_URL}/products?limit=100&page={page}')
        data = resp.json()
        products = data.get('data', [])
        meta = data.get('meta', {})

        if not products:
            break

        all_products.extend(products)
        total = meta.get('total', len(all_products))

        if page == 1:
            print(f'[PRODUCTS] Total: {total}')

        if len(all_products) >= total:
            break
        page += 1

    print(f'[PRODUCTS] Fetched {len(all_products)} products')

    # Re-map categories
    print('\n[REMAP] Re-mapping categories...')
    updated = 0
    not_matched = 0
    already_correct = 0
    cat_stats = {}

    for i, p in enumerate(all_products):
        brand = str(p.get('brand', '')).lower().strip()
        name = str(p.get('name', '')).lower().strip()
        current_cat = p.get('category_id', 0)

        # Try to find in Sephora data
        key1 = f'{brand}|{name}'
        key2 = name[:50]
        key3 = f'{brand[:20]}|{name[:30]}'

        cat_info = sephora_lookup.get(key1) or sephora_lookup.get(key2) or sephora_lookup.get(key3)

        if cat_info:
            new_cat_id = get_category_id(cat_info['secondary'], cat_info['primary'])

            # Track stats
            cat_name = cat_names.get(new_cat_id, f'ID:{new_cat_id}')
            cat_stats[cat_name] = cat_stats.get(cat_name, 0) + 1

            if new_cat_id != current_cat:
                # Update via API
                resp = session.put(f"{API_URL}/products/{p['id']}",
                                  json={'category_id': new_cat_id})
                if resp.status_code in [200, 201]:
                    updated += 1
                else:
                    print(f'  [ERR] {p["id"]}: {resp.text[:50]}')
            else:
                already_correct += 1
        else:
            not_matched += 1
            # Keep original category in stats
            cat_name = cat_names.get(current_cat, f'ID:{current_cat}')
            cat_stats[cat_name] = cat_stats.get(cat_name, 0) + 1

        # Progress
        if (i + 1) % 500 == 0:
            print(f'  Progress: {i+1}/{len(all_products)} ({updated} updated)')

        # Rate limiting
        if updated % 50 == 0 and updated > 0:
            time.sleep(0.1)

    print(f'\n[RESULT]')
    print(f'  Updated: {updated}')
    print(f'  Already correct: {already_correct}')
    print(f'  Not matched: {not_matched}')

    print(f'\n[DISTRIBUTION]')
    for cat, count in sorted(cat_stats.items(), key=lambda x: -x[1])[:20]:
        print(f'  {cat}: {count}')


if __name__ == '__main__':
    main()
