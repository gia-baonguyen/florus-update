#!/usr/bin/env python3
"""
Generate User Events for Recommendation System Testing
=======================================================
This script generates realistic user events (views, cart, purchases)
to simulate user behavior for the recommendation system.

Usage:
    python generate_user_events.py --users 100 --events 5000
"""

import random
import time
import json
import requests
import argparse
from datetime import datetime, timedelta
from typing import List, Dict

# Configuration
API_BASE_URL = "http://localhost:8081/api"
ADMIN_EMAIL = "admin@florus.com"
ADMIN_PASSWORD = "admin123"

# Event type weights (probability of each event type)
EVENT_WEIGHTS = {
    "product_view": 50,      # Most common
    "product_click": 20,
    "add_to_cart": 15,
    "add_to_wishlist": 5,
    "purchase": 8,
    "review": 2,
}

class UserEventGenerator:
    def __init__(self):
        self.session = requests.Session()
        self.token = None
        self.products = []
        self.categories = []
        self.users = []

    def login(self) -> bool:
        """Authenticate with the API"""
        print(f"[AUTH] Logging in as {ADMIN_EMAIL}...")
        try:
            response = self.session.post(
                f"{API_BASE_URL}/auth/login",
                json={"email": ADMIN_EMAIL, "password": ADMIN_PASSWORD}
            )
            if response.status_code == 200:
                data = response.json()
                self.token = data.get("token")
                self.session.headers["Authorization"] = f"Bearer {self.token}"
                print("[AUTH] Login successful!")
                return True
            else:
                print(f"[AUTH] Login failed: {response.text}")
                return False
        except Exception as e:
            print(f"[AUTH] Error: {e}")
            return False

    def fetch_products(self) -> bool:
        """Fetch all products from the API"""
        print("[DATA] Fetching products...")
        try:
            response = self.session.get(f"{API_BASE_URL}/products?limit=1000")
            if response.status_code == 200:
                data = response.json()
                self.products = data.get("products", [])
                print(f"[DATA] Found {len(self.products)} products")
                return True
            else:
                print(f"[DATA] Failed to fetch products: {response.text}")
                return False
        except Exception as e:
            print(f"[DATA] Error: {e}")
            return False

    def fetch_categories(self) -> bool:
        """Fetch all categories"""
        print("[DATA] Fetching categories...")
        try:
            response = self.session.get(f"{API_BASE_URL}/categories")
            if response.status_code == 200:
                self.categories = response.json()
                print(f"[DATA] Found {len(self.categories)} categories")
                return True
        except Exception as e:
            print(f"[DATA] Error: {e}")
        return False

    def create_test_users(self, num_users: int) -> List[Dict]:
        """Create test users for event generation"""
        print(f"[USERS] Creating {num_users} test users...")
        users = []

        for i in range(num_users):
            user = {
                "id": 1000 + i,  # Simulated user IDs
                "email": f"testuser{i}@florus.com",
                "preferred_categories": random.sample(
                    [c["id"] for c in self.categories] if self.categories else [1, 2, 3],
                    k=min(3, len(self.categories) if self.categories else 3)
                ),
                "activity_level": random.choice(["low", "medium", "high"]),
            }
            users.append(user)

        self.users = users
        print(f"[USERS] Created {len(users)} simulated users")
        return users

    def select_event_type(self) -> str:
        """Select an event type based on weights"""
        types = list(EVENT_WEIGHTS.keys())
        weights = list(EVENT_WEIGHTS.values())
        return random.choices(types, weights=weights, k=1)[0]

    def select_product(self, user: Dict) -> Dict:
        """Select a product based on user preferences"""
        if not self.products:
            return None

        # 60% chance to select from preferred categories
        if random.random() < 0.6 and user.get("preferred_categories"):
            preferred = [p for p in self.products
                        if p.get("category_id") in user["preferred_categories"]]
            if preferred:
                return random.choice(preferred)

        # Otherwise random product
        return random.choice(self.products)

    def generate_event(self, user: Dict, event_type: str, product: Dict) -> Dict:
        """Generate a single event"""
        # Random timestamp in last 90 days
        days_ago = random.randint(0, 90)
        hours_ago = random.randint(0, 23)
        timestamp = datetime.now() - timedelta(days=days_ago, hours=hours_ago)

        event = {
            "user_id": user["id"],
            "session_id": f"session_{user['id']}_{days_ago}",
            "event_type": event_type,
            "product_id": product["id"],
            "category_id": product.get("category_id"),
            "created_at": timestamp.isoformat(),
        }

        # Add event-specific data
        if event_type == "product_view":
            event["metadata"] = json.dumps({"time_spent": random.randint(5, 300)})
        elif event_type == "add_to_cart":
            event["quantity"] = random.randint(1, 3)
            event["price"] = product.get("price", 0)
        elif event_type == "purchase":
            event["quantity"] = random.randint(1, 2)
            event["price"] = product.get("price", 0)
            event["order_id"] = random.randint(1000, 9999)
        elif event_type == "review":
            event["rating"] = random.randint(3, 5)  # Bias towards positive reviews

        return event

    def send_event(self, event: Dict) -> bool:
        """Send event to the API"""
        try:
            response = self.session.post(
                f"{API_BASE_URL}/events/track",
                json=event
            )
            return response.status_code in [200, 201]
        except Exception as e:
            return False

    def generate_user_journey(self, user: Dict, num_events: int) -> List[Dict]:
        """Generate a realistic user journey"""
        events = []
        viewed_products = []

        for _ in range(num_events):
            event_type = self.select_event_type()

            # Realistic constraints:
            # - Can only add to cart after viewing
            # - Can only purchase after adding to cart
            # - Reviews only after purchase

            if event_type in ["add_to_cart", "add_to_wishlist", "purchase", "review"]:
                if not viewed_products:
                    event_type = "product_view"
                    product = self.select_product(user)
                    viewed_products.append(product)
                else:
                    product = random.choice(viewed_products)
            else:
                product = self.select_product(user)
                if event_type == "product_view" and product:
                    viewed_products.append(product)

            if product:
                event = self.generate_event(user, event_type, product)
                events.append(event)

        return events

    def run(self, num_users: int = 50, events_per_user: int = 20):
        """Main execution flow"""
        print("=" * 60)
        print("FLORUS BEAUTY - USER EVENT GENERATOR")
        print("=" * 60)

        # Login
        if not self.login():
            print("\n[ERROR] Failed to authenticate")
            return

        # Fetch data
        self.fetch_products()
        self.fetch_categories()

        if not self.products:
            print("\n[ERROR] No products found. Please import products first.")
            return

        # Create users
        self.create_test_users(num_users)

        # Generate and send events
        print(f"\n[EVENTS] Generating ~{num_users * events_per_user} events...")
        total_events = 0
        successful_events = 0

        for user in self.users:
            # Vary events per user based on activity level
            if user["activity_level"] == "high":
                user_events = events_per_user * 2
            elif user["activity_level"] == "low":
                user_events = events_per_user // 2
            else:
                user_events = events_per_user

            events = self.generate_user_journey(user, user_events)

            for event in events:
                total_events += 1
                if self.send_event(event):
                    successful_events += 1

                if total_events % 100 == 0:
                    print(f"[EVENTS] Progress: {total_events} events sent...")

            # Rate limiting
            time.sleep(0.05)

        print("\n" + "=" * 60)
        print("GENERATION SUMMARY")
        print("=" * 60)
        print(f"Total events generated: {total_events}")
        print(f"Successfully sent: {successful_events}")
        print(f"Failed: {total_events - successful_events}")


def main():
    parser = argparse.ArgumentParser(description="Generate user events for testing")
    parser.add_argument("--users", type=int, default=50, help="Number of simulated users")
    parser.add_argument("--events", type=int, default=20, help="Events per user")
    args = parser.parse_args()

    generator = UserEventGenerator()
    generator.run(num_users=args.users, events_per_user=args.events)


if __name__ == "__main__":
    main()
