package main

import (
	"flag"
	"github.com/IgorPestretsov/LoyaltySystem/internal/handlers"
	"github.com/IgorPestretsov/LoyaltySystem/internal/middlewares"
	"github.com/IgorPestretsov/LoyaltySystem/internal/orderBroker"
	"github.com/IgorPestretsov/LoyaltySystem/internal/sqlStorage"
	"github.com/IgorPestretsov/LoyaltySystem/internal/storage"
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"log"
	"net/http"
)

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS" envDefault:"localhost:8081"`
	DatabaseURI          string `env:"DATABASE_URI" envDefault:"password=P@ssw0rd dbname=loyaltySystem sslmode=disable host=localhost port=5432 user=user "`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS" envDefault:"http://localhost:8080"`
	TokenSecret          string `env:"TOKEN_SECRET" envDefault:"SuperSecret"`
}

var cfg Config
var s storage.Storage
var tokenAuth *jwtauth.JWTAuth

func main() {
	parseFlags(&cfg)

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	tokenAuth = jwtauth.New("HS256", []byte(cfg.TokenSecret), nil)
	s = sqlStorage.NewSQLStorage(cfg.DatabaseURI)
	b := orderBroker.New(s, cfg.AccrualSystemAddress)
	b.Start()
	log.Fatal(http.ListenAndServe(cfg.RunAddress, router()))
}

func parseFlags(config *Config) {
	flag.StringVar(&config.RunAddress, "a", config.RunAddress, "IP address and port on which service will run")
	flag.StringVar(&config.DatabaseURI, "d", config.DatabaseURI, "Service database URI")
	flag.StringVar(&config.AccrualSystemAddress, "r", config.AccrualSystemAddress, "Accrual system connection address")
	flag.StringVar(&config.TokenSecret, "s", config.TokenSecret, "Secret for hashing tokens")
}
func router() http.Handler {
	r := chi.NewRouter()
	//Protected group
	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(jwtauth.Authenticator)
		r.Use(middleware.Compress(5))
		r.Use(middlewares.Decompress)
		r.Post("/api/user/orders", func(rw http.ResponseWriter, r *http.Request) {
			handlers.SaveOrder(rw, r, s)
		})
		r.Get("/api/user/orders", func(rw http.ResponseWriter, r *http.Request) {
			handlers.GetAllUserOrders(rw, r, s)
		})
		r.Get("/api/user/balance", func(rw http.ResponseWriter, r *http.Request) {
			handlers.GetBalance(rw, r, s)
		})
		r.Get("/api/user/withdrawals", func(rw http.ResponseWriter, r *http.Request) {
			handlers.GeWithdrawals(rw, r, s)
		})

		r.Post("/api/user/balance/withdraw", func(rw http.ResponseWriter, r *http.Request) {
			handlers.Withdraw(rw, r, s)
		})
	})
	//Unprotected group
	r.Group(func(r chi.Router) {
		r.Post("/api/user/register", func(rw http.ResponseWriter, r *http.Request) {
			handlers.RegisterUser(rw, r, s, tokenAuth)
		})
		r.Post("/api/user/login", func(rw http.ResponseWriter, r *http.Request) {
			handlers.Login(rw, r, s, tokenAuth)
		})
	})
	return r
}
