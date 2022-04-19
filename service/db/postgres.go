package db

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)

const (
	maxOpenConn = 15
	maxIdleConn = 10
	maxLifetime = 5 * time.Minute
)

type Postgres struct {
	db *sql.DB
}

func (p *Postgres) Open() error {
	pg, err := sql.Open("postgres", pgDSN)
	if err != nil {
		return err
	}

	pg.SetMaxOpenConns(maxOpenConn)
	pg.SetMaxIdleConns(maxIdleConn)
	pg.SetConnMaxLifetime(maxLifetime)

	if err = pg.Ping(); err != nil {
		return err
	}

	log.Println("connected to database")
	p.db = pg

	return nil
}

func (p *Postgres) Close() error {
	return p.db.Close()
}

func (p *Postgres) Ping() error {
	return p.db.Ping()
}
