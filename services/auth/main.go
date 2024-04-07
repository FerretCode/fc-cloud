package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/ferretcode-freelancing/fc-cloud/services/auth/routes"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("DSN")

	if dsn == "" {
		log.Fatal(errors.New("the dsn environment variable was not found"))
	}

	dsn = strings.ReplaceAll(dsn, "\n", "")

	dsn = strings.Replace(dsn, "POSTGRES_SERVICE_HOST", os.Getenv("POSTGRES_SERVICE_HOST"), 1)
	dsn = strings.Replace(dsn, "POSTGRES_SERVICE_PORT", os.Getenv("POSTGRES_SERVICE_PORT"), 1)

	fmt.Println(dsn)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal(err)
	}

	db.AutoMigrate(&routes.CloudUser{})

	host := os.Getenv("FC_SESSION_CACHE_SERVICE_HOST")
	port := os.Getenv("FC_SESSION_CACHE_SERVICE_PORT")

	rdb := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: "",
		DB:       0,
	})

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Get("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		err := routes.Login(w, r, *db, *rdb)

		if err != nil {
			fmt.Println(err)

			http.Error(w, "There was an error authenticating you. Please try again later.", http.StatusInternalServerError)
		}
	})

	r.Get("/auth/callback", func(w http.ResponseWriter, r *http.Request) {
		err := routes.Callback(w, r, *db, *rdb)

		if err != nil {
			fmt.Println(err)

			http.Error(w, "There was an error authenticating you. Please try again later.", http.StatusInternalServerError)
		}
	})

	http.ListenAndServe(":3000", r)
}
