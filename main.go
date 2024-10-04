package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/LahcenHaouch/goserver/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

var id int = 1

type apiConfig struct {
	fileServerHits int
	database       *database.Queries
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		cfg.fileServerHits++
		fmt.Println("count: ", cfg.fileServerHits)
		next.ServeHTTP(res, req)
	})
}
func (cfg *apiConfig) countHandler(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(200)
	res.Header().Set("Content-Type", "text/html; charset=utf-8")
	res.Write([]byte(fmt.Sprintf(`
		<html>
			<body>
    			<h1>Welcome, Chirpy Admin</h1>
       			<p>Chirpy has been visited %d times!</p>
          	</body>
        </html>`, cfg.fileServerHits)))
}
func (cfg *apiConfig) resetHandler(res http.ResponseWriter, req *http.Request) {
	cfg.fileServerHits = 0
	res.WriteHeader(200)
	res.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileServerHits)))
}

type CreateUser struct {
	Email string `json:"email"`
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (c *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var rUser CreateUser
	if err := decoder.Decode(&rUser); err != nil {
		respondWithError(w, map[string]string{"error": "Invalid request"}, 400)
		return
	}

	user, err := c.database.CreateUser(r.Context(), sql.NullString{String: rUser.Email, Valid: true})
	if err != nil {
		respondWithError(w, map[string]string{"error": "Error creating user"}, 500)
		return
	}

	newUser := User{
		ID:        user.ID.UUID,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
		Email:     user.Email.String,
	}

	body, err := json.Marshal(newUser)
	if err != nil {
		respondWithError(w, map[string]string{"error": "Error marshalling body"}, 500)
	}

	w.WriteHeader(201)
	w.Write(body)
}

func healthzHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(200)
	res.Write([]byte("OK"))
}

func respondWithError(w http.ResponseWriter, body map[string]string, status int) {
	data, err := json.Marshal(body)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(data)
}

func removeBadWords(body string) string {
	words := strings.Split(body, " ")
	result := make([]string, 0)

	for _, word := range words {
		if isBadWord(word) {
			result = append(result, "****")
		} else {
			result = append(result, word)
		}
	}

	return strings.Join(result, " ")
}

func isBadWord(word string) bool {
	switch strings.ToLower(word) {
	case "kerfuffle", "sharbert", "fornax":
		return true
	default:
		return false
	}
}

type PostChirp struct {
	Body string `json:"body"`
}

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

func handleGetChirps(w http.ResponseWriter, r *http.Request) {
	f, err := os.ReadFile("./database.json")
	if err != nil {
		respondWithError(w, map[string]string{"body": "Error opening database.json"}, 500)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(f)
}

func handleCreateChirp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var chirp PostChirp
	error := map[string]string{"error": "Something went wrong"}
	if err := decoder.Decode(&chirp); err != nil {
		fmt.Println("error decoding")
		respondWithError(w, error, 400)
		return
	}

	if len(chirp.Body) > 140 {
		respondWithError(w, error, 400)
		return
	}

	if _, err := os.Stat("./database.json"); errors.Is(err, os.ErrNotExist) {
		if _, err := os.Create("./database.json"); err != nil {
			fmt.Println("error creating")
			respondWithError(w, error, 500)
			return
		}
	} else if err != nil {
		fmt.Println("error else if")
		respondWithError(w, error, 500)
		return
	}

	f, err := os.ReadFile("./database.json")
	if err != nil {
		fmt.Println("error reading")
		respondWithError(w, error, 500)
		return
	}

	var data []Chirp

	if len(f) > 0 {
		if err := json.Unmarshal(f, &data); err != nil {
			fmt.Println("error json unmarshal")
			log.Fatal(err)
			respondWithError(w, error, 500)
			return
		}
	}

	newId := id
	if len(data) > 0 {
		newId = data[len(data)-1].Id + 1
	}

	newChirp := Chirp{Id: newId, Body: removeBadWords(chirp.Body)}

	data = append(data, newChirp)

	dt, err := json.Marshal(data)
	if err != nil {
		respondWithError(w, error, 500)
		return
	}
	if err := os.WriteFile("./database.json", dt, 0644); err != nil {
		respondWithError(w, error, 500)
		return
	}

	newBody, err := json.Marshal(newChirp)
	if err != nil {
		respondWithError(w, error, 500)
	}

	w.WriteHeader(201)
	w.Write(newBody)
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)

	if err != nil {
		fmt.Println("Error connecting to db: %v", err)
		return
	}

	dbQueries := database.New(db)
	api := apiConfig{fileServerHits: 0, database: dbQueries}

	mux := http.NewServeMux()
	serv := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	mux.Handle("/app/", api.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", healthzHandler)
	mux.HandleFunc("/admin/reset", api.resetHandler)
	mux.HandleFunc("GET /api/chirps", handleGetChirps)
	mux.HandleFunc("POST /api/chirps", handleCreateChirp)
	mux.HandleFunc("POST /api/users", api.handleCreateUser)
	mux.HandleFunc("GET /admin/metrics", api.countHandler)

	fmt.Println("Listening on port:8080")
	if err := serv.ListenAndServe(); err != nil {
		fmt.Println("Error 500: %w", err)
	}
}
