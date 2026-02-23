package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Config struct {
	AccessToken string `json:"access_token"`
	SchemaName  string `json:"schema_name"`
	BackendURL  string `json:"backend_url"`
	RabbitMQURL string `json:"rabbitmq_url"`
}

func (c *Config) GetAccessToken() string { return c.AccessToken }
func (c *Config) GetSchemaName() string  { return c.SchemaName }
func (c *Config) GetBackendURL() string  { return c.BackendURL }

var (
	GlobalConfig *Config
	upgrader     = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

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
		log.Printf("Erro ao fazer upgrade para WebSocket: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("Nova conexão WebSocket estabelecida de: %s", r.RemoteAddr)

	for {
		var req Request
		if err := conn.ReadJSON(&req); err != nil {
			log.Printf("Conexão fechada ou erro ao ler JSON de %s: %v", r.RemoteAddr, err)
			return
		}

		log.Printf("WebSocket: Ação recebida [%s] de %s", req.Action, r.RemoteAddr)

		switch req.Action {
		case "ping":
			log.Printf("WebSocket: Respondendo ping de %s", r.RemoteAddr)
			conn.WriteJSON(Response{Status: "ok", Message: "pong"})

		case "get_printers":
			printers, err := getPrinters()
			if err != nil {
				log.Printf("WebSocket: Erro ao listar impressoras: %v", err)
				conn.WriteJSON(Response{Status: "error", Message: err.Error()})
				continue
			}
			log.Printf("WebSocket: %d impressoras listadas para %s", len(printers), r.RemoteAddr)
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
			printerName = resolvePrinterName(printerName)

			// Envia para impressão
			log.Printf("WebSocket: Solicitando impressão na impressora [%s] (tamanho texto: %d)", printerName, len(text))
			if err := printToOS(printerName, text); err != nil {
				log.Printf("WebSocket: Erro ao imprimir: %v", err)
				conn.WriteJSON(Response{Status: "error", Message: err.Error()})
				continue
			}

			log.Printf("WebSocket: Impressão enviada com sucesso para [%s]", printerName)
			conn.WriteJSON(Response{Status: "ok", Message: "Impressão enviada"})

		case "config":
			dataMap, ok := req.Data.(map[string]interface{})
			if !ok {
				conn.WriteJSON(Response{Status: "error", Message: "Formato inválido em Data (esperado objeto)"})
				continue
			}

			config := &Config{
				AccessToken: dataMap["access_token"].(string),
				SchemaName:  dataMap["schema_name"].(string),
				BackendURL:  dataMap["backend_url"].(string),
				RabbitMQURL: dataMap["rabbitmq_url"].(string),
			}

			if config.AccessToken == "" || config.SchemaName == "" || config.BackendURL == "" || config.RabbitMQURL == "" {
				conn.WriteJSON(Response{Status: "error", Message: "Configuração incompleta. access_token, schema_name, backend_url e rabbitmq_url são obrigatórios."})
				continue
			}

			log.Printf("WebSocket: Iniciando aplicação de configuração para schema [%s]", config.SchemaName)
			GlobalConfig = config
			log.Printf("WebSocket: Configuração aplicada com sucesso. Backend: %s, RabbitMQ: %s", GlobalConfig.BackendURL, GlobalConfig.RabbitMQURL)

			// Inicia ou reinicia o consumidor RabbitMQ
			log.Printf("WebSocket: Disparando início do consumidor RabbitMQ...")
			go startRabbitMQConsumer()

			conn.WriteJSON(Response{Status: "ok", Message: "Configuração aplicada"})

		default:
			conn.WriteJSON(Response{Status: "error", Message: fmt.Sprintf("Ação desconhecida: %s", req.Action)})
		}
	}
}
