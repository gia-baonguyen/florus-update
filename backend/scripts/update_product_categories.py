#!/usr/bin/env python3
"""Update product categories based on Sephora secondary_category data"""

import requests
import pandas as pd
from pathlib import Path

API_URL = 'http://localhost:8081/api'
DATA_DIR = Path(__file__).parent.parent / 'data'

# Category mapping (source -> slug)
CATEGORY_MAPPING = {
    'Hair Styling & Treatments': 'hair-styling',
    'Hair Styling': 'hair-styling',
    'Eye': 'eye', 'Face': 'face',
    'Moisturizers': 'moisturizer', 'Treatments': 'treatments',
    'Shampoo & Conditioner': 'shampoo-conditioner',
    'Lip': 'lip', 'Cleansers': 'cleanser',
    'Candles & Home Scents': 'candles-home-scents',
    'Brushes & Applicators': 'brushes-applicators',
    'Body Moisturizers': 'body-moisturizers',
    'Eye Care': 'eye', 'Masks': 'mask', 'Cheek': 'cheek',
    'Sunscreen': 'sunscreen', 'Bath & Shower': 'bath-shower',
    'High Tech Tools': 'high-tech-tools', 'Body Care': 'body-care',
    'Self Tanners': 'self-tanners', 'Lip Balms & Treatments': 'lip-balms',
    'Nail': 'nail', 'Accessories': 'accessories',
    'Value & Gift Sets': 'gift-sets',
    'Skincare': 'skincare', 'Makeup': 'makeup', 'Hair': 'hair',
    'Fragrance': 'fragrance', 'Bath & Body': 'bath-body',
    'Tools & Brushes': 'tools', 'Mini Size': 'mini',
    'Gifts': 'gifts', 'Men': 'men', 'Wellness': 'wellness',
}


def main():
    # Login
    session = requests.Session()
    resp = session.post(f'{API_URL}/auth/login',
                       json={'email': 'admin@florus.com', 'password': 'admin123'})
    token = resp.json().get('data', {}).get('token')
    session.headers['Authorization'] = f'Bearer {token}'
    print('[AUTH] Logged in')

    # Get all categories
    resp = session.get(f'{API_URL}/categories')
    categories = {c['slug']: c['id'] for c in resp.json().get('data', [])}
    cat_names = {c['id']: c['name'] for c in resp.json().get('data', [])}
    print(f'[CATEGORIES] Loaded {len(categories)} categories')

    def get_category_id(secondary, primary):
        if secondary and secondary in CATEGORY_MAPPING:
            slug = CATEGORY_MAPPING[secondary]
            if slug in categories:
                return categories[slug]
        if primary and primary in CATEGORY_MAPPING:
            slug = CATEGORY_MAPPING[primary]
            if slug in categories:
                return categories[slug]
        return categories.get('skincare', 14)

    # Load Sephora data
    df = pd.read_csv(DATA_DIR / 'product_info.csv')
    print(f'[DATA] Loaded {len(df)} Sephora products')

    # Build lookup by brand+name
    sephora_lookup = {}
    for _, row in df.iterrows():
        brand = str(row.get('brand_name', '')).lower().strip()
        name = str(row.get('product_name', '')).lower().strip()
        key = f'{brand}|{name}'
        sephora_lookup[key] = {
            'secondary': row.get('secondary_category'),
            'primary': row.get('primary_category')
        }

    # Get products from API
    resp = session.get(f'{API_URL}/products?limit=10000')
    data = resp.json().get('data', {})
    products = data.get('products', []) if isinstance(data, dict) else data
    print(f'[API] Found {len(products)} products in database')

    # Match and update
    updated = 0
    category_stats = {}

    for p in products:
        brand = str(p.get('brand', '')).lower().strip()
        name = str(p.get('name', '')).lower().strip()
        key = f'{brand}|{name}'

        if key in sephora_lookup:
            cat_info = sephora_lookup[key]
            new_cat_id = get_category_id(cat_info['secondary'], cat_info['primary'])

            # Track stats
            cat_name = cat_names.get(new_cat_id, 'Unknown')
            category_stats[cat_name] = category_stats.get(cat_name, 0) + 1

            if new_cat_id != p.get('category_id'):
                # Update via API
                resp = session.put(f"{API_URL}/products/{p['id']}",
                                  json={'category_id': new_cat_id})
                if resp.status_code in [200, 201]:
                    old_name = cat_names.get(p.get('category_id'), 'Unknown')
                    print(f"  [OK] {p['name'][:40]} | {old_name} -> {cat_name}")
                    updated += 1
                else:
                    print(f"  [ERR] {p['name'][:40]} | {resp.text[:50]}")

    print(f'\n[RESULT] Updated {updated} products')
    print(f'\n[DISTRIBUTION]')
    for cat, count in sorted(category_stats.items(), key=lambda x: -x[1]):
        print(f'  {cat}: {count}')


if __name__ == '__main__':
    main()
