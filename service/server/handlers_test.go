package server

import (
	"database/sql"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/adruzhkin/atm-service-golang/service/db"
	"github.com/adruzhkin/atm-service-golang/service/jwt"
	"github.com/adruzhkin/atm-service-golang/service/mocks"
	"github.com/adruzhkin/atm-service-golang/service/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// ---------------------------------------------------------------------------
// CheckHealth
// ---------------------------------------------------------------------------

func TestCheckHealth(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepo(ctrl)
	repo.EXPECT().Ping().Return(nil)

	port := 8080
	server := New(&port)
	server.DB = repo

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	server.CheckHealth().ServeHTTP(rr, req)

	res := rr.Result()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.JSONEq(t, `{"status":"ok"}`, string(body))
}

func TestCheckHealth_DBDown(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepo(ctrl)
	repo.EXPECT().Ping().Return(errors.New("connection refused"))

	port := 8080
	server := New(&port)
	server.DB = repo

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	server.CheckHealth().ServeHTTP(rr, req)

	res := rr.Result()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusServiceUnavailable, res.StatusCode)
	assert.JSONEq(t, `{"status":"service unavailable"}`, string(body))
}

// ---------------------------------------------------------------------------
// SignupCustomer
// ---------------------------------------------------------------------------

func TestSignupCustomer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	accAfter := models.Account{ID: 0, Number: "100000000099"}
	cusAfter := models.Customer{
		ID: 7, FirstName: "Natasha", LastName: "Romanov",
		Email: "natasha@gmail.com", Account: &accAfter,
	}

	repo := mocks.NewMockRepo(ctrl)
	repo.EXPECT().CreateCustomer(gomock.Any()).DoAndReturn(func(cus *models.Customer) error {
		assert.Equal(t, "Natasha", cus.FirstName)
		assert.Equal(t, "Romanov", cus.LastName)
		assert.Equal(t, "natasha@gmail.com", cus.Email)
		assert.NotEmpty(t, cus.PINHash)
		*cus = cusAfter
		return nil
	})

	jwtMock := mocks.NewMockJWT(ctrl)
	jwtMock.EXPECT().GenerateTokenPair(7, 0).Return("access_token", "refresh_token", nil)

	port := 8080
	server := New(&port)
	server.DB = repo
	server.JWT = jwtMock

	reqBody := `{"first_name":"Natasha","last_name":"Romanov","email":"natasha@gmail.com","pin_number":"1234"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/signup", strings.NewReader(reqBody))
	rr := httptest.NewRecorder()

	server.SignupCustomer().ServeHTTP(rr, req)

	res := rr.Result()
	body, _ := io.ReadAll(res.Body)
	expBody := `{"jwt":"access_token","refresh_token":"refresh_token","customer":{"id":7,"first_name":"Natasha","last_name":"Romanov","email":"natasha@gmail.com","account":{"id":0,"number":"100000000099"}}}`

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.JSONEq(t, expBody, string(body))
}

func TestSignupCustomer_InvalidJSON(t *testing.T) {
	port := 8080
	server := New(&port)

	req := httptest.NewRequest(http.MethodPost, "/auth/signup", strings.NewReader(`{bad json`))
	rr := httptest.NewRecorder()

	server.SignupCustomer().ServeHTTP(rr, req)

	res := rr.Result()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.JSONEq(t, `{"error":"failed to parse customer request body"}`, string(body))
}

func TestSignupCustomer_DuplicateEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepo(ctrl)
	repo.EXPECT().CreateCustomer(gomock.Any()).Return(&pq.Error{Code: "23505"})

	port := 8080
	server := New(&port)
	server.DB = repo

	reqBody := `{"first_name":"Natasha","last_name":"Romanov","email":"natasha@gmail.com","pin_number":"1234"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/signup", strings.NewReader(reqBody))
	rr := httptest.NewRecorder()

	server.SignupCustomer().ServeHTTP(rr, req)

	res := rr.Result()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusConflict, res.StatusCode)
	assert.JSONEq(t, `{"error":"email already registered"}`, string(body))
}

// ---------------------------------------------------------------------------
// LoginCustomer
// ---------------------------------------------------------------------------

