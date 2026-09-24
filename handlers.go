package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/lisaandeves/chirpy/internal/database"
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

func (cfg *apiConfig) handlerAddUser(writer http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	params := userParams{}
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

	resp := userResponse{
		Id:        user.ID.String(),
		CreatedAt: user.CreatedAt.String(),
		UpdatedAt: user.UpdatedAt.String(),
		Email:     user.Email,
	}
	writeResponseJson(writer, resp, http.StatusCreated)
}

func (cfg *apiConfig) handlerAddChirp(writer http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	params := chirpParams{}
	err := decoder.Decode(&params)
	if err != nil {
		writeErrorJson(writer, err, "Couldn't decode parameters", http.StatusInternalServerError)
		return
	}

	chirp, err := cfg.db.CreateChirp(req.Context(), database.CreateChirpParams{
		Body:   params.Body,
		UserID: uuid.MustParse(params.UserId),
	})
	if err != nil {
		writeErrorJson(writer, err, "Couldn't create chirp", http.StatusInternalServerError)
		return
	}
	if len(params.Body) > 140 {
		writeErrorJson(writer, nil, "Chirp is too long", http.StatusBadRequest)
		return
	}

	resp := chirpResponse{
		Id:        chirp.ID.String(),
		CreatedAt: chirp.CreatedAt.String(),
		UpdatedAt: chirp.UpdatedAt.String(),
		Body:      chirpCensor(chirp.Body),
		UserId:    chirp.UserID.String(),
	}
	writeResponseJson(writer, resp, http.StatusCreated)
}

func (cfg *apiConfig) handlerGetAllChirps(writer http.ResponseWriter, req *http.Request) {
	chirps, err := cfg.db.GetAllChirps(req.Context())
	if err != nil {
		writeErrorJson(writer, err, "Couldn't retrieve chirps", http.StatusInternalServerError)
		return
	}

	resp := []chirpResponse{}
	for _, chirp := range chirps {
		resp = append(resp, chirpResponse{
			Id:        chirp.ID.String(),
			CreatedAt: chirp.CreatedAt.String(),
			UpdatedAt: chirp.UpdatedAt.String(),
			Body:      chirp.Body,
			UserId:    chirp.UserID.String(),
		})
	}
	writeResponseJson(writer, resp, http.StatusOK)
}

func (cfg *apiConfig) handlerGetChirp(writer http.ResponseWriter, req *http.Request) {
	chirp_id_str := req.PathValue("id")
	chirp_id, err := uuid.Parse(chirp_id_str)
	if err != nil {
		writeErrorJson(writer, err, "Invalid chirp ID", http.StatusBadRequest)
		return
	}
	chirp, err := cfg.db.GetChirp(req.Context(), chirp_id)
	if err != nil {
		writeErrorJson(writer, err, "Chirp not found", http.StatusNotFound)
		return
	}

	resp := chirpResponse{
		Id:        chirp.ID.String(),
		CreatedAt: chirp.CreatedAt.String(),
		UpdatedAt: chirp.UpdatedAt.String(),
		Body:      chirp.Body,
		UserId:    chirp.UserID.String(),
	}
	writeResponseJson(writer, resp, http.StatusOK)
}
