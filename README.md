# ATM Service Rest API with Postgres and JWT Example

### Build, Run and Test

- Clone the project and go to the project root directory `atm-service-golang`
- Run docker compose command to start Postgres DB:
    ```
    docker-compose up --build
    ```
- Go to the service directory `atm-service-golang/service` and run backend server:
    ```
    go run ./cmd
    ```
- Tests implemented partially for server handlers using go-mock. To run all tests:
    ```
    go test ./server
    ```
### API Endpoints
#### Service expose five endpoints

1. Request to check app status:
```
curl http://localhost:5000/api/v1/health
```
Response:
```
{"status":"service is up and running"}
```
<br>

2. Request to signup new customer:
```
curl -w "\n" -s -X POST -H 'Accept: application/json' -H 'Content-Type: application/json' --data '{"first_name": "Natasha", "last_name": "Romanov", "email": "natasha@gmail.com", "pin_number": "1234", "account_number": "100000000099"}' http://localhost:5000/api/v1/auth/signup
```
Response:
```
{"jwt":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjdXMiOjEsImFjYyI6MSwiZXhwIjoxNjUwMzUwMjEyfQ.gUBuS-j7VoDp9CdSc_F3f2VfhTXneNKS4WkPHE-f0ow","customer":{"id":1,"first_name":"Natasha","last_name":"Romanov","email":"natasha@gmail.com","account":{"id":1,"number":"100000000099"}}}
```
It returns JWT token that is valid for 2 minutes for testing purposes.
It is recommended to copy the token to env variable:
```
TOKEN=eyJhbGciOiJIUzI1NiIsI...
```
<br>

3. Request to login existing user:
```
curl -w "\n" -s -X POST -H 'Accept: application/json' -H 'Content-Type: application/json' --data '{"pin_number": "1234", "account_number": "100000000099"}' http://localhost:5000/api/v1/auth/login
```
Response:
```
{"jwt":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjdXMiOjEsImFjYyI6MSwiZXhwIjoxNjUwMzUxNDE4fQ._CNJIng6uwVgYoZjjVgddEHnSW4ZyI1Md-CHu4H5IK8","customer":{"id":1,"first_name":"Natasha","last_name":"Romanov","email":"natasha@gmail.com","account":{"id":1,"number":"100000000099"}}}
```
If token has expired, you can login one more time to refresh the token.

<br>

4. Request to view account info, i.e. number, balance, transactions history:
```
curl -w "\n" -H 'Accept: application/json' -H "Authorization: Bearer ${TOKEN}" http://localhost:5000/api/v1/accounts/1
```
Response:
```
{"id":1,"number":"100000000099","balance":"0.00"}
```

<br>

5. Request to deposit or withdraw money from the account. API supports two types of transactions, 'deposit' and 'withdraw':
```
curl -w "\n" -s -X POST -H 'Accept: application/json' -H 'Content-Type: application/json' -H "Authorization: Bearer ${TOKEN}" --data '{"type": "deposit", "amount": "50.00", "account_id": 1}' http://localhost:5000/api/v1/transactions
```
Response:
```
{"id":"5b623eec-dc19-4849-a7ae-23251f69aeb2","amount":"50.00","created_at":"2022-04-19T07:01:15.685864Z","account_id":1}
```