func TestLoginCustomer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	credentials := models.CustomerCredentials{PINNumber: "1234", Email: "natasha@gmail.com"}
	acc := models.Account{Number: "100000000099"}
	cus := models.Customer{
		ID: 7, FirstName: "Natasha", LastName: "Romanov",
		Email: "natasha@gmail.com", Account: &acc,
	}

	repo := mocks.NewMockRepo(ctrl)
	repo.EXPECT().GetCustomerByCredentials(&credentials).Return(&cus, nil)

	jwtMock := mocks.NewMockJWT(ctrl)
	jwtMock.EXPECT().GenerateTokenPair(7, 0).Return("access_token", "refresh_token", nil)

	port := 8080
	server := New(&port)
	server.DB = repo
	server.JWT = jwtMock

	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{"pin_number":"1234","email":"natasha@gmail.com"}`))
	rr := httptest.NewRecorder()

	server.LoginCustomer().ServeHTTP(rr, req)

	res := rr.Result()
	body, _ := io.ReadAll(res.Body)
	expBody := `{"jwt":"access_token","refresh_token":"refresh_token","customer":{"id":7,"first_name":"Natasha","last_name":"Romanov","email":"natasha@gmail.com","account":{"id":0,"number":"100000000099"}}}`

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.JSONEq(t, expBody, string(body))
}

func TestLoginCustomer_InvalidCredentials(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepo(ctrl)
	repo.EXPECT().GetCustomerByCredentials(gomock.Any()).Return(nil, errors.New("invalid credentials"))

	port := 8080
	server := New(&port)
	server.DB = repo

	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{"pin_number":"0000","email":"natasha@gmail.com"}`))
	rr := httptest.NewRecorder()

	server.LoginCustomer().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Result().StatusCode)
}

// ---------------------------------------------------------------------------
// RefreshToken
// ---------------------------------------------------------------------------

func TestRefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jwtMock := mocks.NewMockJWT(ctrl)
	jwtMock.EXPECT().VerifyRefreshToken("old_refresh").Return(&jwt.Claims{
		CustomerID: 7, AccountID: 3, TokenType: "refresh",
	}, nil)
	jwtMock.EXPECT().GenerateTokenPair(7, 3).Return("new_access", "new_refresh", nil)

	port := 8080
	server := New(&port)
	server.JWT = jwtMock

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", strings.NewReader(`{"refresh_token":"old_refresh"}`))
	rr := httptest.NewRecorder()

	server.RefreshToken().ServeHTTP(rr, req)

	res := rr.Result()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.JSONEq(t, `{"jwt":"new_access","refresh_token":"new_refresh"}`, string(body))
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jwtMock := mocks.NewMockJWT(ctrl)
	jwtMock.EXPECT().VerifyRefreshToken("bad_token").Return(nil, errors.New("token expired"))

	port := 8080
	server := New(&port)
	server.JWT = jwtMock

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", strings.NewReader(`{"refresh_token":"bad_token"}`))
	rr := httptest.NewRecorder()

	server.RefreshToken().ServeHTTP(rr, req)

	res := rr.Result()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	assert.JSONEq(t, `{"error":"invalid or expired refresh token"}`, string(body))
}

// ---------------------------------------------------------------------------
// Authenticate middleware
// ---------------------------------------------------------------------------

func TestAuthenticate_MissingAuthHeader(t *testing.T) {
	port := 8080
	server := New(&port)

	dummy := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/accounts/1", nil)
	rr := httptest.NewRecorder()

	server.Authenticate(dummy).ServeHTTP(rr, req)

	res := rr.Result()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	assert.JSONEq(t, `{"error":"jwt token not provided"}`, string(body))
}

func TestAuthenticate_InvalidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jwtMock := mocks.NewMockJWT(ctrl)
	jwtMock.EXPECT().VerifyAccessToken("invalid_token").Return(nil, errors.New("invalid token"))

	port := 8080
	server := New(&port)
	server.JWT = jwtMock

	dummy := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/accounts/1", nil)
	req.Header.Set("Authorization", "Bearer invalid_token")
	rr := httptest.NewRecorder()

	server.Authenticate(dummy).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Result().StatusCode)
}

// ---------------------------------------------------------------------------
// GetAccount
// ---------------------------------------------------------------------------

func TestGetAccount_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	txID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	txTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	repo := mocks.NewMockRepo(ctrl)
	repo.EXPECT().GetAccountByID(5).Return(&models.Account{ID: 5, Number: "000000000005"}, nil)
	repo.EXPECT().GetTransactionsByAccountID(5).Return([]models.Transaction{
		{ID: txID, AmountInCents: 1050, CreatedAt: txTime, AccountID: 5},
	}, nil)
	repo.EXPECT().GetTransactionsBalanceByAccountID(5).Return(1050, nil)

	port := 8080
	server := New(&port)
	server.DB = repo

	req := httptest.NewRequest(http.MethodGet, "/accounts/5", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "5"})
	ctx := jwt.CoupleClaims(req.Context(), &jwt.Claims{CustomerID: 1, AccountID: 5})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	server.GetAccount().ServeHTTP(rr, req)

	res := rr.Result()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.JSONEq(t, `{"id":5,"number":"000000000005","balance":"10.50","transactions":[{"id":"aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee","amount":"10.50","created_at":"2025-01-01T12:00:00Z"}]}`, string(body))
}

