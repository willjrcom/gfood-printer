package api

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/willjrcom/gfood-printer/internal/service/rabbitmq"
)

type Config interface {
	GetAccessToken() string
	GetSchemaName() string
	GetBackendURL() string
}

func FetchPrintContent(config Config, ex, id string) (string, error) {
	var path string
	switch ex {
	case rabbitmq.GROUP_ITEM_EX:
		path = "/order-print/kitchen/" + id
	case rabbitmq.ORDER_DELIVERY_EX:
		// TODO: Confirmar se existe rota específica para delivery ou se usa a mesma de order
		path = "/order-print/" + id
	default:
		path = "/order-print/" + id
	}

	url := config.GetBackendURL() + path

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("erro ao criar requisição: %v", err)
	}

	req.Header.Set("access-token", config.GetAccessToken())

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro ao buscar dados no backend: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("erro no backend (Status %d): %s", resp.StatusCode, string(body))
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("erro ao ler corpo da resposta: %v", err)
	}

	return string(content), nil
}
