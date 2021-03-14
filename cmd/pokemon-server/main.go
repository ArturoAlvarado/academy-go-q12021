package main

import (
	"log"
	"net/http"
	"pokemon-api/pkg/api"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/pokemon/{pokemonID}", api.GetCsv).Methods(http.MethodGet)
	log.Fatal(http.ListenAndServe(":8080", r))
}
