package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Response struct {
	USDBRL struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

func main() {
	http.HandleFunc("/cotacao", handler)
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest(http.MethodGet, "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req = req.WithContext(ctx)
	c := &http.Client{}
	res, err := c.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var result Response
	if err := json.Unmarshal(body, &result); err != nil {
		log.Fatal(err)
	}

	err = insertData(ctx, result.USDBRL.Bid)
	if err != nil {
		log.Fatal(err)
	}

	w.Write([]byte(result.USDBRL.Bid))
}

func insertData(ctx context.Context, bid string) error {
	db, err := sql.Open("sqlite3", "./quotations.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	query := `
	CREATE TABLE IF NOT EXISTS quotations(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			price TEXT NOT NULL,
    	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	stmt, err := db.PrepareContext(ctx, "INSERT INTO quotations (price) VALUES (?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, bid)
	if err != nil {
		return err
	}

	return nil
}
