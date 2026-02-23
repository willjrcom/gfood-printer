package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/streadway/amqp"
	"github.com/willjrcom/gfood-printer/api"
	"github.com/willjrcom/gfood-printer/internal/service/rabbitmq"
)

var (
	rabbitService *rabbitmq.RabbitMQ
	stopChan      chan struct{}
)

type PrintMessage struct {
	Id          string `json:"id"`
	PrinterName string `json:"printer_name"`
}

func startRabbitMQConsumer() {
	if GlobalConfig == nil {
		log.Println("RabbitMQ: Configuração global não definida")
		return
	}

	// Se já houver um consumidor rodando, para ele
	if stopChan != nil {
		close(stopChan)
		time.Sleep(1 * time.Second)
	}

	stopChan = make(chan struct{})

	go func() {
		for {
			select {
			case <-stopChan:
				log.Println("RabbitMQ: Parando consumidor")
				if rabbitService != nil {
					rabbitService.Close()
				}
				return
			default:
				if err := connectAndConsume(); err != nil {
					log.Printf("RabbitMQ: Erro na conexão (tentando em 5s): %v", err)
					time.Sleep(5 * time.Second)
					continue
				}
				return
			}
		}
	}()
}

func connectAndConsume() error {
	var err error
	rabbitService, err = rabbitmq.NewInstance(GlobalConfig.RabbitMQURL)
	if err != nil {
		return err
	}

	// Exchanges (Tópicos) para consumir
	exchanges := []string{rabbitmq.GROUP_ITEM_EX, rabbitmq.ORDER_EX, rabbitmq.ORDER_DELIVERY_EX}

	for _, ex := range exchanges {
		msgs, err := rabbitService.ConsumeMessages(ex, GlobalConfig.SchemaName)
		if err != nil {
			log.Printf("RabbitMQ: Erro ao consumir fila para %s: %v", ex, err)
			continue
		}

		go func(ex string, deliveries <-chan amqp.Delivery) {
			for d := range deliveries {
				var msg PrintMessage
				if err := json.Unmarshal(d.Body, &msg); err != nil {
					log.Printf("RabbitMQ: Erro ao decodificar mensagem: %v", err)
					d.Nack(false, false)
					continue
				}
				log.Printf("RabbitMQ: Mensagem recebida para %s: %s", ex, msg.Id)

				// Busca conteúdo via API isolada
				content, err := api.FetchPrintContent(GlobalConfig, ex, msg.Id)
				if err != nil {
					log.Printf("RabbitMQ: Erro ao buscar conteúdo para ID %s: %v", msg.Id, err)
					d.Nack(false, true) // Requeue em caso de erro de rede/api
					continue
				}

				// Imprime
				if err := printToOS(msg.PrinterName, content); err != nil {
					log.Printf("RabbitMQ: Erro ao imprimir conteúdo para ID %s: %v", msg.Id, err)
					d.Nack(false, true)
				} else {
					log.Printf("RabbitMQ: Impressão enviada com sucesso para ID: %s", msg.Id)
					d.Ack(false)
				}
			}
		}(ex, msgs)
	}

	// Mantém a conexão aberta até erro ou sinal de parada
	errChan := make(chan *amqp.Error)
	rabbitService.NotifyClose(errChan)

	select {
	case amqpErr := <-errChan:
		return amqpErr
	case <-stopChan:
		return nil
	}
}
