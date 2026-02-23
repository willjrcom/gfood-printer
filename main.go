package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/ws", wsHandler)

	fmt.Println("Print Agent rodando na porta :8089")
	log.Fatal(http.ListenAndServe(":8089", nil))
}
