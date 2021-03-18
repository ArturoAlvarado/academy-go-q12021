package api

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"pokemon-api/entities"
	"strconv"

	"github.com/gorilla/mux"
)

//GetCsv pokemon handlers
func GetCsv(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")

	if val, ok := pathParams["pokemonID"]; ok {
		pokemonID, err := strconv.Atoi(val)
		if checkError(err, "need a pokemonID", w) {
			return
		}

		pokemonJSON, error := readCsv(pokemonID)

		if error != "" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(error))
		}
		w.Write(pokemonJSON)
	}
}

func readCsv(pokemonID int) (pokemonJSON []byte, error string) {

	csvFile, err := os.Open("assets/pokemon.csv")
	if err != nil {
		fmt.Println(err)
	}
	defer csvFile.Close()

	csvLines, err := csv.NewReader(csvFile).ReadAll()
	if err != nil {
		fmt.Println(err)
	}

	pokemonMap := make(map[int]string)

	for _, line := range csvLines {
		pokemonIndex, _ := strconv.Atoi(line[0])
		pokemonMap[pokemonIndex] = line[1]
	}

	pokemonName, ok := pokemonMap[pokemonID]

	if ok {
		resultPokemon := make(map[string]string)
		resultPokemon["id"] = strconv.Itoa(pokemonID)
		resultPokemon["name"] = pokemonName
		pokemonJSON, _ = json.Marshal(resultPokemon)
		return
	}
	error = `{"message": "pokemon not found"}`
	return
}

//GetFromExternal get pokemon from external api
func GetFromExternal(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response, err := http.Get("https://pokeapi.co/api/v2/pokemon")

	if checkError(err, "error getting pokemons", w) {
		return
	}
	defer response.Body.Close()

	data := &entities.Pokemons{}
	if checkError(err, "error parsing pokemon", w) {
		return
	}

	decodeError := json.NewDecoder(response.Body).Decode(&data)

	if checkError(decodeError, "error parsing pokemon", w) {
		return
	}

	pokemonJSON, _ := json.Marshal(data.Results)

	file, err := os.Create("./pokemons.csv")
	defer file.Close()

	if checkError(err, "could not create csv", w) {
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

		checkError(err, "could write to csv", w)
	}
	w.Write(pokemonJSON)

}

func checkError(err error, message string, w http.ResponseWriter) bool {
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(message))
		return true
	}
	return false
}
