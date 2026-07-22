package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/store-platform/store/internal/app"
	"github.com/store-platform/store/internal/identity"
	identitypostgres "github.com/store-platform/store/internal/identity/postgres"
	"github.com/store-platform/store/internal/identity/security"
	"github.com/store-platform/store/internal/jobs"
	"github.com/store-platform/store/internal/platform/config"
	"github.com/store-platform/store/internal/platform/database"
	"github.com/store-platform/store/internal/platform/logging"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	logger := logging.New(cfg.LogLevel)

	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	idRepo := identitypostgres.NewRepository(pool)
	adminUsersRepo := identitypostgres.NewAdminUsersRepository(pool)
	if err := ensureBootstrapSystemAdmin(ctx, adminUsersRepo); err != nil {
		logger.Error("bootstrap admin failed", "error", err)
		os.Exit(1)
	}

	idSvc := identity.NewService(idRepo, cfg.Session.StoreTTL, cfg.Session.AdminTTL, cfg.Security.SessionSecret, logger)
	verifySvc := identity.NewVerificationService(pool, jobs.NewRepository(pool), cfg.App, cfg.Customer)
	handler := app.NewRouter(cfg, pool, idSvc, verifySvc, logger)

	srv := &http.Server{
		Addr:         cfg.HTTP.Addr,
		Handler:      handler,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	go func() {
		logger.Info("api listening", slog.String("addr", cfg.HTTP.Addr))
		logger.Info("mercado pago order webhook",
			slog.String("method", "POST"),
			slog.String("path", "/api/v1/webhooks/mercado-pago/orders"),
			slog.Bool("webhook_secret_configured", cfg.Payments.MercadoPago.WebhookSecret != ""),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}

func ensureBootstrapSystemAdmin(ctx context.Context, adminRepo *identitypostgres.AdminUsersRepository) error {
	exists, err := adminRepo.ExistsSystemAdmin(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	email := strings.TrimSpace(os.Getenv("ADMIN_BOOTSTRAP_EMAIL"))
	name := strings.TrimSpace(os.Getenv("ADMIN_BOOTSTRAP_NAME"))
	password := os.Getenv("ADMIN_BOOTSTRAP_PASSWORD")
	if email == "" {
		email = "admin@loja.local"
	}
	if name == "" {
		name = "Administrador"
	}
	if password == "" {
		password = "ChangeMe123!"
	}
	hash, err := security.HashPassword(password)
	if err != nil {
		return err
	}
	return adminRepo.CreateBootstrapSystemAdmin(ctx, email, name, hash)
}
