package main

import (
	"fmt"
	"math/rand"
	"time"
)

// Customer representa um cliente
type Customer struct {
	ID               int       `json:"id"`
	Name             string    `json:"name"`
	Email            string    `json:"email"`
	Phone            string    `json:"phone"`
	City             string    `json:"city"`
	Country          string    `json:"country"`
	RegistrationDate time.Time `json:"registration_date"`
	CreatedAt        time.Time `json:"created_at"`
}

// Product representa um produto
type Product struct {
	ID         int       `json:"id"`
	SKU        string    `json:"sku"`
	Name       string    `json:"name"`
	Category   string    `json:"category"`
	Price      float64   `json:"price"`
	Cost       float64   `json:"cost"`
	Stock      int       `json:"stock"`
	Weight     float64   `json:"weight"`
	Dimensions string    `json:"dimensions"`
	CreatedAt  time.Time `json:"created_at"`
}

// Order representa um pedido
type Order struct {
	ID              int       `json:"id"`
	CustomerID      int       `json:"customer_id"`
	ProductID       int       `json:"product_id"`
	Quantity        int       `json:"quantity"`
	UnitPrice       float64   `json:"unit_price"`
	TotalPrice      float64   `json:"total_price"`
	OrderDate       time.Time `json:"order_date"`
	Status          string    `json:"status"`
	ShippingAddress string    `json:"shipping_address"`
	CreatedAt       time.Time `json:"created_at"`
}

// generateCustomers gera uma lista de clientes fictícios
func generateCustomers(count int) []Customer {
	fmt.Printf("   🔧 Gerando %d clientes...\n", count)

	firstNames := []string{
		"João", "Maria", "Pedro", "Ana", "Carlos", "Lucia", "José", "Fernanda",
		"Paulo", "Mariana", "Bruno", "Juliana", "Ricardo", "Camila", "André",
		"Beatriz", "Roberto", "Carla", "Daniel", "Patrícia", "Rafael", "Aline",
		"Marcelo", "Renata", "Gustavo", "Cristina", "Eduardo", "Sandra", "Felipe",
		"Mônica", "Rodrigo", "Vanessa", "Thiago", "Priscila", "Leonardo", "Débora",
	}

	lastNames := []string{
		"Silva", "Santos", "Oliveira", "Souza", "Rodrigues", "Ferreira", "Alves",
		"Pereira", "Lima", "Gomes", "Costa", "Ribeiro", "Martins", "Carvalho",
		"Almeida", "Lopes", "Soares", "Fernandes", "Vieira", "Barbosa", "Rocha",
		"Dias", "Monteiro", "Cardoso", "Reis", "Araújo", "Cavalcanti", "Nascimento",
	}

	cities := []string{
		"São Paulo", "Rio de Janeiro", "Belo Horizonte", "Salvador", "Brasília",
		"Curitiba", "Recife", "Porto Alegre", "Manaus", "Belém", "Goiânia",
		"Guarulhos", "Campinas", "São Luís", "Maceió", "Natal", "Teresina",
		"João Pessoa", "Campo Grande", "Cuiabá", "Florianópolis", "Vitória",
	}

	countries := []string{
		"Brasil", "Argentina", "Chile", "Uruguai", "Paraguai", "Bolívia",
		"Peru", "Colômbia", "Venezuela", "Equador",
	}

	customers := make([]Customer, count)

	for i := 0; i < count; i++ {
		firstName := firstNames[rand.Intn(len(firstNames))]
		lastName := lastNames[rand.Intn(len(lastNames))]
		name := fmt.Sprintf("%s %s", firstName, lastName)

		// Gerar email único
		email := fmt.Sprintf("%s.%s%d@example.com",
			toLowerCase(firstName), toLowerCase(lastName), i+1)

		// Gerar telefone
		phone := fmt.Sprintf("(%02d) %04d-%04d",
			rand.Intn(99)+11, rand.Intn(9999)+1000, rand.Intn(9999)+1000)

		customers[i] = Customer{
			Name:             name,
			Email:            email,
			Phone:            phone,
			City:             cities[rand.Intn(len(cities))],
			Country:          countries[rand.Intn(len(countries))],
			RegistrationDate: randomDate(),
		}
	}

	return customers
}

