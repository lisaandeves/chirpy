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
		writer.WriteHeader(http.StatusForbidden)
		writer.Write([]byte("dev environment only"))
		return
	}

	err := cfg.db.DeleteAllUsers(req.Context())
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("Couldn't delete all users: " + err.Error()))
		return
	}

	cfg.fileserverHits.Store(0)
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("Database and hits reset"))
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

	decoder := json.NewDecoder(req.Body)
	params := chirpJson{}
	err := decoder.Decode(&params)
	if err != nil {
		writeErrorJson(writer, err, "Couldn't decode parameters", http.StatusInternalServerError)
		return
	}
	if len(params.Body) > 140 {
		writeErrorJson(writer, nil, "Chirp is too long", http.StatusBadRequest)
		return
	}

	resp := responseValid{CleanedBody: chirpCensor(params.Body)}
	writeResponseJson(writer, resp, http.StatusOK)
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

	decoder := json.NewDecoder(req.Body)
	params := emailJson{}
	err := decoder.Decode(&params)
	if err != nil {
		writeErrorJson(writer, err, "Couldn't decode parameters", http.StatusInternalServerError)
		return
	}

	user, err := cfg.db.CreateUser(req.Context(), params.Email)
	if err != nil {
		writeErrorJson(writer, err, "Couldn't create user", http.StatusInternalServerError)
		return
	}

	resp := userJson{
		Id:        user.ID.String(),
		CreatedAt: user.CreatedAt.String(),
		UpdatedAt: user.UpdatedAt.String(),
		Email:     user.Email,
	}
	writeResponseJson(writer, resp, http.StatusCreated)
}
