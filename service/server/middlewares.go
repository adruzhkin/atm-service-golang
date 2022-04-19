package server

import (
	"net/http"
	"strings"

	"github.com/adruzhkin/atm-service-golang/service/jwt"
	"github.com/adruzhkin/atm-service-golang/service/utils"
)

func (s *Server) Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		bearerToken := strings.Split(authHeader, " ")

		if len(bearerToken) < 2 {
			utils.RespondWithError(w, http.StatusUnauthorized, "jwt token not provided")
			return
		}

		token := bearerToken[1]
		claims, err := s.JWT.Verify(token)
		if err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
			return
		}

		ctx := jwt.CoupleClaims(r.Context(), claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
