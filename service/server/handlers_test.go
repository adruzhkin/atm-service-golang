package server

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/adruzhkin/atm-service-golang/service/mocks"
	"github.com/adruzhkin/atm-service-golang/service/models"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
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

	accBefore := models.Account{
		Number: "100000000099",
	}

	cusBefore := models.Customer{
		FirstName: "Natasha",
		LastName:  "Romanov",
		Email:     "natasha@gmail.com",
		PINHash:   "03ac674216f3e15c761ee1a5e255f067953623c8b388b4459e13f978d7c846f4",
		Account:   &accBefore,
	}

	accAfter := models.Account{
		ID:     0,
		Number: "100000000099",
	}

	cusAfter := models.Customer{
		ID:        7,
		FirstName: "Natasha",
		LastName:  "Romanov",
		Email:     "natasha@gmail.com",
		PINHash:   "03ac674216f3e15c761ee1a5e255f067953623c8b388b4459e13f978d7c846f4",
		Account:   &accAfter,
	}

	repo := mocks.NewMockRepo(ctrl)
	repo.EXPECT().CreateCustomer(&cusBefore).SetArg(0, cusAfter).Return(nil)

	jwt := mocks.NewMockJWT(ctrl)
	jwt.EXPECT().Generate(7, 0).Return("token", nil)

	port := 5000
	server := New(&port)
	server.DB = repo
	server.JWT = jwt

	reqBody := `{"first_name": "Natasha", "last_name": "Romanov", "email": "natasha@gmail.com", "pin_number": "1234", "account_number": "100000000099"}`
	reader := strings.NewReader(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/signup", reader)
	rr := httptest.NewRecorder()

	handler := server.SignupCustomer()
	handler.ServeHTTP(rr, req)

	res := rr.Result()
	resBody, _ := ioutil.ReadAll(res.Body)
	expBody := `{"jwt": "token", "customer": {"id": 7, "first_name": "Natasha", "last_name": "Romanov", "email": "natasha@gmail.com", "account": {"id": 0, "number": "100000000099"}}}`

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.JSONEq(t, expBody, string(resBody))
}

func TestLoginCustomer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	credentials := models.CustomerCredentials{
		PINNumber:     "1234",
		AccountNumber: "100000000099",
	}

	acc := models.Account{
		Number: "100000000099",
	}

	cus := models.Customer{
		ID:        7,
		FirstName: "Natasha",
		LastName:  "Romanov",
		Email:     "natasha@gmail.com",
		PINHash:   "03ac674216f3e15c761ee1a5e255f067953623c8b388b4459e13f978d7c846f4",
		Account:   &acc,
	}

	repo := mocks.NewMockRepo(ctrl)
	repo.EXPECT().GetCustomerByCredentials(&credentials).Return(&cus, nil)

	jwt := mocks.NewMockJWT(ctrl)
	jwt.EXPECT().Generate(7, 0).Return("token", nil)

	port := 5000
	server := New(&port)
	server.DB = repo
	server.JWT = jwt

	reqBody := `{"pin_number": "1234", "account_number": "100000000099"}`
	reader := strings.NewReader(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", reader)
	rr := httptest.NewRecorder()

	handler := server.LoginCustomer()
	handler.ServeHTTP(rr, req)

	res := rr.Result()
	resBody, _ := ioutil.ReadAll(res.Body)
	expBody := `{"jwt": "token", "customer": {"id": 7, "first_name": "Natasha", "last_name": "Romanov", "email": "natasha@gmail.com", "account": {"id": 0, "number": "100000000099"}}}`

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.JSONEq(t, expBody, string(resBody))
}
