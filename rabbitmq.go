package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/streadway/amqp"
	"github.com/willjrcom/gfood-printer/api"
	"github.com/willjrcom/gfood-printer/internal/service/rabbitmq"
)

const maxRetries = 3

var (
	rabbitService *rabbitmq.RabbitMQ
	stopChan      chan struct{}
	retryMap      sync.Map
)

type PrintMessage struct {
	Path        string `json:"path"`
	PrinterName string `json:"printer_name"`
}

// nackWithRetry faz Nack com requeue até maxRetries vezes; depois descarta.
func nackWithRetry(d amqp.Delivery, label string) {
	key := fmt.Sprintf("%x", md5.Sum(d.Body))
	val, _ := retryMap.LoadOrStore(key, 0)
	count := val.(int)

	if count >= maxRetries {
		log.Printf("RabbitMQ: Descartando mensagem após %d tentativas [%s]", maxRetries, label)
		retryMap.Delete(key)
		d.Nack(false, false)
		return
	}

	retryMap.Store(key, count+1)
	log.Printf("RabbitMQ: Tentativa %d/%d — requeue [%s]", count+1, maxRetries, label)
	d.Nack(false, true)
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
	exchanges := []string{rabbitmq.SHIFT_EX, rabbitmq.GROUP_ITEM_EX, rabbitmq.ORDER_EX}

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
					d.Nack(false, false) // JSON inválido: descarta direto
					continue
				}
				log.Printf("RabbitMQ: Mensagem recebida para %s: %s", ex, msg.Path)

				// Busca conteúdo via API (timeout de 10s interno no http.Client)
				content, err := api.FetchPrintContent(GlobalConfig, msg.Path)
				if err != nil {
					log.Printf("RabbitMQ: Erro ao buscar conteúdo para ID %s: %v", msg.Path, err)
					nackWithRetry(d, msg.Path)
					continue
				}

				// Imprime
				printerName := resolvePrinterName(msg.PrinterName)
				if err := printToOS(printerName, content); err != nil {
					log.Printf("RabbitMQ: Erro ao imprimir conteúdo para ID %s: %v", msg.Path, err)
					nackWithRetry(d, msg.Path)
				} else {
					log.Printf("RabbitMQ: Impressão enviada com sucesso para ID: %s", msg.Path)
					key := fmt.Sprintf("%x", md5.Sum(d.Body))
					retryMap.Delete(key)
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
