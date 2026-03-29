package jwt

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type ctxKey int

const (
	_ ctxKey = iota
	ctxKeyClaims
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

type Token struct {
	Secret             string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
}

type Claims struct {
	CustomerID int    `json:"cus"`
	AccountID  int    `json:"acc"`
	TokenType  string `json:"typ"`
	jwt.RegisteredClaims
}

func (tkn *Token) GenerateTokenPair(cus int, acc int) (string, string, error) {
	accessToken, err := tkn.generateToken(cus, acc, TokenTypeAccess, tkn.AccessTokenExpiry)
	if err != nil {
		return "", "", fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := tkn.generateToken(cus, acc, TokenTypeRefresh, tkn.RefreshTokenExpiry)
	if err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (tkn *Token) generateToken(cus int, acc int, tokenType string, expiry time.Duration) (string, error) {
	if tkn.Secret == "" {
		return "", errors.New("failed to verify jwt secret environment variable")
	}

	key := []byte(tkn.Secret)

	claims := Claims{
		CustomerID: cus,
		AccountID:  acc,
		TokenType:  tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: &jwt.NumericDate{Time: time.Now().Add(expiry)},
		},
	}

	unsignedToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := unsignedToken.SignedString(key)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func (tkn *Token) VerifyAccessToken(strToken string) (*Claims, error) {
	return tkn.verifyToken(strToken, TokenTypeAccess)
}

func (tkn *Token) VerifyRefreshToken(strToken string) (*Claims, error) {
	return tkn.verifyToken(strToken, TokenTypeRefresh)
}

func (tkn *Token) verifyToken(strToken string, expectedType string) (*Claims, error) {
	if tkn.Secret == "" {
		return nil, errors.New("failed to verify jwt secret environment variable")
	}

	key := []byte(tkn.Secret)
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(strToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return key, nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid jwt token")
	}

	if claims.TokenType != expectedType {
		return nil, fmt.Errorf("invalid token type: expected %s, got %s", expectedType, claims.TokenType)
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
