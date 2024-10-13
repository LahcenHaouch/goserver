package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/LahcenHaouch/goserver/api"
	"github.com/LahcenHaouch/goserver/internal/database"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

func init() {
	godotenv.Load()
}

func main() {
	dbURL := os.Getenv("DB_URL")
	tokenSecret := os.Getenv("TOKEN_SECRET")
	db, err := sql.Open("postgres", dbURL)

	if err != nil {
		fmt.Println("Error connecting to db: %w", err)
		return
	}

	dbQueries := database.New(db)
	api := api.ApiConfig{FileServerHits: 0, Database: dbQueries, TokenSecret: tokenSecret}

	mux := http.NewServeMux()
	serv := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	mux.Handle("/app/", api.MiddlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", api.HealthzHandler)
	mux.HandleFunc("/admin/reset", api.ResetHandler)
	mux.HandleFunc("GET /api/chirps", api.HandleGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpId}", api.HandleGetChirp)
	mux.HandleFunc("POST /api/chirps", api.HandleCreateChirp)
	mux.HandleFunc("POST /api/users", api.HandleCreateUser)
	mux.HandleFunc("POST /api/login", api.HandleLogin)
	mux.HandleFunc("GET /admin/metrics", api.CountHandler)

	fmt.Println("Listening on port:8080")
	if err := serv.ListenAndServe(); err != nil {
		fmt.Println("Error 500: %w", err)
	}
}
