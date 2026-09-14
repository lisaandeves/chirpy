package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(writer, req)
	})
}

func (cfg *apiConfig) hitsHandler(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	writer.Write(fmt.Appendf(nil, "Hits: %d", cfg.fileserverHits.Load()))
}

func (cfg *apiConfig) resetHitsHandler(writer http.ResponseWriter, _ *http.Request) {
	cfg.fileserverHits.Store(0)
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(http.StatusText(http.StatusOK)))
}

func readinessHandler(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(http.StatusText(http.StatusOK)))
}

func main() {
	addr := ":8080"
	cfg := apiConfig{}
	mux := http.NewServeMux()

	fileHandler := http.StripPrefix("/app", cfg.middlewareMetricsInc(http.FileServer(http.Dir("."))))
	mux.Handle("/app/", fileHandler)
	mux.HandleFunc("/healthz", readinessHandler)
	mux.HandleFunc("/metrics", cfg.hitsHandler)
	mux.HandleFunc("/reset", cfg.resetHitsHandler)

	server := http.Server{Addr: addr, Handler: mux}
	log.Fatal(server.ListenAndServe())
}
