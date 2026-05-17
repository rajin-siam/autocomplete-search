from locust import HttpUser, task, between

class PlaceSearchUser(HttpUser):
    wait_time = between(1, 3)
    
    @task(3)
    def search_dhaka(self):
        self.client.get("/api?q=dhaka")
    
    @task(2)
    def search_chittagong(self):
        self.client.get("/api?q=chittagong")
    
    @task(2)
    def search_sylhet(self):
        self.client.get("/api?q=sylhet")
    
    @task(1)
    def search_fuzzy(self):
        self.client.get("/api?q=dhka")
    
    @task(1)
    def search_short(self):
        self.client.get("/api?q=dh")
