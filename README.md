# ATM Service REST API

A REST API in Go simulating basic ATM operations (signup, login, deposit, withdraw, balance check). Uses PostgreSQL for storage, JWT for authentication, and runs via Docker Compose.

### Build, Run and Test

- Clone the project and go to the project root directory `atm-service-golang`
- Create your own `.env` file from the sample:
    ```bash
    cp .env_sample .env
    ```
- Run docker compose to start the backend server and Postgres DB:
    ```bash
    docker compose up --build
    ```
- The API is available at `http://localhost:8080/api/v1`
- To run tests:
    ```bash
    cd service && go test ./server/...
    ```

### API Endpoints

#### 1. Health check
```bash
curl http://localhost:8080/api/v1/health
```
Response:
```json
{"status":"service is up and running"}
```

#### 2. Sign up a new customer
```bash
curl -s -X POST -H 'Content-Type: application/json' \
  --data '{"first_name": "Natasha", "last_name": "Romanov", "email": "natasha@gmail.com", "pin_number": "1234"}' \
  http://localhost:8080/api/v1/auth/signup
```
Response includes an access token (default 15min expiry) and a refresh token (default 7 days):
```json
{"jwt":"<access_token>","refresh_token":"<refresh_token>","customer":{"id":1,"first_name":"Natasha","last_name":"Romanov","email":"natasha@gmail.com","account":{"id":1,"number":"000000000001"}}}
```
Save the access token for authenticated requests:
```bash
TOKEN=<access_token>
```

#### 3. Log in an existing customer
```bash
curl -s -X POST -H 'Content-Type: application/json' \
  --data '{"pin_number": "1234", "email": "natasha@gmail.com"}' \
  http://localhost:8080/api/v1/auth/login
```
With jq, save the token directly:
```bash
TOKEN=$(curl -s -X POST -H 'Content-Type: application/json' \
  --data '{"pin_number": "1234", "email": "natasha@gmail.com"}' \
  http://localhost:8080/api/v1/auth/login | jq -r '.jwt')
```

#### 4. Refresh tokens
When the access token expires, use the refresh token to get a new pair:
```bash
curl -s -X POST -H 'Content-Type: application/json' \
  --data '{"refresh_token": "<refresh_token>"}' \
  http://localhost:8080/api/v1/auth/refresh
```
Response:
```json
{"jwt":"<new_access_token>","refresh_token":"<new_refresh_token>"}
```

#### 5. View account info (balance, transactions)
```bash
curl -H "Authorization: Bearer ${TOKEN}" http://localhost:8080/api/v1/accounts/1
```
Response:
```json
{"id":1,"number":"000000000001","balance":"0.00"}
```

#### 6. Deposit or withdraw money
Supports `deposit` and `withdraw` transaction types:
```bash
curl -s -X POST -H 'Content-Type: application/json' \
  -H "Authorization: Bearer ${TOKEN}" \
  --data '{"type": "deposit", "amount": "50.00", "account_id": 1}' \
  http://localhost:8080/api/v1/transactions
```
Response:
```json
{"id":"5b623eec-dc19-4849-a7ae-23251f69aeb2","amount":"50.00","created_at":"2022-04-19T07:01:15.685864Z"}
```