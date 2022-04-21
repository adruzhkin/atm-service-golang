package server

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/adruzhkin/atm-service-golang/service/jwt"
	"github.com/adruzhkin/atm-service-golang/service/models"
	"github.com/adruzhkin/atm-service-golang/service/utils"
	"github.com/gorilla/mux"
)

func (s *Server) CheckHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := s.DB.Ping()
		if err != nil {
			utils.RespondWithStatus(w, http.StatusServiceUnavailable, "service unavailable")
			return
		}

		utils.RespondWithStatus(w, http.StatusOK, "service is up and running")
	}
}

func (s *Server) SignupCustomer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse request body.
		cusReq := models.CustomerRequestBody{}
		err := json.NewDecoder(r.Body).Decode(&cusReq)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "failed to parse customer request body")
			return
		}

		// Save in db.
		cus := models.Customer{
			FirstName: cusReq.FirstName,
			LastName:  cusReq.LastName,
			Email:     cusReq.Email,
			PINHash:   utils.GeneratePINHash(cusReq.PINNumber),
			Account:   &models.Account{Number: cusReq.AccountNumber},
		}
		err = s.DB.CreateCustomer(&cus)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		token, err := s.JWT.Generate(cus.ID, cus.AccountID)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		}

		cus.OmitValues()
		cusVerified := models.CustomerVerified{
			JWT:      token,
			Customer: &cus,
		}

		utils.RespondWithJSON(w, http.StatusOK, cusVerified)
	}
}

func (s *Server) LoginCustomer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse request body.
		credentials := models.CustomerCredentials{}
		err := json.NewDecoder(r.Body).Decode(&credentials)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "failed to parse customer credentials")
			return
		}

		// Query db.
		cus, err := s.DB.GetCustomerByCredentials(&credentials)
		if err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
			return
		}

		token, err := s.JWT.Generate(cus.ID, cus.AccountID)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		}

		cus.OmitValues()
		cusVerified := models.CustomerVerified{
			JWT:      token,
			Customer: cus,
		}

		utils.RespondWithJSON(w, http.StatusOK, cusVerified)
	}
}

func (s *Server) GetAccount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse account id from vars.
		vars := mux.Vars(r)
		idVar, err := strconv.Atoi(vars["id"])
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "failed to parse account id")
			return
		}

		// Parse account id from token.
		claims, err := jwt.DecoupleClaims(r.Context())
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		id := claims.AccountID

		if idVar != id {
			utils.RespondWithError(w, http.StatusForbidden, "not authorized to fetch account data")
			return
		}

		// Query db for account.
		acc, err := s.DB.GetAccountByID(id)
		if err != nil {
			switch err {
			case sql.ErrNoRows:
				utils.RespondWithError(w, http.StatusNotFound, "account not found")
				return
			default:
				utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
				return
			}
		}

		// Query db for transactions
		txs, err := s.DB.GetTransactionsByAccountID(id)
		if err != nil {
			switch err {
			case sql.ErrNoRows:
				acc.Transactions = []models.Transaction{}
			default:
				utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
				return
			}
		} else {
			acc.Transactions = txs
		}

		// Query db for balance.
		balance, err := s.DB.GetTransactionsBalanceByAccountID(id)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		balanceTx := models.Transaction{AmountInCents: balance}
		strBalance := balanceTx.ParseToAmount()
		acc.Balance = strBalance

		// Omit unnecessary values
		for i, tx := range txs {
			txs[i].Amount = tx.ParseToAmount()
			txs[i].OmitAmountInCents()
			txs[i].OmitAccountID()
		}

		utils.RespondWithJSON(w, http.StatusOK, acc)
	}
}

func (s *Server) PostTransaction() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse request body.
		trbReq := models.TransactionRequestBody{}
		err := json.NewDecoder(r.Body).Decode(&trbReq)
		if err != nil {
			log.Println(err)
			utils.RespondWithError(w, http.StatusBadRequest, "failed to parse transaction request body")
			return
		}

		// Parse account id from token.
		claims, err := jwt.DecoupleClaims(r.Context())
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		id := claims.AccountID

		if trbReq.AccountID != id {
			utils.RespondWithError(w, http.StatusForbidden, "not authorized to post transaction")
			return
		}

		// Query db for balance.
		balance, err := s.DB.GetTransactionsBalanceByAccountID(id)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Verify account funds
		amountInCents, err := trbReq.ParseToAmountInCents()
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		hasFunds := trbReq.HasSufficientFunds(balance, amountInCents)

		if !hasFunds {
			utils.RespondWithError(w, http.StatusBadRequest, "non-sufficient funds")
			return
		}

		// Save in db.
		tx := models.Transaction{
			AmountInCents: amountInCents,
			AccountID:     trbReq.AccountID,
		}
		err = s.DB.CreateTransaction(&tx)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		tx.Amount = tx.ParseToAmount()
		tx.OmitAmountInCents()

		utils.RespondWithJSON(w, http.StatusOK, tx)
	}
}
