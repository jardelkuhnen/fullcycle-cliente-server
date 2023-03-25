package main

import (
	"context"
	"database/sql"
	_ "database/sql/driver"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/jardelkuhnen/cotacao-server/models"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"net/http"
	"time"
)

const (
	CREATE_TABLE       = "CREATE TABLE IF NOT EXISTS cotacao (id text PRIMARY KEY, code text, code_in text, name text, high text, low text, var_bid text, pct_change text, bid text, ask text, time_stamp text, create_date text)"
	SQL_INSERT_COTACAO = "INSERT INTO cotacao (id, code, code_in, name, high, low, var_bid, pct_change, bid, ask, time_stamp, create_date) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	URL_API_COTACAO    = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
)

var DB *sql.DB

func main() {
	DB = createDbConnection()
	mux := configureHandlers()
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}

func configureHandlers() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", GetContacao)

	return mux
}

func createDbConnection() *sql.DB {
	ctx, _ := context.WithTimeout(context.Background(), time.Millisecond*10)
	dbCon, err := sql.Open("sqlite3", "./cotacao.db")
	if err != nil {
		panic(err)
	}
	stmt, err := dbCon.PrepareContext(ctx, CREATE_TABLE)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		panic(err)
	}

	return dbCon
}

func GetContacao(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*200)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, URL_API_COTACAO, nil)
	checkErr(err, w, http.StatusInternalServerError, "Error creating request with time out")

	resp, err := http.DefaultClient.Do(req)
	checkErr(err, w, http.StatusInternalServerError, "Error getting response from cotation API")

	resultJson, err := io.ReadAll(resp.Body)
	checkErr(err, w, http.StatusInternalServerError, "Error getting body from response")
	defer resp.Body.Close()

	var cotacao models.Cotacao
	err = json.Unmarshal(resultJson, &cotacao)
	checkErr(err, w, http.StatusInternalServerError, "Error getting body from response")

	saveOnDatabase(ctx, w, &cotacao)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write([]byte(cotacao.USD.Bid))
	return
}

func saveOnDatabase(ctx context.Context, w http.ResponseWriter, cotacao *models.Cotacao) {
	ctxDatabase, cancel := context.WithTimeout(ctx, time.Millisecond*10)
	defer cancel()
	stmt, err := DB.PrepareContext(ctxDatabase, SQL_INSERT_COTACAO)
	checkErr(err, w, http.StatusInternalServerError, "Error on insert on database")
	_, err = stmt.ExecContext(ctx,
		uuid.New().String(),
		cotacao.USD.Code,
		cotacao.USD.Codein,
		cotacao.USD.Name,
		cotacao.USD.High,
		cotacao.USD.Low,
		cotacao.USD.VarBid,
		cotacao.USD.PctChange,
		cotacao.USD.Bid,
		cotacao.USD.Ask,
		cotacao.USD.Timestamp,
		cotacao.USD.CreateDate)
	checkErr(err, w, http.StatusInternalServerError, "Error on insert on database")
}

func checkErr(err error, w http.ResponseWriter, httpStatus int, messsage string) {
	if err != nil {
		w.WriteHeader(httpStatus)
		w.Write([]byte(messsage))
		return
	}
}
