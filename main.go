package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/ws", wsHandler)

	fmt.Println("Print Agent rodando em ws://localhost:8089/ws")
	log.Fatal(http.ListenAndServe("localhost:8089", nil))
}
