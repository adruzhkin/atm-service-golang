package server

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/adruzhkin/atm-service-golang/service/jwt"
	"github.com/adruzhkin/atm-service-golang/service/mocks"
	"github.com/adruzhkin/atm-service-golang/service/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCheckHealth(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepo(ctrl)
	repo.EXPECT().Ping().Return(nil)

	port := 5000
	server := New(&port)
	server.DB = repo

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	handler := server.CheckHealth()
	handler.ServeHTTP(rr, req)

	res := rr.Result()
	resBody, _ := ioutil.ReadAll(res.Body)
	expBody := `{"status":"service is up and running"}`

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.JSONEq(t, expBody, string(resBody))
}

func TestSignupCustomer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	accAfter := models.Account{
		ID:     0,
		Number: "100000000099",
	}

	cusAfter := models.Customer{
		ID:        7,
		FirstName: "Natasha",
		LastName:  "Romanov",
		Email:     "natasha@gmail.com",
		Account:   &accAfter,
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

	port := 5000
	server := New(&port)
	server.DB = repo
	server.JWT = jwtMock

	reqBody := `{"first_name": "Natasha", "last_name": "Romanov", "email": "natasha@gmail.com", "pin_number": "1234", "account_number": "100000000099"}`
	reader := strings.NewReader(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/signup", reader)
	rr := httptest.NewRecorder()

	handler := server.SignupCustomer()
	handler.ServeHTTP(rr, req)

	res := rr.Result()
	resBody, _ := ioutil.ReadAll(res.Body)
	expBody := `{"jwt": "access_token", "refresh_token": "refresh_token", "customer": {"id": 7, "first_name": "Natasha", "last_name": "Romanov", "email": "natasha@gmail.com", "account": {"id": 0, "number": "100000000099"}}}`

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.JSONEq(t, expBody, string(resBody))
}

func TestLoginCustomer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	credentials := models.CustomerCredentials{
		PINNumber: "1234",
		Email:     "natasha@gmail.com",
	}

	acc := models.Account{
		Number: "100000000099",
	}

	cus := models.Customer{
		ID:        7,
		FirstName: "Natasha",
		LastName:  "Romanov",
		Email:     "natasha@gmail.com",
		Account:   &acc,
	}

	repo := mocks.NewMockRepo(ctrl)
	repo.EXPECT().GetCustomerByCredentials(&credentials).Return(&cus, nil)

	jwtMock := mocks.NewMockJWT(ctrl)
	jwtMock.EXPECT().GenerateTokenPair(7, 0).Return("access_token", "refresh_token", nil)

	port := 5000
	server := New(&port)
	server.DB = repo
	server.JWT = jwtMock

	reqBody := `{"pin_number": "1234", "email": "natasha@gmail.com"}`
	reader := strings.NewReader(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", reader)
	rr := httptest.NewRecorder()

	handler := server.LoginCustomer()
	handler.ServeHTTP(rr, req)

	res := rr.Result()
	resBody, _ := ioutil.ReadAll(res.Body)
	expBody := `{"jwt": "access_token", "refresh_token": "refresh_token", "customer": {"id": 7, "first_name": "Natasha", "last_name": "Romanov", "email": "natasha@gmail.com", "account": {"id": 0, "number": "100000000099"}}}`

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.JSONEq(t, expBody, string(resBody))
}

func TestRefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jwtMock := mocks.NewMockJWT(ctrl)
	jwtMock.EXPECT().VerifyRefreshToken("old_refresh").Return(&jwt.Claims{
		CustomerID: 7,
		AccountID:  3,
		TokenType:  "refresh",
	}, nil)
	jwtMock.EXPECT().GenerateTokenPair(7, 3).Return("new_access", "new_refresh", nil)

	port := 5000
	server := New(&port)
	server.JWT = jwtMock

	reqBody := `{"refresh_token": "old_refresh"}`
	reader := strings.NewReader(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", reader)
	rr := httptest.NewRecorder()

	handler := server.RefreshToken()
	handler.ServeHTTP(rr, req)

	res := rr.Result()
	resBody, _ := ioutil.ReadAll(res.Body)
	expBody := `{"jwt": "new_access", "refresh_token": "new_refresh"}`

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.JSONEq(t, expBody, string(resBody))
}
