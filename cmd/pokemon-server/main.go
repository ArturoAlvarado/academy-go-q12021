package main

import (
	"log"
	"net/http"
	"pokemon-api/pkg/handlers"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/pokemon/{pokemonID}", handlers.GetCsv).Methods(http.MethodGet)
	r.HandleFunc("/pokemons", handlers.GetFromExternal).Methods(http.MethodGet)

	r.HandleFunc("/concurrently", handlers.GetConcurrently).Methods(http.MethodGet)
	log.Fatal(http.ListenAndServe(":8080", r))
}
