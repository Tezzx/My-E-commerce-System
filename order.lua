wrk.method = "POST"
wrk.path = "/home/order/"
wrk.headers["Content-Type"] = "application/json"
wrk.headers["Authorization"] = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyaWQiOjIsInN1YiI6InVzZXJfYXV0aCIsImV4cCI6MTc3OTc4MDI0MCwiaWF0IjoxNzc5NjkzODQwfQ.YEKLcFxPiZkQh2PetBMkY-249x_S94tpK7xQB7kC-lE"
wrk.body = '{"goodsId": 1, "buyNum": 1}'