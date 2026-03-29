package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/adruzhkin/atm-service-golang/service/db"
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

		// Validate input.
		if err := cusReq.Validate(); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Hash PIN.
		pinHash, err := models.GeneratePINHash(cusReq.PINNumber)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "failed to hash pin")
			return
		}

		// Save in db.
		cus := models.Customer{
			FirstName: cusReq.FirstName,
			LastName:  cusReq.LastName,
			Email:     cusReq.Email,
			PINHash:   pinHash,
			Account:   &models.Account{},
		}
		err = s.DB.CreateCustomer(&cus)
		if err != nil {
			if db.IsDuplicateKeyError(err) {
				utils.RespondWithError(w, http.StatusConflict, "email already registered")
				return
			}
			slog.Error("create customer failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		accessToken, refreshToken, err := s.JWT.GenerateTokenPair(cus.ID, cus.AccountID)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		cus.OmitValues()
		cusVerified := models.CustomerVerified{
			JWT:          accessToken,
			RefreshToken: refreshToken,
			Customer:     &cus,
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

		// Validate input.
		if err := credentials.Validate(); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Query db.
		cus, err := s.DB.GetCustomerByCredentials(&credentials)
		if err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
			return
		}

		accessToken, refreshToken, err := s.JWT.GenerateTokenPair(cus.ID, cus.AccountID)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		cus.OmitValues()
		cusVerified := models.CustomerVerified{
			JWT:          accessToken,
			RefreshToken: refreshToken,
			Customer:     cus,
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
				slog.Error("get account failed", "error", err, "account_id", id)
				utils.RespondWithError(w, http.StatusInternalServerError, "internal server error")
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
				slog.Error("get transactions failed", "error", err, "account_id", id)
				utils.RespondWithError(w, http.StatusInternalServerError, "internal server error")
				return
			}
		} else {
			acc.Transactions = txs
		}

		// Query db for balance.
		balance, err := s.DB.GetTransactionsBalanceByAccountID(id)
		if err != nil {
			slog.Error("get balance failed", "error", err, "account_id", id)
			utils.RespondWithError(w, http.StatusInternalServerError, "internal server error")
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
			slog.Error("failed to decode transaction request", "error", err)
			utils.RespondWithError(w, http.StatusBadRequest, "failed to parse transaction request body")
			return
		}

		// Validate input.
		if err := trbReq.Validate(); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
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

		// Parse amount.
		amountInCents, err := trbReq.ParseToAmountInCents()
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Save in db (atomic balance check + insert).
		tx := models.Transaction{
			AmountInCents: amountInCents,
			AccountID:     trbReq.AccountID,
		}
		err = s.DB.CreateTransactionWithBalanceCheck(&tx)
		if err != nil {
			if errors.Is(err, db.ErrInsufficientFunds) {
				utils.RespondWithError(w, http.StatusBadRequest, err.Error())
				return
			}
			slog.Error("create transaction failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		tx.Amount = tx.ParseToAmount()
		tx.OmitAmountInCents()

		utils.RespondWithJSON(w, http.StatusOK, tx)
	}
}

func (s *Server) RefreshToken() http.HandlerFunc {
	type refreshRequest struct {
		RefreshToken string `json:"refresh_token"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req refreshRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "failed to parse refresh token request")
			return
		}

		if req.RefreshToken == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "refresh_token is required")
			return
		}

		claims, err := s.JWT.VerifyRefreshToken(req.RefreshToken)
		if err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "invalid or expired refresh token")
			return
		}

		accessToken, refreshToken, err := s.JWT.GenerateTokenPair(claims.CustomerID, claims.AccountID)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, models.TokenPair{
			JWT:          accessToken,
			RefreshToken: refreshToken,
		})
	}
}
