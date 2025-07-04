package main

import (
	"fmt"
	"strings"

	"github.com/fsvxavier/nexs-lib/json"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func main() {
	// Dados de exemplo
	jsonString := `{"id":123,"username":"johndoe","email":"john@example.com"}`

	// Criando um reader a partir da string
	reader := strings.NewReader(jsonString)

	// Decodificando com o provider padrão
	var user User
	err := json.DecodeReader(reader, &user)
	if err != nil {
		fmt.Printf("Error decoding with default provider: %v\n", err)
		return
	}

	fmt.Printf("Default provider DecodeReader: %+v\n", user)

	// Usando DecodeReader com diferentes providers
	providersToTest := []struct {
		Name     string
		Provider json.ProviderType
	}{
		{"Stdlib", json.Stdlib},
		{"JSONIter", json.JSONIter},
		{"GoccyJSON", json.GoccyJSON},
		{"JSONParser", json.JSONParser},
	}

	for _, p := range providersToTest {
		// Reset reader e user
		reader = strings.NewReader(jsonString)
		user = User{}

		provider := json.New(p.Provider)
		err := provider.DecodeReader(reader, &user)
		if err != nil {
			fmt.Printf("Error decoding with %s: %v\n", p.Name, err)
			continue
		}

		fmt.Printf("%s DecodeReader: %+v\n", p.Name, user)
	}

	// Exemplo com NewDecoder
	jsonData := `[
		{"id":1,"username":"user1","email":"user1@example.com"},
		{"id":2,"username":"user2","email":"user2@example.com"}
	]`

	reader = strings.NewReader(jsonData)
	decoder := json.NewDecoder(reader)

	// Lendo um array de usuários
	var users []User
	err = decoder.Decode(&users)
	if err != nil {
		fmt.Printf("Error decoding array: %v\n", err)
		return
	}

	fmt.Println("Decoded array of users:")
	for i, u := range users {
		fmt.Printf("  %d: %+v\n", i, u)
	}
}
