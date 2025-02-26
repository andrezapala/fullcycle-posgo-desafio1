package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func fetchCotacao(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result["bid"], nil
}

func saveCotacaoToFile(value string) error {
	content := fmt.Sprintf("Dólar: %s", value)
	return ioutil.WriteFile("cotacao.txt", []byte(content), 0644)
}

func main() {
	ctx := context.Background()

	bid, err := fetchCotacao(ctx)
	if err != nil {
		log.Println("Erro ao obter cotação:", err)
		return
	}

	if err := saveCotacaoToFile(bid); err != nil {
		log.Println("Erro ao salvar cotação no arquivo:", err)
	}
}
