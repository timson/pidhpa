import requests
import json
import time

URL = "http://localhost:5000/produce"
MESSAGE = {"message": "test"}

def send_requests():
    while True:
        start_time = time.time()

        for _ in range(50):  # Send 50 requests
            try:
                response = requests.post(URL, json=MESSAGE)
                print(f"Sent: {MESSAGE}, Response: {response.status_code}")
            except requests.exceptions.RequestException as e:
                print(f"Request failed: {e}")

        # Ensure 1-second interval
        elapsed_time = time.time() - start_time
        sleep_time = max(0, 1.0 - elapsed_time)
        time.sleep(sleep_time)

if __name__ == "__main__":
    send_requests()
