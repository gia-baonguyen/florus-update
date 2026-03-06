"""
Simple script to populate Redis with test recommendations
This simulates what the Spark ALS pipeline would produce.
"""

import redis
import json
import random
from datetime import datetime

# Redis connection
REDIS_HOST = "localhost"
REDIS_PORT = 6379

# Sample product IDs from Sephora dataset
PRODUCT_IDS = list(range(9780, 9900))  # Sample product range

def get_redis_client():
    return redis.Redis(host=REDIS_HOST, port=REDIS_PORT, decode_responses=True)

def generate_user_recommendations(user_id: int, num_recs: int = 20) -> list:
    """Generate random recommendations for a user."""
    products = random.sample(PRODUCT_IDS, min(num_recs, len(PRODUCT_IDS)))
    return [
        {"product_id": pid, "score": round(random.uniform(0.5, 1.0), 4)}
        for pid in products
    ]

def generate_similar_products(product_id: int, num_similar: int = 10) -> list:
    """Generate random similar products."""
    available = [p for p in PRODUCT_IDS if p != product_id]
    similar = random.sample(available, min(num_similar, len(available)))
    return [
        {"product_id": pid, "score": round(random.uniform(0.3, 1.0), 4)}
        for pid in similar
    ]

def main():
    print("=" * 60)
    print("Test Spark Recommendations Generator")
    print(f"Started at: {datetime.now().isoformat()}")
    print("=" * 60)

    try:
        r = get_redis_client()
        r.ping()
        print(f"Connected to Redis at {REDIS_HOST}:{REDIS_PORT}")
    except Exception as e:
        print(f"Failed to connect to Redis: {e}")
        return

    # Generate recommendations for test users (IDs 1-100)
    user_count = 0
    for user_id in range(1, 101):
        recs = generate_user_recommendations(user_id)
        key = f"spark:recommendations:user:{user_id}"
        r.set(key, json.dumps(recs))
        r.expire(key, 86400 * 7)  # 7 days TTL
        user_count += 1

    print(f"Generated recommendations for {user_count} users")

    # Generate similar products
    product_count = 0
    for product_id in PRODUCT_IDS:
        sims = generate_similar_products(product_id)
        key = f"spark:similar:product:{product_id}"
        r.set(key, json.dumps(sims))
        r.expire(key, 86400 * 7)
        product_count += 1

    print(f"Generated similarities for {product_count} products")

    # Set metadata (RFC3339 format with timezone for Go parsing)
    r.set("spark:recommendations:last_updated", datetime.now().astimezone().isoformat())
    r.set("spark:recommendations:user_count", user_count)
    r.set("spark:recommendations:product_count", product_count)

    print("=" * 60)
    print(f"Completed at: {datetime.now().isoformat()}")
    print("=" * 60)

    # Verify
    print("\nVerification:")
    print(f"  Last updated: {r.get('spark:recommendations:last_updated')}")
    print(f"  User count: {r.get('spark:recommendations:user_count')}")
    print(f"  Product count: {r.get('spark:recommendations:product_count')}")

    # Sample check
    sample_key = "spark:recommendations:user:1"
    sample_data = r.get(sample_key)
    if sample_data:
        recs = json.loads(sample_data)
        print(f"\nSample recommendations for user 1:")
        for rec in recs[:3]:
            print(f"    Product {rec['product_id']}: score {rec['score']}")

if __name__ == "__main__":
    main()
