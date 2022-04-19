package jwt

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type ctxKey int

const (
	_ ctxKey = iota
	ctxKeyClaims
)

type Token struct {
	Secret string
}

type Claims struct {
	CustomerID int `json:"cus"`
	AccountID  int `json:"acc"`
	jwt.RegisteredClaims
}

func (tkn *Token) Generate(cus int, acc int) (string, error) {
	if tkn.Secret == "" {
		return "", errors.New("failed to verify jwt secret environment variable")
	}

	key := []byte(tkn.Secret)
	expiry := time.Now().Add(2 * time.Minute)

	claims := Claims{
		CustomerID: cus,
		AccountID:  acc,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: &jwt.NumericDate{Time: expiry},
		},
	}

	unsignedToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := unsignedToken.SignedString(key)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func (tkn *Token) Verify(strToken string) (*Claims, error) {
	if tkn.Secret == "" {
		return nil, errors.New("failed to verify jwt secret environment variable")
	}

	key := []byte(tkn.Secret)
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(strToken, claims, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid jwt token")
	}

	return claims, nil
}

func CoupleClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, ctxKeyClaims, claims)
}

func DecoupleClaims(ctx context.Context) (*Claims, error) {
	if claims, ok := ctx.Value(ctxKeyClaims).(*Claims); ok {
		return claims, nil
	}
	return nil, errors.New("failed to parse jwt claims")
}
