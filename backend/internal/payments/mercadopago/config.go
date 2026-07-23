package mercadopago

import (
	"time"

	"github.com/store-platform/store/internal/platform/config"
)

type Config struct {
	Environment     string
	BaseURL         string
	AccessToken     string
	WebhookSecret   string
	ApplicationID   string
	PixExpiration   string
	RequestTimeout  time.Duration
	TestAutoApprove bool
}

func ConfigFromPayments(cfg config.PaymentsConfig) Config {
	mp := cfg.MercadoPago
	if mp.RequestTimeout <= 0 {
		mp.RequestTimeout = 10 * time.Second
	}
	if mp.PixExpiration == "" {
		mp.PixExpiration = "PT24H"
	}
	if mp.BaseURL == "" {
		mp.BaseURL = "https://api.mercadopago.com"
	}
	return Config{
		Environment:     mp.Environment,
		BaseURL:         mp.BaseURL,
		AccessToken:     mp.AccessToken,
		WebhookSecret:   mp.WebhookSecret,
		ApplicationID:   mp.ApplicationID,
		PixExpiration:   mp.PixExpiration,
		RequestTimeout:  mp.RequestTimeout,
		TestAutoApprove: mp.TestAutoApprove,
	}
}