func TestGetAccount_NoTransactions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepo(ctrl)
	repo.EXPECT().GetAccountByID(5).Return(&models.Account{ID: 5, Number: "000000000005"}, nil)
	repo.EXPECT().GetTransactionsByAccountID(5).Return(nil, sql.ErrNoRows)
	repo.EXPECT().GetTransactionsBalanceByAccountID(5).Return(0, nil)

	port := 8080
	server := New(&port)
	server.DB = repo

	req := httptest.NewRequest(http.MethodGet, "/accounts/5", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "5"})
	ctx := jwt.CoupleClaims(req.Context(), &jwt.Claims{CustomerID: 1, AccountID: 5})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	server.GetAccount().ServeHTTP(rr, req)

	res := rr.Result()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.JSONEq(t, `{"id":5,"number":"000000000005","balance":"0.00"}`, string(body))
}

func TestGetAccount_InvalidID(t *testing.T) {
	port := 8080
	server := New(&port)

	req := httptest.NewRequest(http.MethodGet, "/accounts/abc", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})
	rr := httptest.NewRecorder()

	server.GetAccount().ServeHTTP(rr, req)

	res := rr.Result()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.JSONEq(t, `{"error":"failed to parse account id"}`, string(body))
}

func TestGetAccount_NoClaims(t *testing.T) {
	port := 8080
	server := New(&port)

	req := httptest.NewRequest(http.MethodGet, "/accounts/5", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "5"})
	rr := httptest.NewRecorder()

	server.GetAccount().ServeHTTP(rr, req)

	res := rr.Result()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.JSONEq(t, `{"error":"failed to parse jwt claims"}`, string(body))
}

func TestGetAccount_Forbidden(t *testing.T) {
	port := 8080
	server := New(&port)

	req := httptest.NewRequest(http.MethodGet, "/accounts/99", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "99"})
	ctx := jwt.CoupleClaims(req.Context(), &jwt.Claims{CustomerID: 1, AccountID: 5})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	server.GetAccount().ServeHTTP(rr, req)

	res := rr.Result()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusForbidden, res.StatusCode)
	assert.JSONEq(t, `{"error":"not authorized to fetch account data"}`, string(body))
}

func TestGetAccount_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepo(ctrl)
	repo.EXPECT().GetAccountByID(5).Return(nil, sql.ErrNoRows)

	port := 8080
	server := New(&port)
	server.DB = repo

	req := httptest.NewRequest(http.MethodGet, "/accounts/5", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "5"})
	ctx := jwt.CoupleClaims(req.Context(), &jwt.Claims{CustomerID: 1, AccountID: 5})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	server.GetAccount().ServeHTTP(rr, req)

	res := rr.Result()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusNotFound, res.StatusCode)
	assert.JSONEq(t, `{"error":"account not found"}`, string(body))
}

func TestGetAccount_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepo(ctrl)
	repo.EXPECT().GetAccountByID(5).Return(nil, errors.New("connection reset"))

	port := 8080
	server := New(&port)
	server.DB = repo

	req := httptest.NewRequest(http.MethodGet, "/accounts/5", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "5"})
	ctx := jwt.CoupleClaims(req.Context(), &jwt.Claims{CustomerID: 1, AccountID: 5})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	server.GetAccount().ServeHTTP(rr, req)

	res := rr.Result()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	assert.JSONEq(t, `{"error":"internal server error"}`, string(body))
}

func TestGetAccount_BalanceDBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepo(ctrl)
	repo.EXPECT().GetAccountByID(5).Return(&models.Account{ID: 5, Number: "000000000005"}, nil)
	repo.EXPECT().GetTransactionsByAccountID(5).Return(nil, sql.ErrNoRows)
	repo.EXPECT().GetTransactionsBalanceByAccountID(5).Return(0, errors.New("db error"))

	port := 8080
	server := New(&port)
	server.DB = repo

	req := httptest.NewRequest(http.MethodGet, "/accounts/5", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "5"})
	ctx := jwt.CoupleClaims(req.Context(), &jwt.Claims{CustomerID: 1, AccountID: 5})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	server.GetAccount().ServeHTTP(rr, req)

	res := rr.Result()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	assert.JSONEq(t, `{"error":"internal server error"}`, string(body))
}

// ---------------------------------------------------------------------------
// PostTransaction
// ---------------------------------------------------------------------------

