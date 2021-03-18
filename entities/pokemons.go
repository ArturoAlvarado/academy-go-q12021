package entities

//Pokemons APi response struct
type Pokemons struct {
	count    int
	next     string
	previous string
	Results  []Pokemon `json:"results,omitempty"`
}

//Pokemon struct
type Pokemon struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}
