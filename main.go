package main

import (
	"log"
	"net/http"
)

func main() {
	addr := ":8080"
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(".")))
	server := http.Server{Addr: addr, Handler: mux}
	log.Fatal(server.ListenAndServe())
}
