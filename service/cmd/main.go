package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/adruzhkin/atm-service-golang/service/db"
	"github.com/adruzhkin/atm-service-golang/service/jwt"
	"github.com/adruzhkin/atm-service-golang/service/server"
)

var (
	jwtSecret = os.Getenv("JWT_SECRET")
	port      = flag.Int("port", 5000, "http port to listen on")
	timeout   = flag.Duration("timeout", 5*time.Second, "timeout for graceful shutdown")
)

func main() {
	flag.Parse()

	s := server.New(port)
	s.JWT = &jwt.Token{Secret: jwtSecret}
	s.DB = &db.Postgres{}

	for i := 0; i < 3; i++ {
		// Pause before connection to wait for db docker container build
		time.Sleep(5 * time.Second)
		err := s.DB.Open()
		if err != nil {
			log.Printf("cannot open db connection: %s\n", err)
			continue
		}

		break
	}
	defer s.DB.Close()

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Println("server is up and running")
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("unexpected shuttdown: %s\n", err)
		}
	}()

	<-stopCh
	s.ShutdownGracefully(timeout)
}
