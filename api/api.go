package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/LahcenHaouch/goserver/internal/auth"
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
	Email    string `json:"email"`
	Password string `json:"password"`
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

	hashedPassword, err := auth.HashPassword(rUser.Password)
	if err != nil {
		utils.RespondWithError(w, map[string]string{"error": "Error hashing password"}, 500)
		return
	}

	user, err := c.Database.CreateUser(r.Context(), database.CreateUserParams{Email: sql.NullString{String: rUser.Email, Valid: true}, HashedPassword: hashedPassword})
	if err != nil {
		utils.RespondWithError(w, map[string]string{"error": "Error creating user"}, 500)
		return
	}

	newUser := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
		Email:     user.Email.String,
	}

	body, err := json.Marshal(newUser)
	if err != nil {
		utils.RespondWithError(w, map[string]string{"error": "Error marshalling body"}, 500)
		return
	}

	w.WriteHeader(201)
	w.Write(body)
}

type PostChirp struct {
	Body   string `json:"body"`
	UserId string `json:"user_id"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserId    uuid.UUID `json:"user_id"`
}

func parseDbChirp(chirp database.Chirp) Chirp {
	return Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt.Time,
		UpdatedAt: chirp.UpdatedAt.Time,
		Body:      chirp.Body.String,
		UserId:    chirp.UserID.UUID,
	}
}

func parseDbChirps(chirps []database.Chirp) []Chirp {
	var parsed []Chirp

	for _, chirp := range chirps {
		parsed = append(parsed, parseDbChirp(chirp))
	}

	return parsed
}

func (c *ApiConfig) HandleGetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := c.Database.GetChirps(r.Context())
	if err != nil {
		utils.RespondWithError(w, map[string]string{"error": "error fetching chirps from database"}, 500)
		return
	}

	body, err := json.Marshal(parseDbChirps(chirps))
	if err != nil {
		utils.RespondWithError(w, map[string]string{"error": "error converting chirps to []byte"}, 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func (c *ApiConfig) HandleGetChirp(w http.ResponseWriter, r *http.Request) {
	pathChirpId := r.PathValue("chirpId")
	chirpId, err := uuid.Parse(pathChirpId)
	if err != nil {
		utils.RespondWithError(w, map[string]string{"error": "error parsing chirp id"}, 400)
	}

	dbChirp, err := c.Database.GetChirp(r.Context(), chirpId)
	if err != nil {
		utils.RespondWithError(w, map[string]string{"error": "error fetching chirp"}, 500)
	}
	chirp := parseDbChirp(dbChirp)

	body, err := json.Marshal(chirp)
	if err != nil {
		utils.RespondWithError(w, map[string]string{"error": "error parsing chirp"}, 500)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func (c *ApiConfig) HandleCreateChirp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var chirp PostChirp
	if err := decoder.Decode(&chirp); err != nil {
		utils.RespondWithError(w, map[string]string{"error": "error decoding body"}, 400)
		return
	}

	if len(chirp.Body) > 140 {
		utils.RespondWithError(w, map[string]string{"error": "body length > 140"}, 400)
		return
	}

	userId, err := uuid.Parse(chirp.UserId)
	if err != nil {
		utils.RespondWithError(w, map[string]string{"error": "error parsing user_id"}, 500)
		return
	}

	newChirp, err := c.Database.CreateChirp(r.Context(), database.CreateChirpParams{Body: sql.NullString{String: utils.RemoveBadWords(chirp.Body), Valid: true}, UserID: uuid.NullUUID{UUID: userId, Valid: true}})
	if err != nil {
		utils.RespondWithError(w, map[string]string{"error": "error creating chirp"}, 500)
		return
	}

	jsonChirp := Chirp{
		ID:        newChirp.ID,
		CreatedAt: newChirp.CreatedAt.Time,
		UpdatedAt: newChirp.UpdatedAt.Time,
		Body:      newChirp.Body.String,
		UserId:    newChirp.UserID.UUID,
	}

	newBody, err := json.Marshal(jsonChirp)
	if err != nil {
		utils.RespondWithError(w, map[string]string{"error": "error marshalling response body"}, 500)
		return
	}

	w.WriteHeader(201)
	w.Write(newBody)
}

type Login struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (c *ApiConfig) HandleLogin(w http.ResponseWriter, r *http.Request) {
	body := r.Body
	defer r.Body.Close()

	var login Login
	if err := json.NewDecoder(body).Decode(&login); err != nil {
		utils.RespondWithError(w, map[string]string{"error": "Error parsing login data"}, 500)
		return
	}

	user, err := c.Database.GetUser(r.Context(), sql.NullString{String: login.Email, Valid: true})
	if err != nil {
		http.Error(w, "Incorrect email or password", 401)
		return
	}

	if err = auth.CheckPasswordHash(login.Password, user.HashedPassword); err != nil {
		http.Error(w, "Incorrect email or password", 401)
		return
	}

	u := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
		Email:     user.Email.String,
	}

	newBody, err := json.Marshal(u)
	if err != nil {
		http.Error(w, "Error marshalling user", 500)
		return
	}

	w.WriteHeader(200)
	w.Write(newBody)
}
