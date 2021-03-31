package api

import (
	"encoding/csv"
	"encoding/json"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"pokemon-api/entities"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"

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

		pokemonJSON, error := readCsv(pokemonID, w)

		if error != "" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(error))
		}
		w.Write(pokemonJSON)
	}
}

func readCsv(pokemonID int, w http.ResponseWriter) (pokemonJSON []byte, error string) {

	fileLocation, _ := filepath.Abs("assets/pokemon.csv")
	csvFile, err := os.Open(fileLocation)
	if checkError(err, "couldn't open csv", w) {
		return
	}

	defer csvFile.Close()

	csvLines, err := csv.NewReader(csvFile).ReadAll()
	if checkError(err, "couldn't read csv", w) {
		return
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

	result, err := readCsvConcurrently(w, numOfWorker, items, itemsPerWorkers, idtype)
	if err != "" {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err))
	}
	w.Write(result)

}

type SafeMap struct {
	mu  sync.Mutex
	res map[int]string
}

var (
	indexCounter int32
	resultMap    = SafeMap{res: make(map[int]string)}
	waitgroup    sync.WaitGroup
)

func worker(id int, lines [][]string, jobs int) {
	defer waitgroup.Done()
	for i := 0; i < jobs; i++ {
		resultMap.mu.Lock()
		if indexCounter < int32(len(lines)) {
			line := lines[indexCounter]
			pokemonIndex, _ := strconv.Atoi(line[0])
			resultMap.res[pokemonIndex] = line[1]
			atomic.AddInt32(&indexCounter, 2)
		} else {
			i = jobs
		}
		resultMap.mu.Unlock()
		runtime.Gosched()
	}
}

func readCsvConcurrently(w http.ResponseWriter, numOfWorker float64, items float64, itemsPerWorkers float64, idtype string) (pokemonJSON []byte, error string) {
	fileLocation, _ := filepath.Abs("assets/pokemons.csv")
	csvFile, err := os.Open(fileLocation)
	indexCounter = 0
	resultMap.res = make(map[int]string)
	if idtype == "odd" {
		indexCounter = 1
	}

	if err != nil {
		error = "{message:couldn't open csv}"
		return
	}

	defer csvFile.Close()

	csvLines, err := csv.NewReader(csvFile).ReadAll()

	if err != nil {
		error = "{message:couldn't read csv}"
		return
	}

	waitgroup.Add(int(numOfWorker))

	for w := 1; w <= int(numOfWorker); w++ {
		go worker(w, csvLines, int(itemsPerWorkers))
	}

	waitgroup.Wait()

	pokemonJSON, err = json.Marshal(resultMap.res)

	if err != nil {
		error = "{message:couldn't parse response}"
		return
	}

	return
}

func checkError(err error, message string, w http.ResponseWriter) bool {
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(message))
		return true
	}
	return false
}
