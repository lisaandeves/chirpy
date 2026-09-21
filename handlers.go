package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func (cfg *apiConfig) handlerMetrics(writer http.ResponseWriter, _ *http.Request) {
	httpString := `<html>
  	<body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  	</body>
	</html>`

	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	writer.Write(fmt.Appendf(nil, httpString, cfg.fileserverHits.Load()))
}

func (cfg *apiConfig) handlerReset(writer http.ResponseWriter, req *http.Request) {
	platform := os.Getenv("PLATFORM")
	if platform != "dev" {
		respondWithErrorJson(writer, nil, "Forbidden", http.StatusForbidden)
		return
	}

	err := cfg.db.DeleteAllUsers(req.Context())
	if err != nil {
		respondWithErrorJson(writer, err, "Couldn't delete all users", http.StatusInternalServerError)
		return
	}
	cfg.fileserverHits.Store(0)
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(http.StatusText(http.StatusOK)))
}

func handlerReadiness(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(http.StatusText(http.StatusOK)))
}

func handlerChirpValidate(writer http.ResponseWriter, req *http.Request) {
	type chirpJson struct {
		Body string `json:"body"`
	}
	type responseValid struct {
		CleanedBody string `json:"cleaned_body"`
	}

	writer.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(req.Body)
	params := chirpJson{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithErrorJson(writer, err, "Couldn't decode parameters", http.StatusInternalServerError)
		return
	}
	if len(params.Body) > 140 {
		respondWithErrorJson(writer, nil, "Chirp is too long", http.StatusBadRequest)
		return
	}
	resp := responseValid{CleanedBody: chirpCensor(params.Body)}
	dat, err := json.Marshal(resp)
	if err != nil {
		//Unreachable
	}
	writer.WriteHeader(200)
	writer.Write(dat)
}

func (cfg *apiConfig) handlerAddUser(writer http.ResponseWriter, req *http.Request) {
	type emailJson struct {
		Email string `json:"email"`
	}
	type userJson struct {
		Id        string `json:"id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Email     string `json:"email"`
	}

	writer.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(req.Body)
	params := emailJson{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithErrorJson(writer, err, "Couldn't decode parameters", http.StatusInternalServerError)
		return
	}

	user, err := cfg.db.CreateUser(req.Context(), params.Email)
	if err != nil {
		respondWithErrorJson(writer, err, "Couldn't create user", http.StatusInternalServerError)
		return
	}
	resp := userJson{
		Id:        user.ID.String(),
		CreatedAt: user.CreatedAt.String(),
		UpdatedAt: user.UpdatedAt.String(),
		Email:     user.Email,
	}
	dat, err := json.Marshal(resp)
	if err != nil {
		//Unreachable
	}
	writer.WriteHeader(http.StatusCreated)
	writer.Write(dat)
}
