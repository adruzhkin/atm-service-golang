package db

import (
	"fmt"
	"os"
)

// Environment variables to connect to db.
var (
	dbHost    = os.Getenv("DB_HOST")
	dbPort    = os.Getenv("DB_PORT")
	dbName    = os.Getenv("DB_NAME")
	dbUser    = os.Getenv("DB_USER")
	dbPass    = os.Getenv("DB_PASS")
	dbSSLMode = os.Getenv("DB_SSLMODE")
)

var pgDSN string

func init() {
	if dbSSLMode == "" {
		dbSSLMode = "disable"
	}
	pgDSN = fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=%v",
		dbUser, dbPass, dbHost, dbPort, dbName, dbSSLMode)
}
