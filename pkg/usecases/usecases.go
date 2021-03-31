package usecases

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"pokemon-api/entities"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
)

var (
	indexCounter int32
	resultMap    = entities.SafeMap{Res: make(map[int]string)}
	waitgroup    sync.WaitGroup
)

func ReadCsv(pokemonID int, w http.ResponseWriter) (pokemonJSON []byte, error string) {

	fileLocation, _ := filepath.Abs("assets/pokemon.csv")
	csvFile, err := os.Open(fileLocation)
	if CheckError(err, "couldn't open csv", w) {
		return
	}

	defer csvFile.Close()

	csvLines, err := csv.NewReader(csvFile).ReadAll()
	if CheckError(err, "couldn't read csv", w) {
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

func worker(id int, data *[][]string, jobs int) {
	defer waitgroup.Done()
	for i := 0; i < jobs; i++ {
		resultMap.Mu.Lock()
		lines := *data
		if indexCounter < int32(len(lines)) {
			line := lines[indexCounter]
			pokemonIndex, _ := strconv.Atoi(line[0])
			resultMap.Res[pokemonIndex] = line[1]
			atomic.AddInt32(&indexCounter, 2)
		} else {
			i = jobs
		}
		resultMap.Mu.Unlock()
		runtime.Gosched()
	}
}

func ReadCsvConcurrently(w http.ResponseWriter, numOfWorker float64, items float64, itemsPerWorkers float64, idtype string) (pokemonJSON []byte, error string) {
	fileLocation, _ := filepath.Abs("assets/pokemons.csv")
	csvFile, err := os.Open(fileLocation)
	indexCounter = 0
	resultMap.Res = make(map[int]string)
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
		go worker(w, &csvLines, int(itemsPerWorkers))
	}

	waitgroup.Wait()

	pokemonJSON, err = json.Marshal(resultMap.Res)

	if err != nil {
		error = "{message:couldn't parse response}"
		return
	}

	return
}

func CheckError(err error, message string, w http.ResponseWriter) bool {
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(message))
		return true
	}
	return false
}
