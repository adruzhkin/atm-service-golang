package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/adruzhkin/atm-service-golang/service/db"
	"github.com/adruzhkin/atm-service-golang/service/jwt"
	"github.com/gorilla/mux"
)

type Server struct {
	http.Server
	DB  db.Repo
	JWT jwt.JWT
}

func New(port *int) *Server {
	router := mux.NewRouter().StrictSlash(true)
	api := router.PathPrefix("/api/v1").Subrouter()

	s := &Server{}
	s.Addr = fmt.Sprintf(":%d", *port)
	s.initRoutes(api)

	return s
}

func (s *Server) ShutdownGracefully(timeout *time.Duration) {
	log.Printf("graceful shutdown with %v timeout\n", timeout)

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		log.Fatalf("unexpected shuttdown: %s\n", err)
	}

	log.Println("server shutdown gracefully")
}

func (s *Server) initRoutes(r *mux.Router) {
	r.HandleFunc("/health", s.CheckHealth()).Methods(http.MethodGet)
	r.HandleFunc("/auth/signup", s.SignupCustomer()).Methods(http.MethodPost)
	r.HandleFunc("/auth/login", s.LoginCustomer()).Methods(http.MethodPost)

	r.HandleFunc("/accounts/{id:[0-9]+}", s.Authenticate(s.GetAccount())).Methods(http.MethodGet)
	r.HandleFunc("/transactions", s.Authenticate(s.PostTransaction())).Methods(http.MethodPost)

	s.Handler = r
}
