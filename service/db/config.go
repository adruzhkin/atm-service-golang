package db

import (
	"fmt"
)

// Environment variables to connect to db.
var (
	dbHost = "localhost"
	dbPort = "15432"
	dbName = "atm"
	dbUser = "postgres"
	dbPass = "mypass"
)

var pgDSN = fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable",
	dbUser, dbPass, dbHost, dbPort, dbName)
