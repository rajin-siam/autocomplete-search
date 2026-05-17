from locust import HttpUser, task, between, constant
import random
import json

class PlaceSearchUser(HttpUser):
    """Base user class for place search load testing"""
    wait_time = between(1, 3)  # Wait 1-3 seconds between requests
    
    def on_start(self):
        """Initialize test data when user starts"""
        # Common search terms based on the integration tests
        self.search_queries = [
            "Dhaka",
            "dhka",  # Typo test
            "dh",    # Prefix test
            "Dinajpur",
            "nonexistent",  # No results test
            "Cox's Bazar",  # Special characters test
            "ঢাকা",  # Unicode test
            "London",
            "Paris",
            "Tokyo",
            "New York",
            "Sydney",
            "Berlin",
            "Moscow",
            "Beijing",
            "Mumbai",
            "Cairo",
            "Lima"
        ]
    
    @task
    def search_places(self):
        """Basic search task - random queries"""
        query = random.choice(self.search_queries)
        self.client.get(f"/api?q={query}", name="/api search")
    
    @task(3)
    def exact_match_search(self):
        """Task for exact matches - higher weight"""
        query = random.choice(["Dhaka", "Dinajpur", "London", "Paris"])
        self.client.get(f"/api?q={query}", name="/api exact match")
    
    @task(2)
    def fuzzy_search(self):
        """Task for fuzzy/partial matches"""
        typo_queries = ["dhka", "lndon", "pari", "toky"]
        query = random.choice(typo_queries)
        self.client.get(f"/api?q={query}", name="/api fuzzy search")
    
    @task(1)
    def prefix_search(self):
        """Task for prefix/autocomplete matches"""
        prefix_queries = ["dh", "lo", "pa", "to", "ne", "ca"]
        query = random.choice(prefix_queries)
        self.client.get(f"/api?q={query}", name="/api prefix search")
    
    @task(1)
    def unicode_search(self):
        """Task for Unicode/Bengali searches"""
        unicode_queries = ["ঢাকা", "চট্টগ্রাম", "খুলনা"]
        query = random.choice(unicode_queries)
        self.client.get(f"/api?q={query}", name="/api unicode search")

class MobileUser(PlaceSearchUser):
    """User simulating mobile app behavior - more frequent requests"""
    wait_time = between(0.5, 2)  # More frequent requests
    
    @task
    def mobile_search(self):
        """Mobile users do more searches with shorter queries"""
        query = random.choice(["dh", "lo", "pa", "to", "ny", "sy", "bd", "in", "uk"])
        self.client.get(f"/api?q={query}", name="/api mobile search")

class APIUser(PlaceSearchUser):
    """User simulating API client behavior"""
    wait_time = between(0.1, 0.5)  # Very frequent requests, API calls
    
    @task
    def api_search(self):
        """API clients make many rapid requests"""
        query = random.choice(self.search_queries[:5])  # Smaller subset for speed
        self.client.get(f"/api?q={query}", name="/api rapid search")