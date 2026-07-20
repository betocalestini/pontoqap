package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/store-platform/store/internal/audit"
	"github.com/store-platform/store/internal/billing"
	"github.com/store-platform/store/internal/catalog"
	cataloghttp "github.com/store-platform/store/internal/catalog/transport/http"
	"github.com/store-platform/store/internal/customers"
	customershttp "github.com/store-platform/store/internal/customers/transport/http"
	"github.com/store-platform/store/internal/identity"
	identitypostgres "github.com/store-platform/store/internal/identity/postgres"
	identityhttp "github.com/store-platform/store/internal/identity/transport/http"
	"github.com/store-platform/store/internal/identity/security"
	"github.com/store-platform/store/internal/inventory"
	inventoryhttp "github.com/store-platform/store/internal/inventory/transport/http"
	"github.com/store-platform/store/internal/platform/config"
	"github.com/store-platform/store/internal/platform/database"
	"github.com/store-platform/store/internal/platform/httpx"
	"github.com/store-platform/store/internal/platform/logging"
	"github.com/store-platform/store/internal/sales"
	saleshttp "github.com/store-platform/store/internal/sales/transport/http"
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
	hash, _ := security.HashPassword("ChangeMe123!")
	_ = idRepo.EnsureBootstrapManager(ctx, "gerente@loja.local", "Gerente Demo", hash)

	idSvc := identity.NewService(idRepo, cfg.Session.StoreTTL, cfg.Session.AdminTTL)
	idHandler := identityhttp.NewHandler(idSvc, cfg.Session, cfg.Security)

	catalogHandler := cataloghttp.NewHandler(catalog.NewService(pool))
	customersHandler := customershttp.NewHandler(customers.NewService(pool))
	invSvc := inventory.NewService(pool)
	invHandler := inventoryhttp.NewHandler(invSvc)
	salesHandler := saleshttp.NewHandler(sales.NewService(pool, invSvc, billing.NewService(pool)))

	_ = audit.NewService(pool)

	storeAuth := identityhttp.AuthMiddleware(idSvc, cfg.Session)
	adminAuth := adminAudienceMiddleware(identityhttp.AuthMiddleware(idSvc, cfg.Session))

	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(httpx.RequestIDMiddleware)
	r.Use(httpx.SecurityHeadersMiddleware)
	r.Use(httpx.RecoveryMiddleware(logger))
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.HTTP.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-App-Audience", "Idempotency-Key", "X-Request-ID"},
		AllowCredentials: true,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := database.Ping(r.Context(), pool); err != nil {
			httpx.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unhealthy"})
			return
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/api/v1", func(api chi.Router) {
		api.Route("/auth", idHandler.Routes)
		api.Route("/catalog", catalogHandler.PublicRoutes)
		api.Route("/customers", customersHandler.StoreRoutes)

		api.Group(func(me chi.Router) {
			me.Use(storeAuth)
			me.Use(identityhttp.RequireAuth)
			me.Route("/me", salesHandler.MeRoutes)
		})

		api.Route("/admin", func(admin chi.Router) {
			admin.Use(adminAuth)
			admin.Group(func(priv chi.Router) {
				priv.Use(identityhttp.RequireAuth)
				priv.With(identityhttp.RequirePermission("products.read")).Get("/categories", catalogHandler.ListCategoriesAdmin)
				priv.With(identityhttp.RequirePermission("products.write")).Post("/categories", catalogHandler.CreateCategory)
				priv.With(identityhttp.RequirePermission("products.read")).Get("/products", catalogHandler.ListProductsAdmin)
				priv.With(identityhttp.RequirePermission("products.write")).Post("/products", catalogHandler.CreateProduct)
				priv.With(identityhttp.RequirePermission("products.write")).Patch("/skus/{skuId}/price", catalogHandler.ChangePrice)
				priv.Route("/customers", func(cr chi.Router) {
					cr.With(identityhttp.RequirePermission("customers.read")).Get("/", customersHandler.List)
					cr.With(identityhttp.RequirePermission("customers.approve")).Patch("/{id}/approve", customersHandler.Approve)
					cr.With(identityhttp.RequirePermission("customers.change_limit")).Patch("/{id}/credit-limit", customersHandler.ChangeLimit)
				})
				priv.With(identityhttp.RequirePermission("inventory.adjust")).Post("/inventory/entries", invHandler.RegisterEntry)
			})
		})
	})

	srv := &http.Server{
		Addr:         cfg.HTTP.Addr,
		Handler:      r,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	go func() {
		logger.Info("api listening", slog.String("addr", cfg.HTTP.Addr))
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

func adminAudienceMiddleware(auth func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("X-App-Audience", "admin")
			next.ServeHTTP(w, r)
		}))
	}
}
