package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/LahcenHaouch/goserver/internal/database"
	"github.com/LahcenHaouch/goserver/utils"
	"github.com/google/uuid"
)

type ApiConfig struct {
	FileServerHits int
	Database       *database.Queries
}

func (a ApiConfig) HealthzHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(200)
	res.Write([]byte("OK"))
}

func (cfg *ApiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		cfg.FileServerHits++
		fmt.Println("count: ", cfg.FileServerHits)
		next.ServeHTTP(res, req)
	})
}
func (cfg *ApiConfig) CountHandler(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(200)
	res.Header().Set("Content-Type", "text/html; charset=utf-8")
	res.Write([]byte(fmt.Sprintf(`
		<html>
			<body>
    			<h1>Welcome, Chirpy Admin</h1>
       			<p>Chirpy has been visited %d times!</p>
          	</body>
        </html>`, cfg.FileServerHits)))
}
func (cfg *ApiConfig) ResetHandler(res http.ResponseWriter, req *http.Request) {
	cfg.FileServerHits = 0
	res.WriteHeader(200)
	res.Write([]byte(fmt.Sprintf("Hits: %d", cfg.FileServerHits)))
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

func (c *ApiConfig) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var rUser CreateUser
	if err := decoder.Decode(&rUser); err != nil {
		utils.RespondWithError(w, map[string]string{"error": "Invalid request"}, 400)
		return
	}

	user, err := c.Database.CreateUser(r.Context(), sql.NullString{String: rUser.Email, Valid: true})
	if err != nil {
		utils.RespondWithError(w, map[string]string{"error": "Error creating user"}, 500)
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
		utils.RespondWithError(w, map[string]string{"error": "Error marshalling body"}, 500)
	}

	w.WriteHeader(201)
	w.Write(body)
}

type PostChirp struct {
	Body string `json:"body"`
}

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

func (c *ApiConfig) HandleGetChirps(w http.ResponseWriter, r *http.Request) {
	// [todo]: fetch chirps from Database
	f, err := os.ReadFile("./database.json")
	if err != nil {
		utils.RespondWithError(w, map[string]string{"body": "Error opening database.json"}, 500)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(f)
}

func (c *ApiConfig) HandleCreateChirp(w http.ResponseWriter, r *http.Request) {
	// [todo]: create chirp in Database
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var chirp PostChirp
	error := map[string]string{"error": "Something went wrong"}
	if err := decoder.Decode(&chirp); err != nil {
		fmt.Println("error decoding")
		utils.RespondWithError(w, error, 400)
		return
	}

	if len(chirp.Body) > 140 {
		utils.RespondWithError(w, error, 400)
		return
	}

	if _, err := os.Stat("./database.json"); errors.Is(err, os.ErrNotExist) {
		if _, err := os.Create("./database.json"); err != nil {
			fmt.Println("error creating")
			utils.RespondWithError(w, error, 500)
			return
		}
	} else if err != nil {
		fmt.Println("error else if")
		utils.RespondWithError(w, error, 500)
		return
	}

	f, err := os.ReadFile("./database.json")
	if err != nil {
		fmt.Println("error reading")
		utils.RespondWithError(w, error, 500)
		return
	}

	var data []Chirp

	if len(f) > 0 {
		if err := json.Unmarshal(f, &data); err != nil {
			fmt.Println("error json unmarshal")
			log.Fatal(err)
			utils.RespondWithError(w, error, 500)
			return
		}
	}

	// newId := id
	// if len(data) > 0 {
	// 	newId = data[len(data)-1].Id + 1
	// }

	// newChirp := Chirp{Id: newId, Body: utils.RemoveBadWords(chirp.Body)}

	// data = append(data, newChirp)

	// dt, err := json.Marshal(data)
	// if err != nil {
	// 	utils.RespondWithError(w, error, 500)
	// 	return
	// }
	// if err := os.WriteFile("./database.json", dt, 0644); err != nil {
	// 	utils.RespondWithError(w, error, 500)
	// 	return
	// }

	// newBody, err := json.Marshal(newChirp)
	// if err != nil {
	// 	utils.RespondWithError(w, error, 500)
	// }

	// w.WriteHeader(201)
	// w.Write(newBody)
}
