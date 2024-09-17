package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type apiConfig struct {
	fileServerHits int
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

func healthzHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(200)
	res.Write([]byte("OK"))
}

func respondWithError(body map[string]string, w http.ResponseWriter, status int) {
	data, err := json.Marshal(body)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(data)
}

func handleValidateChirp(w http.ResponseWriter, r *http.Request) {
	type ValidateChirp struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var chirp ValidateChirp
	error := map[string]string{"error": "Something went wrong"}
	if err := decoder.Decode(&chirp); err != nil {
		respondWithError(error, w, 400)
		return
	}

	if len(chirp.Body) > 140 {
		respondWithError(error, w, 400)
		return
	}

	body := map[string]bool{"valid": true}
	data, err := json.Marshal(body)

	if err != nil {
		data, err := json.Marshal(error)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write(data)
		return
	}

	w.WriteHeader(200)
	w.Write(data)
}

func main() {
	api := apiConfig{fileServerHits: 0}
	mux := http.NewServeMux()
	serv := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	mux.Handle("/app/", api.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", healthzHandler)
	mux.HandleFunc("/api/reset", api.resetHandler)
	mux.HandleFunc("POST /api/validate_chirp", handleValidateChirp)
	mux.HandleFunc("GET /admin/metrics", api.countHandler)

	fmt.Println("Listening on port:8080")
	if err := serv.ListenAndServe(); err != nil {
		fmt.Println("Error 500: %w", err)
	}
}
