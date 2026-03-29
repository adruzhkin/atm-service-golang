package jwt

type JWT interface {
	GenerateTokenPair(cus int, acc int) (string, string, error)
	VerifyAccessToken(strToken string) (*Claims, error)
	VerifyRefreshToken(strToken string) (*Claims, error)
}
