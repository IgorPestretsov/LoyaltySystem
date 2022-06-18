package main

import (
	"flag"
	"github.com/IgorPestretsov/LoyaltySystem/internal/handlers"
	"github.com/IgorPestretsov/LoyaltySystem/internal/middlewares"
	"github.com/IgorPestretsov/LoyaltySystem/internal/sqlStorage"
	"github.com/IgorPestretsov/LoyaltySystem/internal/storage"
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
)

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS" envDefault:"localhost:8080"`
	DatabaseURI          string `env:"DATABASE_URI" envDefault:"password=P@ssw0rd dbname=loyaltySystem sslmode=disable host=localhost port=5432 user=user "`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS" envDefault:"http://localhost:8080"`
}

func main() {
	var cfg Config
	var s storage.Storage
	parseFlags(&cfg)

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	s = sqlStorage.NewSQLStorage(cfg.DatabaseURI)

	r := chi.NewRouter()
	r.Use(middleware.Compress(5))
	r.Use(middlewares.Decompress)

	r.Post("/api/user/register", func(rw http.ResponseWriter, r *http.Request) {
		handlers.RegisterUser(rw, r, s)
	})
	log.Fatal(http.ListenAndServe(cfg.RunAddress, r))
}

func parseFlags(config *Config) {
	flag.StringVar(&config.RunAddress, "a", config.RunAddress, "IP address and port on which service will run")
	flag.StringVar(&config.DatabaseURI, "d", config.DatabaseURI, "Service database URI")
	flag.StringVar(&config.AccrualSystemAddress, "r", config.AccrualSystemAddress, "Accrual system connection address")
}
