package jwt

type JWT interface {
	Generate(cus int, acc int) (string, error)
	Verify(strToken string) (*Claims, error)
}
