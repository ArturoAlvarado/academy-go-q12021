package handlers

import (
	"encoding/csv"
	"encoding/json"
	"math"
	"net/http"
	"os"
	"pokemon-api/entities"
	"pokemon-api/pkg/usecases"
	"strconv"

	"github.com/gorilla/mux"
)

//GetCsv pokemon handlers
func GetCsv(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")

	if val, ok := pathParams["pokemonID"]; ok {
		pokemonID, err := strconv.Atoi(val)
		if usecases.CheckError(err, "need a pokemonID", w) {
			return
		}

		pokemonJSON, error := usecases.ReadCsv(pokemonID, w)

		if error != "" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(error))
		}
		w.Write(pokemonJSON)
	}
}

//GetFromExternal get pokemon from external api
func GetFromExternal(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response, err := http.Get("https://pokeapi.co/api/v2/pokemon")

	if usecases.CheckError(err, "error getting pokemons", w) {
		return
	}
	defer response.Body.Close()

	data := &entities.Pokemons{}
	if usecases.CheckError(err, "error parsing pokemon", w) {
		return
	}

	decodeError := json.NewDecoder(response.Body).Decode(&data)

	if usecases.CheckError(decodeError, "error parsing pokemon", w) {
		return
	}

	pokemonJSON, _ := json.Marshal(data.Results)

	file, err := os.Create("./pokemons.csv")
	defer file.Close()

	if usecases.CheckError(err, "could not create csv", w) {
		return
	}

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for index, value := range data.Results {
		s := make([]string, 0)
		s = append(s, strconv.Itoa(index))
		s = append(s, value.Name)
		s = append(s, value.URL)
		err := writer.Write(s)

		usecases.CheckError(err, "could write to csv", w)
	}
	w.Write(pokemonJSON)

}

func GetConcurrently(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	items, _ := strconv.ParseFloat(r.URL.Query().Get("items"), 64)
	itemsPerWorkers, _ := strconv.ParseFloat(r.URL.Query().Get("items_per_workers"), 64)
	idtype := r.URL.Query().Get("type")

	if items == 0 || itemsPerWorkers == 0 || (idtype != "even" && idtype != "odd") {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{message: invalid params}"))
		return
	}
	numOfWorker := math.Ceil(items / itemsPerWorkers)

	result, err := usecases.ReadCsvConcurrently(w, numOfWorker, items, itemsPerWorkers, idtype)
	if err != "" {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err))
	}
	w.Write(result)

}