// generateProducts gera uma lista de produtos fictícios
func generateProducts(count int) []Product {
	fmt.Printf("   🔧 Gerando %d produtos...\n", count)

	categories := []string{
		"Eletrônicos", "Roupas", "Casa e Jardim", "Esportes", "Livros",
		"Beleza", "Automotive", "Ferramentas", "Brinquedos", "Alimentação",
		"Saúde", "Música", "Informática", "Celulares", "Câmeras",
	}

	productNames := []string{
		"Smartphone", "Notebook", "Tablet", "Fone de Ouvido", "Camiseta",
		"Calça Jeans", "Tênis", "Relógio", "Óculos", "Mochila",
		"Mesa", "Cadeira", "Luminária", "Tapete", "Almofada",
		"Panela", "Frigideira", "Liquidificador", "Microondas", "Geladeira",
		"Bicicleta", "Patins", "Bola", "Raquete", "Halteres",
		"Livro", "Revista", "Caderno", "Caneta", "Lápis",
		"Perfume", "Shampoo", "Creme", "Batom", "Esmalte",
		"Pneu", "Óleo", "Filtro", "Bateria", "Lâmpada",
		"Furadeira", "Martelo", "Chave", "Parafuso", "Prego",
		"Boneca", "Carrinho", "Jogo", "Quebra-cabeça", "Pelúcia",
	}

	products := make([]Product, count)

	for i := 0; i < count; i++ {
		category := categories[rand.Intn(len(categories))]
		name := productNames[rand.Intn(len(productNames))]

		// Gerar SKU único
		sku := fmt.Sprintf("SKU-%s-%05d",
			category[:3], i+1)

		// Preços baseados na categoria
		var basePrice float64
		switch category {
		case "Eletrônicos", "Informática":
			basePrice = float64(rand.Intn(2000) + 500)
		case "Roupas", "Beleza":
			basePrice = float64(rand.Intn(200) + 50)
		case "Casa e Jardim":
			basePrice = float64(rand.Intn(500) + 100)
		default:
			basePrice = float64(rand.Intn(300) + 30)
		}

		cost := basePrice * (0.6 + rand.Float64()*0.2) // 60-80% do preço

		products[i] = Product{
			SKU:        sku,
			Name:       fmt.Sprintf("%s %s %d", name, category, i+1),
			Category:   category,
			Price:      roundPrice(basePrice),
			Cost:       roundPrice(cost),
			Stock:      rand.Intn(1000) + 10,
			Weight:     roundWeight(rand.Float64()*10 + 0.1),
			Dimensions: generateDimensions(),
		}
	}

	return products
}

// generateOrders gera uma lista de pedidos fictícios
func generateOrders(count, maxCustomerID, maxProductID int) []Order {
	fmt.Printf("   🔧 Gerando %d pedidos...\n", count)

	statuses := []string{"pending", "processing", "shipped", "delivered", "cancelled"}
	addresses := []string{
		"Rua das Flores, 123 - Centro",
		"Av. Paulista, 456 - Bela Vista",
		"Rua Augusta, 789 - Consolação",
		"Av. Brasil, 321 - Jardins",
		"Rua Oscar Freire, 654 - Cerqueira César",
		"Av. Faria Lima, 987 - Itaim Bibi",
		"Rua Estados Unidos, 147 - Jardim Paulista",
		"Av. Rebouças, 258 - Pinheiros",
		"Rua Consolação, 369 - República",
		"Av. Ipiranga, 741 - Centro",
	}

	orders := make([]Order, count)

	for i := 0; i < count; i++ {
		customerID := rand.Intn(maxCustomerID) + 1
		productID := rand.Intn(maxProductID) + 1
		quantity := rand.Intn(10) + 1
		unitPrice := float64(rand.Intn(500)+10) + rand.Float64()
		totalPrice := float64(quantity) * unitPrice

		orders[i] = Order{
			CustomerID:      customerID,
			ProductID:       productID,
			Quantity:        quantity,
			UnitPrice:       roundPrice(unitPrice),
			TotalPrice:      roundPrice(totalPrice),
			OrderDate:       randomDate(),
			Status:          statuses[rand.Intn(len(statuses))],
			ShippingAddress: addresses[rand.Intn(len(addresses))],
		}
	}

	return orders
}

// Funções auxiliares

func toLowerCase(s string) string {
	// Conversão simples para minúsculas (apenas para este exemplo)
	result := ""
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			result += string(r + 32)
		} else {
			result += string(r)
		}
	}
	return result
}

func randomDate() time.Time {
	min := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	max := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC).Unix()
	delta := max - min

	sec := rand.Int63n(delta) + min
	return time.Unix(sec, 0)
}

func roundPrice(price float64) float64 {
	return float64(int(price*100)) / 100
}

func roundWeight(weight float64) float64 {
	return float64(int(weight*1000)) / 1000
}

func generateDimensions() string {
	length := rand.Intn(100) + 10
	width := rand.Intn(100) + 10
	height := rand.Intn(50) + 5

	return fmt.Sprintf("%dx%dx%d cm", length, width, height)
}
