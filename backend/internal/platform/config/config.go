package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv      string
	LogLevel    string
	DatabaseURL string
	HTTP        HTTPConfig
	Security    SecurityConfig
	Session     SessionConfig
	Worker      WorkerConfig
	Payments    PaymentsConfig
	UploadDir   string
	SMTP        SMTPConfig
	App         AppConfig
	Customer    CustomerConfig
}

type SMTPConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

type AppConfig struct {
	StoreWebURL string
	AdminWebURL string
}

type CustomerConfig struct {
	DefaultCreditLimitCents int64
}

type HTTPConfig struct {
	Addr            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	CORSOrigins     []string
}

type SecurityConfig struct {
	SessionSecret     string
	CSRFSecret        string
	EncryptionKey     string
	CookieSecure      bool
	AdminMFARequired  bool
}

type SessionConfig struct {
	StoreCookie string
	AdminCookie string
	StoreTTL    time.Duration
	AdminTTL    time.Duration
}

type WorkerConfig struct {
	PollInterval time.Duration
	WorkerID     string
}

type MercadoPagoConfig struct {
	Environment     string
	BaseURL         string
	AccessToken     string
	WebhookSecret   string
	ApplicationID   string
	PixExpiration   string
	RequestTimeout  time.Duration
	TestAutoApprove bool
	WebhookDebug    bool
}

type PaymentsConfig struct {
	Provider      string
	WebhookSecret string
	MercadoPago   MercadoPagoConfig
}

