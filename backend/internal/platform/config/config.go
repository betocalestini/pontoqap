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
}

type HTTPConfig struct {
	Addr            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	CORSOrigins     []string
}

type SecurityConfig struct {
	SessionSecret string
	CSRFSecret    string
	EncryptionKey string
	CookieSecure  bool
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

type PaymentsConfig struct {
	Provider      string
	WebhookSecret string
}

func Load() (Config, error) {
	_ = godotenv.Load()

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
			SessionSecret: env("SESSION_SECRET", "dev-session-secret-change-in-production"),
			CSRFSecret:    env("CSRF_SECRET", "dev-csrf-secret-change-in-production"),
			EncryptionKey: env("ENCRYPTION_KEY", "dev-encryption-key-32-bytes!!"),
			CookieSecure:  env("COOKIE_SECURE", "false") == "true",
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
		},
		UploadDir: env("UPLOAD_DIR", "./data/uploads"),
	}

	if len(cfg.Security.SessionSecret) < 16 {
		return cfg, fmt.Errorf("SESSION_SECRET must be at least 16 characters")
	}
	return cfg, nil
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
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
