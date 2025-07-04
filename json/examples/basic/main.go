package main

import (
	"fmt"

	"github.com/fsvxavier/nexs-lib/json"
)

type Person struct {
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Address string `json:"address"`
}

func main() {
	// Exemplo com o provider padrão (stdlib)
	personJSON := `{"name":"John Doe","age":30,"address":"123 Main St"}`

	var person Person
	err := json.Unmarshal([]byte(personJSON), &person)
	if err != nil {
		fmt.Printf("Error unmarshaling with default provider: %v\n", err)
		return
	}

	fmt.Printf("Default provider (stdlib): %+v\n", person)

	// Exemplo com o provider jsoniter
	jsoniterProvider := json.New(json.JSONIter)

	person = Person{} // Reset
	err = jsoniterProvider.Unmarshal([]byte(personJSON), &person)
	if err != nil {
		fmt.Printf("Error unmarshaling with jsoniter: %v\n", err)
		return
	}

	fmt.Printf("JSONIter provider: %+v\n", person)

	// Exemplo de marshal com goccy/go-json
	goccyProvider := json.New(json.GoccyJSON)

	person = Person{
		Name:    "Jane Doe",
		Age:     28,
		Address: "456 Oak St",
	}

	data, err := goccyProvider.Marshal(person)
	if err != nil {
		fmt.Printf("Error marshaling with goccy: %v\n", err)
		return
	}

	fmt.Printf("GoccyJSON marshal result: %s\n", string(data))
}
