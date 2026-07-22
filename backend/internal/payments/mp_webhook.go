package payments

import (
	"encoding/json"
	"fmt"
)

type mercadoPagoOrderWebhook struct {
	Type   string `json:"type"`
	Action string `json:"action"`
	Data   struct {
		ID string `json:"id"`
	} `json:"data"`
}

func parseMercadoPagoOrderWebhook(body []byte) (mercadoPagoOrderWebhook, error) {
	var p mercadoPagoOrderWebhook
	if err := json.Unmarshal(body, &p); err != nil {
		return mercadoPagoOrderWebhook{}, err
	}
	if p.Type != "order" {
		return mercadoPagoOrderWebhook{}, fmt.Errorf("tipo de notificação inválido")
	}
	if p.Data.ID == "" {
		return mercadoPagoOrderWebhook{}, fmt.Errorf("data.id ausente")
	}
	return p, nil
}
