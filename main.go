package main

import (
	"database/sql"
	"log"
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
	polkaKey := os.Getenv("POLKA_KEY")
	db, err := sql.Open("postgres", dbURL)

	if err != nil {
		log.Printf("error connecting to db: %q", err)
		return
	}

	dbQueries := database.New(db)
	api := api.ApiConfig{FileServerHits: 0, Database: dbQueries, TokenSecret: tokenSecret, PolkaKey: polkaKey}

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
	mux.HandleFunc("PUT /api/users", api.HandleUpdateUser)
	mux.HandleFunc("POST /api/login", api.HandleLogin)
	mux.HandleFunc("POST /api/refresh", api.HandleRefresh)
	mux.HandleFunc("POST /api/revoke", api.HandleRevoke)
	mux.HandleFunc("POST /api/polka/webhooks", api.HandleWebHook)
	mux.HandleFunc("GET /admin/metrics", api.CountHandler)
	mux.HandleFunc("DELETE /api/chirps/{chirpId}", api.HandleDeleteChirp)

	log.Println("listening on port:", serv.Addr[1:])
	if err := serv.ListenAndServe(); err != nil {
		log.Printf("error starting server: %q", err)
	}
}