func Load() (Config, error) {
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")

	cfg := Config{
		AppEnv:      env("APP_ENV", "development"),
		LogLevel:    env("LOG_LEVEL", "info"),
		DatabaseURL: env("DATABASE_URL", "postgres://store:store@localhost:5432/store?sslmode=disable"),
		HTTP: HTTPConfig{
			Addr:         env("HTTP_ADDR", ":8080"),
			ReadTimeout:  durationEnv("HTTP_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: durationEnv("HTTP_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:  durationEnv("HTTP_IDLE_TIMEOUT", 60*time.Second),
			CORSOrigins:  splitCSV(env("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:5174")),
		},
		Security: SecurityConfig{
			SessionSecret:    env("SESSION_SECRET", "dev-session-secret-change-in-production"),
			CSRFSecret:       env("CSRF_SECRET", "dev-csrf-secret-change-in-production"),
			EncryptionKey:    env("ENCRYPTION_KEY", "dev-encryption-key-32-bytes!!"),
			CookieSecure:     env("COOKIE_SECURE", "false") == "true",
			AdminMFARequired: envBool("ADMIN_MFA_REQUIRED", env("APP_ENV", "development") == "production"),
		},
		Session: SessionConfig{
			StoreCookie: env("STORE_SESSION_COOKIE", "store_session"),
			AdminCookie: env("ADMIN_SESSION_COOKIE", "admin_session"),
			StoreTTL:    durationEnv("SESSION_TTL_STORE", 24*time.Hour),
			AdminTTL:    durationEnv("SESSION_TTL_ADMIN", 8*time.Hour),
		},
		Worker: WorkerConfig{
			PollInterval: durationEnv("WORKER_POLL_INTERVAL", 5*time.Second),
			WorkerID:     env("WORKER_ID", "worker-1"),
		},
		Payments: PaymentsConfig{
			Provider:      env("PAYMENT_PROVIDER", "sandbox"),
			WebhookSecret: env("PAYMENT_WEBHOOK_SECRET", "sandbox-webhook-secret"),
			MercadoPago: MercadoPagoConfig{
				Environment:     env("MERCADO_PAGO_ENVIRONMENT", "test"),
				BaseURL:         strings.TrimRight(env("MERCADO_PAGO_BASE_URL", "https://api.mercadopago.com"), "/"),
				AccessToken:     strings.TrimSpace(env("MERCADO_PAGO_ACCESS_TOKEN", "")),
				WebhookSecret:   strings.TrimSpace(env("MERCADO_PAGO_WEBHOOK_SECRET", "")),
				ApplicationID:   env("MERCADO_PAGO_APPLICATION_ID", ""),
				PixExpiration:   env("MERCADO_PAGO_PIX_EXPIRATION", "PT24H"),
				RequestTimeout:  time.Duration(intEnv("MERCADO_PAGO_REQUEST_TIMEOUT_SECONDS", 10)) * time.Second,
				TestAutoApprove: envBool("MERCADO_PAGO_TEST_AUTO_APPROVE", false),
				WebhookDebug:    envBool("MERCADO_PAGO_WEBHOOK_DEBUG", false),
			},
		},
		UploadDir: env("UPLOAD_DIR", "internal/catalog/static"),
		SMTP: SMTPConfig{
			Host:     env("SMTP_HOST", "localhost"),
			Port:     intEnv("SMTP_PORT", 1025),
			User:     env("SMTP_USER", ""),
			Password: env("SMTP_PASSWORD", ""),
			From:     env("SMTP_FROM", "Store <noreply@store.local>"),
		},
		App: AppConfig{
			StoreWebURL: env("STORE_WEB_URL", "http://localhost:5173"),
			AdminWebURL: env("ADMIN_WEB_URL", "http://localhost:5174"),
		},
		Customer: CustomerConfig{
			DefaultCreditLimitCents: int64Env("DEFAULT_CUSTOMER_CREDIT_LIMIT_CENTS", 50_000),
		},
	}

	if len(cfg.Security.SessionSecret) < 16 {
		return cfg, fmt.Errorf("SESSION_SECRET must be at least 16 characters")
	}
	cfg.Payments.Provider = NormalizePaymentProvider(cfg.Payments.Provider)
	if IsMercadoPagoProvider(cfg.Payments.Provider) && cfg.Payments.MercadoPago.AccessToken == "" {
		return cfg, fmt.Errorf("MERCADO_PAGO_ACCESS_TOKEN is required when PAYMENT_PROVIDER=mercadopago")
	}
	mp := cfg.Payments.MercadoPago
	if mp.WebhookDebug && cfg.AppEnv == "production" {
		return cfg, fmt.Errorf("MERCADO_PAGO_WEBHOOK_DEBUG must be false in production")
	}
	if mp.TestAutoApprove {
		if cfg.AppEnv == "production" {
			return cfg, fmt.Errorf("MERCADO_PAGO_TEST_AUTO_APPROVE must be false in production")
		}
		if !strings.EqualFold(mp.Environment, "test") {
			return cfg, fmt.Errorf("MERCADO_PAGO_TEST_AUTO_APPROVE requires MERCADO_PAGO_ENVIRONMENT=test")
		}
	}
	return cfg, nil
}

// NormalizePaymentProvider maps aliases (e.g. mercado_pago) to canonical names.
func NormalizePaymentProvider(provider string) string {
	p := strings.ToLower(strings.TrimSpace(provider))
	switch p {
	case "mercado_pago", "mercadopago":
		return "mercadopago"
	default:
		return p
	}
}

// IsMercadoPagoProvider reports whether the provider uses Mercado Pago APIs.
func IsMercadoPagoProvider(provider string) bool {
	return NormalizePaymentProvider(provider) == "mercadopago"
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if v == "" {
		return fallback
	}
	return v == "1" || v == "true" || v == "yes"
}

func durationEnv(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func intEnv(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	var n int
	if _, err := fmt.Sscanf(v, "%d", &n); err != nil {
		return fallback
	}
	return n
}

func int64Env(key string, fallback int64) int64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	var n int64
	if _, err := fmt.Sscanf(v, "%d", &n); err != nil {
		return fallback
	}
	return n
}

func TrimTrailingSlash(s string) string {
	return strings.TrimRight(strings.TrimSpace(s), "/")
}
