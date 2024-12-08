package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
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
	TokenSecret    string
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
	Body string `json:"body"`
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
		return
	}

	dbChirp, err := c.Database.GetChirp(r.Context(), chirpId)
	if err != nil {
		utils.RespondWithError(w, map[string]string{"error": "error fetching chirp"}, 404)
		return
	}
	chirp := parseDbChirp(dbChirp)

	body, err := json.Marshal(chirp)
	if err != nil {
		utils.RespondWithError(w, map[string]string{"error": "error parsing chirp"}, 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func (c *ApiConfig) HandleCreateChirp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, "401 Unauthorized", 401)
		log.Println("184")
		return
	}

	userId, err := auth.ValidateJWT(token, c.TokenSecret)
	if err != nil {
		http.Error(w, "401 Unauthorized", 401)
		log.Println("190")
		return
	}

	var chirp PostChirp
	if err := decoder.Decode(&chirp); err != nil {
		utils.RespondWithError(w, map[string]string{"error": "error decoding body"}, 400)
		return
	}

	if len(chirp.Body) > 140 {
		utils.RespondWithError(w, map[string]string{"error": "body length > 140"}, 400)
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
	type newBody struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

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

	token, err := auth.MakeJWT(user.ID, c.TokenSecret, time.Hour)
	if err != nil {
		http.Error(w, "Error generating jwt token", 500)
		return
	}

	u := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
		Email:     user.Email.String,
	}

	refreshTokenStr, err := auth.MakeRefreshToken()
	if err != nil {
		http.Error(w, "error creating refresh token", 500)
		return
	}

	now := time.Now()
	refreshToken, err := c.Database.CreateRefreshToken(
		r.Context(), database.CreateRefreshTokenParams{
			Token:     refreshTokenStr,
			UserID:    uuid.NullUUID{UUID: u.ID, Valid: true},
			ExpiresAt: sql.NullTime{Time: now.Add(time.Hour * 24 * 60), Valid: true}})

	if err != nil {
		http.Error(w, "error saving refresh token", 500)
		return
	}
	resp := newBody{
		User:         u,
		Token:        token,
		RefreshToken: refreshToken.Token,
	}

	bodyToSend, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Error marshalling user", 500)
		return
	}

	w.WriteHeader(200)
	w.Write(bodyToSend)
}

func (c *ApiConfig) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	tokenStr, err := auth.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, "error reading token", 401)
	}

	token, err := c.Database.GetRefreshToken(r.Context(), tokenStr)
	if err != nil {
		http.Error(w, "error retrieving token", 401)
		return
	}

	if token.RevokedAt.Valid {
		http.Error(w, "error retrieving token", 401)
		return
	}

	accessToken, err := auth.MakeJWT(token.UserID.UUID, c.TokenSecret, time.Hour)
	if err != nil {
		http.Error(w, "error creating access token", 500)
		return
	}

	m := map[string]string{
		"token": accessToken,
	}
	body, err := json.Marshal(m)
	if err != nil {
		http.Error(w, "error marshalling body", 500)
		return
	}

	w.Write(body)
	w.WriteHeader(200)
}

func (c *ApiConfig) HandleRevoke(w http.ResponseWriter, r *http.Request) {
	tokenStr, err := auth.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, "error token string", 401)
		return
	}

	_, err = c.Database.GetRefreshToken(r.Context(), tokenStr)
	if err != nil {
		http.Error(w, "error retrieving token from db", 401)
		return
	}

	now := time.Now()
	nullTime := sql.NullTime{
		Time:  now,
		Valid: true,
	}
	err = c.Database.UpdateRefreshToken(r.Context(), database.UpdateRefreshTokenParams{
		Token:     tokenStr,
		UpdatedAt: nullTime,
		RevokedAt: nullTime,
	})
	if err != nil {
		http.Error(w, "error udpdating token", 500)
		return
	}
	w.WriteHeader(204)
}

func (c *ApiConfig) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	tokenStr, err := auth.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, "unauthorized", 401)
		return
	}

	userId, err := auth.ValidateJWT(tokenStr, c.TokenSecret)
	if err != nil {
		http.Error(w, "unauthorized", 401)
		return
	}

	defer r.Body.Close()

	var user CreateUser
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "error parsing user information", 400)
		return
	}

	hashedPassword, err := auth.HashPassword(user.Password)
	if err != nil {
		http.Error(w, "Internal server error", 500)
		return
	}

	updatedUser, err := c.Database.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             userId,
		Email:          sql.NullString{String: user.Email, Valid: true},
		HashedPassword: hashedPassword,
	})
	if err != nil {
		http.Error(w, "Internal server error", 500)
		return
	}

	body, err := json.Marshal(User{
		ID:        userId,
		Email:     updatedUser.Email.String,
		CreatedAt: updatedUser.CreatedAt.Time,
		UpdatedAt: updatedUser.UpdatedAt.Time,
	})
	if err != nil {
		http.Error(w, "Internal server error", 500)
	}
	w.Write(body)
	w.WriteHeader(200)
}

func (c *ApiConfig) HandleDeleteChirp(w http.ResponseWriter, r *http.Request) {
	tokenStr, err := auth.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, "unauthorized", 401)
		return
	}

	userId, err := auth.ValidateJWT(tokenStr, c.TokenSecret)
	if err != nil {
		http.Error(w, "unauthorized", 401)
		return
	}

	chirpIdP := r.PathValue("chirpId")
	chirpId, err := uuid.Parse(chirpIdP)
	if err != nil {
		http.Error(w, "bad request", 400)
		return
	}

	chirp, err := c.Database.GetChirp(r.Context(), chirpId)
	if err != nil {
		http.Error(w, "not found", 404)
		return
	}

	if chirp.UserID.UUID != userId {
		http.Error(w, "unauthorized", 403)
		return
	}

	err = c.Database.DeleteChirp(r.Context(), chirpId)
	if err != nil {
		http.Error(w, "internal server error", 500)
		return
	}

	w.WriteHeader(204)
}
