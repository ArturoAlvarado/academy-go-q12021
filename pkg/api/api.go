package api

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

//GetCsv pokemon handlers
func GetCsv(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")

	if val, ok := pathParams["pokemonID"]; ok {
		pokemonID, err := strconv.Atoi(val)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "need a pokemonID"}`))
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

	resultPokemon, ok := pokemonMap[pokemonID]

	if ok {
		pokemonJSON, _ = json.Marshal(resultPokemon)
		return
	}
	error = `{"message": "pokemon not found"}`
	return
}
