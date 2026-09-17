package main

import (
	"encoding/json"
	"fmt"
	"net/http"
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

func (cfg *apiConfig) handlerResetHits(writer http.ResponseWriter, _ *http.Request) {
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
	type responseError struct {
		Error string `json:"error"`
	}
	type responseValid struct {
		Valid bool `json:"valid"`
	}

	writer.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(req.Body)
	params := chirpJson{}
	err := decoder.Decode(&params)
	if err != nil {
		resp := responseError{Error: fmt.Sprintf("Error decoding parameters: %s", err)}
		dat, err := json.Marshal(resp)
		if err != nil {
			//Unreachable
		}
		writer.WriteHeader(500)
		writer.Write(dat)
		return
	}
	if len(params.Body) > 140 {
		resp := responseError{Error: "Chirp is too long"}
		dat, err := json.Marshal(resp)
		if err != nil {
			//Unreachable
		}
		writer.WriteHeader(400)
		writer.Write(dat)
		return
	}
	resp := responseValid{Valid: true}
	dat, err := json.Marshal(resp)
	if err != nil {
		//Unreachable
	}
	writer.WriteHeader(200)
	writer.Write(dat)
}
