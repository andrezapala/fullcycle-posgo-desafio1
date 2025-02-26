package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Cotacao struct {
	Bid string `json:"bid"`
}

type ApiResponse struct {
	USDBRL Cotacao `json:"USDBRL"`
}

func fetchCotacao(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var apiResp ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", err
	}

	return apiResp.USDBRL.Bid, nil
}

func saveCotacao(ctx context.Context, db *sql.DB, bid string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	_, err := db.ExecContext(ctx, "INSERT INTO cotacoes (valor, data) VALUES (?, ?)", bid, time.Now())
	return err
}

func cotacaoHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	db, err := sql.Open("sqlite3", "cotacoes.db")
	if err != nil {
		http.Error(w, "Erro ao abrir o banco de dados", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS cotacoes (id INTEGER PRIMARY KEY, valor TEXT, data DATETIME)")
	if err != nil {
		http.Error(w, "Erro ao criar a tabela", http.StatusInternalServerError)
		return
	}

	bid, err := fetchCotacao(ctx)
	if err != nil {
		http.Error(w, "Erro ao buscar cotação", http.StatusInternalServerError)
		log.Println("Erro ao buscar cotação:", err)
		return
	}

	if err := saveCotacao(ctx, db, bid); err != nil {
		log.Println("Erro ao salvar no banco de dados:", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"bid": bid})
}

func main() {
	http.HandleFunc("/cotacao", cotacaoHandler)
	fmt.Println("Servidor rodando na porta 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
