package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Request/Response
type Request struct {
	Action string      `json:"action"`
	Data   interface{} `json:"data"`
}

type Response struct {
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Erro no websocket:", err)
		return
	}
	defer conn.Close()

	log.Println("Cliente conectado")

	for {
		var req Request
		if err := conn.ReadJSON(&req); err != nil {
			log.Println("Cliente desconectou:", err)
			return
		}

		switch req.Action {
		case "ping":
			conn.WriteJSON(Response{Status: "ok", Message: "pong"})

		case "get_printers":
			printers, err := getPrinters()
			if err != nil {
				conn.WriteJSON(Response{Status: "error", Message: err.Error()})
				continue
			}
			conn.WriteJSON(Response{Status: "ok", Data: printers})

		case "print":
			// Validar Data
			dataMap, ok := req.Data.(map[string]interface{})
			if !ok {
				conn.WriteJSON(Response{Status: "error", Message: "Formato inválido em Data (esperado objeto)"})
				continue
			}

			text, _ := dataMap["text"].(string)
			if text == "" {
				conn.WriteJSON(Response{Status: "error", Message: "Campo 'text' é obrigatório"})
				continue
			}

			printerName, _ := dataMap["printer"].(string)
			if printerName == "" {
				printerName = "default"
			}

			// Envia para impressão
			if err := printToOS(printerName, text); err != nil {
				conn.WriteJSON(Response{Status: "error", Message: err.Error()})
				continue
			}

			conn.WriteJSON(Response{Status: "ok", Message: "Impressão enviada"})

		default:
			conn.WriteJSON(Response{Status: "error", Message: fmt.Sprintf("Ação desconhecida: %s", req.Action)})
		}
	}
}