func TestPostTransaction_Deposit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	txID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	txTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	repo := mocks.NewMockRepo(ctrl)
	repo.EXPECT().CreateTransactionWithBalanceCheck(gomock.Any()).DoAndReturn(func(tx *models.Transaction) error {
		assert.Equal(t, 1050, tx.AmountInCents)
		assert.Equal(t, 5, tx.AccountID)
		tx.ID = txID
		tx.CreatedAt = txTime
		return nil
	})

	port := 8080
	server := New(&port)
	server.DB = repo

	reqBody := `{"type":"deposit","amount":"10.50","account_id":5}`
	req := httptest.NewRequest(http.MethodPost, "/transactions", strings.NewReader(reqBody))
	ctx := jwt.CoupleClaims(req.Context(), &jwt.Claims{CustomerID: 1, AccountID: 5})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	server.PostTransaction().ServeHTTP(rr, req)

	res := rr.Result()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.JSONEq(t, `{"id":"aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee","amount":"10.50","created_at":"2025-01-01T12:00:00Z","account_id":5}`, string(body))
}

func TestPostTransaction_Withdraw(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	txID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	txTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	repo := mocks.NewMockRepo(ctrl)
	repo.EXPECT().CreateTransactionWithBalanceCheck(gomock.Any()).DoAndReturn(func(tx *models.Transaction) error {
		assert.Equal(t, -500, tx.AmountInCents)
		assert.Equal(t, 5, tx.AccountID)
		tx.ID = txID
		tx.CreatedAt = txTime
		return nil
	})

	port := 8080
	server := New(&port)
	server.DB = repo

	reqBody := `{"type":"withdraw","amount":"5.00","account_id":5}`
	req := httptest.NewRequest(http.MethodPost, "/transactions", strings.NewReader(reqBody))
	ctx := jwt.CoupleClaims(req.Context(), &jwt.Claims{CustomerID: 1, AccountID: 5})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	server.PostTransaction().ServeHTTP(rr, req)

	res := rr.Result()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.JSONEq(t, `{"id":"aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee","amount":"-5.00","created_at":"2025-01-01T12:00:00Z","account_id":5}`, string(body))
}

func TestPostTransaction_InvalidJSON(t *testing.T) {
	port := 8080
	server := New(&port)

	req := httptest.NewRequest(http.MethodPost, "/transactions", strings.NewReader(`{bad`))
	ctx := jwt.CoupleClaims(req.Context(), &jwt.Claims{CustomerID: 1, AccountID: 5})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	server.PostTransaction().ServeHTTP(rr, req)

	res := rr.Result()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.JSONEq(t, `{"error":"failed to parse transaction request body"}`, string(body))
}

func TestPostTransaction_ValidationError(t *testing.T) {
	port := 8080
	server := New(&port)

	reqBody := `{"type":"invalid","amount":"10.00","account_id":5}`
	req := httptest.NewRequest(http.MethodPost, "/transactions", strings.NewReader(reqBody))
	ctx := jwt.CoupleClaims(req.Context(), &jwt.Claims{CustomerID: 1, AccountID: 5})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	server.PostTransaction().ServeHTTP(rr, req)

	res := rr.Result()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.JSONEq(t, `{"error":"type must be 'deposit' or 'withdraw'"}`, string(body))
}

func TestPostTransaction_Forbidden(t *testing.T) {
	port := 8080
	server := New(&port)

	reqBody := `{"type":"deposit","amount":"10.00","account_id":99}`
	req := httptest.NewRequest(http.MethodPost, "/transactions", strings.NewReader(reqBody))
	ctx := jwt.CoupleClaims(req.Context(), &jwt.Claims{CustomerID: 1, AccountID: 5})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	server.PostTransaction().ServeHTTP(rr, req)

	res := rr.Result()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusForbidden, res.StatusCode)
	assert.JSONEq(t, `{"error":"not authorized to post transaction"}`, string(body))
}

func TestPostTransaction_InvalidAmount(t *testing.T) {
	port := 8080
	server := New(&port)

	reqBody := `{"type":"deposit","amount":"10.5","account_id":5}`
	req := httptest.NewRequest(http.MethodPost, "/transactions", strings.NewReader(reqBody))
	ctx := jwt.CoupleClaims(req.Context(), &jwt.Claims{CustomerID: 1, AccountID: 5})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	server.PostTransaction().ServeHTTP(rr, req)

	res := rr.Result()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.JSONEq(t, `{"error":"amount must be in format '0.00' (e.g. '10.50')"}`, string(body))
}

func TestPostTransaction_InsufficientFunds(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepo(ctrl)
	repo.EXPECT().CreateTransactionWithBalanceCheck(gomock.Any()).Return(db.ErrInsufficientFunds)

	port := 8080
	server := New(&port)
	server.DB = repo

	reqBody := `{"type":"withdraw","amount":"999.00","account_id":5}`
	req := httptest.NewRequest(http.MethodPost, "/transactions", strings.NewReader(reqBody))
	ctx := jwt.CoupleClaims(req.Context(), &jwt.Claims{CustomerID: 1, AccountID: 5})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	server.PostTransaction().ServeHTTP(rr, req)

	res := rr.Result()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.JSONEq(t, `{"error":"non-sufficient funds"}`, string(body))
}
