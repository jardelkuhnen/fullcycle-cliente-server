package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	c := http.Client{Timeout: time.Millisecond * 300}
	resp, err := c.Get("http://localhost:8080/cotacao")
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	writeToFile(string(body))
	fmt.Println(string(body))

}

func writeToFile(body string) {
	f, err := os.Create("arquivo.txt")
	if err != nil {
		panic(err)
	}
	_, err = f.WriteString(fmt.Sprintf("Dólar: { %s }", body))
	if err != nil {
		panic(err)
	}
}
