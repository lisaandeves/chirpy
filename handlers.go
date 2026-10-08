package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/lisaandeves/chirpy/internal/auth"
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

	pwd, err := auth.HashPassword(params.Password)
	if err != nil {
		writeErrorJson(writer, err, "Error creating password", http.StatusInternalServerError)
		return
	}

	user, err := cfg.db.CreateUser(req.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: pwd,
	})
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

func (cfg *apiConfig) handlerUpdateUser(writer http.ResponseWriter, req *http.Request) {
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		writeErrorJson(writer, err, "Invalid login credentials", http.StatusUnauthorized)
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		writeErrorJson(writer, err, "Invalid login credentials", http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(req.Body)
	params := userParams{}
	err = decoder.Decode(&params)
	if err != nil {
		writeErrorJson(writer, err, "Couldn't decode parameters", http.StatusInternalServerError)
		return
	}

	pwd, err := auth.HashPassword(params.Password)
	if err != nil {
		writeErrorJson(writer, err, "Error creating password", http.StatusInternalServerError)
		return
	}

	user, err := cfg.db.UpdateUser(req.Context(), database.UpdateUserParams{
		ID:             userID,
		Email:          params.Email,
		HashedPassword: pwd,
	})
	if err != nil {
		writeErrorJson(writer, err, "Couldn't update user details", http.StatusInternalServerError)
		return
	}

	resp := userWithTokenResponse{
		Id:        user.ID.String(),
		CreatedAt: user.CreatedAt.String(),
		UpdatedAt: user.UpdatedAt.String(),
		Email:     user.Email,
		Token:     token,
	}
	writeResponseJson(writer, resp, http.StatusOK)
}

func (cfg *apiConfig) handlerLoginUser(writer http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	params := userParams{}
	err := decoder.Decode(&params)
	if err != nil {
		writeErrorJson(writer, err, "Couldn't decode parameters", http.StatusInternalServerError)
		return
	}

	var expiresIn time.Duration
	if params.ExpiresInSeconds == nil || *params.ExpiresInSeconds > 3600 {
		expiresIn = time.Hour
	} else {
		expiresIn = time.Duration(*params.ExpiresInSeconds) * time.Second
	}

	user, err := cfg.db.GetUserByEmail(req.Context(), params.Email)
	if err != nil {
		writeErrorJson(writer, err, "Incorrect email or password", http.StatusUnauthorized)
		return
	}

	ok, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil || !ok {
		writeErrorJson(writer, err, "Incorrect email or password", http.StatusUnauthorized)
		return
	}

	refreshToken := auth.MakeRefreshToken()
	_, err = cfg.db.CreateRefreshToken(req.Context(), database.CreateRefreshTokenParams{
		Token:  refreshToken,
		UserID: user.ID,
	})
	if err != nil {
		writeErrorJson(writer, err, "Error authenticating user", http.StatusInternalServerError)
		return
	}

	tokenString, err := auth.MakeJWT(user.ID, cfg.secret, expiresIn)
	if err != nil {
		writeErrorJson(writer, err, "Error authenticating user", http.StatusInternalServerError)
		return
	}

	resp := userWithTokenResponse{
		Id:           user.ID.String(),
		CreatedAt:    user.CreatedAt.String(),
		UpdatedAt:    user.UpdatedAt.String(),
		Email:        user.Email,
		Token:        tokenString,
		RefreshToken: refreshToken,
	}
	writeResponseJson(writer, resp, http.StatusOK)
}

func (cfg *apiConfig) handlerAddChirp(writer http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	params := chirpParams{}
	err := decoder.Decode(&params)
	if err != nil {
		writeErrorJson(writer, err, "Couldn't decode parameters", http.StatusInternalServerError)
		return
	}

	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		writeErrorJson(writer, err, "Invalid login credentials", http.StatusUnauthorized)
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		writeErrorJson(writer, err, "Invalid login credentials", http.StatusUnauthorized)
		return
	}

	chirp, err := cfg.db.CreateChirp(req.Context(), database.CreateChirpParams{
		Body:   params.Body,
		UserID: userID,
	})
	if err != nil {
		writeErrorJson(writer, err, "Couldn't create chirp", http.StatusInternalServerError)
		return
	}
	if len(params.Body) > 140 {
		writeErrorJson(writer, errors.New(""), "Chirp is too long", http.StatusBadRequest)
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

func (cfg *apiConfig) handlerRefresh(writer http.ResponseWriter, req *http.Request) {
	refreshToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		writeErrorJson(writer, err, "Invalid login credentials", http.StatusUnauthorized)
		return
	}

	refreshTokenInfo, err := cfg.db.GetRefreshToken(req.Context(), refreshToken)
	if err != nil ||
		refreshTokenInfo.ExpiresAt.Before(time.Now()) ||
		refreshTokenInfo.RevokedAt.Valid == true {
		writeErrorJson(writer, err, "Invalid login credentials", http.StatusUnauthorized)
		return
	}

	user, err := cfg.db.GetUserByRefreshToken(req.Context(), refreshToken)
	if err != nil {
		writeErrorJson(writer, err, "Invalid login credentials", http.StatusUnauthorized)
		return
	}

	newJWT, err := auth.MakeJWT(user.ID, cfg.secret, time.Hour)
	if err != nil {
		writeErrorJson(writer, err, "Error authenticating user", http.StatusInternalServerError)
		return
	}

	resp := userWithTokenResponse{
		Id:           user.ID.String(),
		CreatedAt:    user.CreatedAt.String(),
		UpdatedAt:    user.UpdatedAt.String(),
		Email:        user.Email,
		Token:        newJWT,
		RefreshToken: refreshToken,
	}
	writeResponseJson(writer, resp, http.StatusOK)
}

func (cfg *apiConfig) handlerRevoke(writer http.ResponseWriter, req *http.Request) {
	refreshToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		writeErrorJson(writer, err, "Invalid login credentials", http.StatusUnauthorized)
		return
	}

	err = cfg.db.RevokeRefreshToken(req.Context(), refreshToken)
	if err != nil {
		writeErrorJson(writer, err, "Error with login credentials", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
	writer.Write(nil)
}
