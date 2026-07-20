package app

import (
	"log/slog"
	"net/http"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/store-platform/store/internal/audit"
	"github.com/store-platform/store/internal/billing"
	billinghttp "github.com/store-platform/store/internal/billing/transport/http"
	"github.com/store-platform/store/internal/catalog"
	cataloghttp "github.com/store-platform/store/internal/catalog/transport/http"
	"github.com/store-platform/store/internal/customers"
	customershttp "github.com/store-platform/store/internal/customers/transport/http"
	"github.com/store-platform/store/internal/forecasting"
	"github.com/store-platform/store/internal/identity"
	identityhttp "github.com/store-platform/store/internal/identity/transport/http"
	"github.com/store-platform/store/internal/inventory"
	inventoryhttp "github.com/store-platform/store/internal/inventory/transport/http"
	"github.com/store-platform/store/internal/jobs"
	"github.com/store-platform/store/internal/payments"
	paymentshttp "github.com/store-platform/store/internal/payments/transport/http"
	"github.com/store-platform/store/internal/platform/config"
	"github.com/store-platform/store/internal/platform/database"
	"github.com/store-platform/store/internal/platform/httpx"
	"github.com/store-platform/store/internal/reports"
	reportshttp "github.com/store-platform/store/internal/reports/transport/http"
	"github.com/store-platform/store/internal/sales"
	saleshttp "github.com/store-platform/store/internal/sales/transport/http"
)

// NewRouter monta o roteador HTTP da API (usado por cmd/api e testes E2E).
func NewRouter(cfg config.Config, pool *pgxpool.Pool, idSvc *identity.Service, verifySvc *identity.VerificationService, logger *slog.Logger) http.Handler {
	jobRepo := jobs.NewRepository(pool)
	billSvc := billing.NewService(pool, jobRepo, cfg.App.StoreWebURL)
	invSvc := inventory.NewService(pool)
	uploadRoot := cfg.UploadDir
	if abs, err := filepath.Abs(uploadRoot); err == nil {
		uploadRoot = abs
	}
	catalog.SetProductImagesUploadDir(uploadRoot)
	logger.Info("product images disk root", "upload_dir", uploadRoot, "images_dir", filepath.Join(uploadRoot, "product-images"))
	catalogHandler := cataloghttp.NewHandler(catalog.NewService(pool), invSvc, uploadRoot)
	customersHandler := customershttp.NewHandler(customers.NewService(pool, verifySvc))
	invHandler := inventoryhttp.NewHandler(invSvc)
	salesHandler := saleshttp.NewHandler(sales.NewService(pool, invSvc, billSvc))
	idHandler := identityhttp.NewHandler(idSvc, verifySvc, cfg.Session, cfg.Security)

	gateway := payments.NewSandboxGateway(cfg.Payments.WebhookSecret)
	paySvc := payments.NewService(pool, gateway, billSvc, cfg.Payments.WebhookSecret)
	payHandler := paymentshttp.NewHandler(paySvc)
	billHandler := billinghttp.NewHandler(billSvc)
	reportsHandler := reportshttp.NewReportsHandler(reports.NewService(pool))
	forecastHandler := reportshttp.NewForecastHandler(forecasting.NewService(pool))

	_ = audit.NewService(pool)

	storeAuth := identityhttp.AuthMiddleware(idSvc, cfg.Session)
	adminAuth := AdminAudienceMiddleware(identityhttp.AuthMiddleware(idSvc, cfg.Session))

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
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-App-Audience", "Idempotency-Key", "X-Request-ID", "X-Webhook-Signature"},
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
		api.Route("/auth", func(ar chi.Router) {
			idHandler.Routes(ar)
			ar.Group(func(authed chi.Router) {
				authed.Use(identityhttp.AuthMiddleware(idSvc, cfg.Session))
				authed.Use(identityhttp.RequireAuth)
				idHandler.AuthenticatedRoutes(authed)
			})
		})
		api.Route("/catalog", catalogHandler.PublicRoutes)
		api.Route("/customers", customersHandler.StoreRoutes)
		api.Route("/webhooks", payHandler.WebhookRoutes)

		api.Group(func(me chi.Router) {
			me.Use(storeAuth)
			me.Use(identityhttp.RequireAuth)
			me.Route("/me", func(m chi.Router) {
				salesHandler.MeRoutes(m)
				billHandler.MeRoutes(m)
				payHandler.MeRoutes(m)
			})
		})

		api.Route("/admin", func(admin chi.Router) {
			admin.Use(adminAuth)
			admin.Group(func(priv chi.Router) {
				priv.Use(identityhttp.RequireAuth)
				priv.Use(identityhttp.RequireAdminMFA(cfg.Security.AdminMFARequired))
				priv.With(identityhttp.RequirePermission("products.read")).Get("/categories", catalogHandler.ListCategoriesAdmin)
				priv.With(identityhttp.RequirePermission("products.write")).Post("/categories", catalogHandler.CreateCategory)
				priv.With(identityhttp.RequirePermission("products.read")).Get("/products", catalogHandler.ListProductsAdmin)
				priv.With(identityhttp.RequirePermission("products.read")).Get("/products/{id}", catalogHandler.GetProductAdmin)
				priv.With(identityhttp.RequirePermission("products.write")).Post("/products", catalogHandler.CreateProduct)
				priv.With(identityhttp.RequirePermission("products.write")).Patch("/products/{id}", catalogHandler.UpdateProduct)
				priv.With(identityhttp.RequirePermission("products.write")).Post("/products/{id}/images", catalogHandler.UploadProductImage)
				priv.With(identityhttp.RequirePermission("products.write")).Delete("/products/{id}/images/{imageId}", catalogHandler.DeleteProductImage)
				priv.With(identityhttp.RequirePermission("products.write")).Patch("/skus/{skuId}", catalogHandler.UpdateSKU)
				priv.With(identityhttp.RequirePermission("products.write")).Patch("/skus/{skuId}/price", catalogHandler.ChangePrice)
				priv.Route("/customers", func(cr chi.Router) {
					cr.With(identityhttp.RequirePermission("customers.read")).Get("/", customersHandler.List)
					cr.With(identityhttp.RequirePermission("customers.approve")).Patch("/{id}/approve", customersHandler.Approve)
					cr.With(identityhttp.RequirePermission("customers.change_limit")).Patch("/{id}/credit-limit", customersHandler.ChangeLimit)
				})
				priv.With(identityhttp.RequirePermission("inventory.read")).Get("/inventory/balances", invHandler.ListBalances)
				priv.With(identityhttp.RequirePermission("inventory.read")).Get("/inventory/movements", invHandler.ListMovements)
				priv.Post("/inventory/movements", invHandler.CreateMovement)
				priv.With(identityhttp.RequirePermission("inventory.entry")).Post("/inventory/entries", invHandler.RegisterEntry)
				priv.Route("/billing", func(br chi.Router) {
					br.With(identityhttp.RequirePermission("billing.read")).Get("/calendar", billHandler.ListCalendar)
					br.With(identityhttp.RequirePermission("settings.write")).Put("/calendar", billHandler.UpsertCalendar)
					br.With(identityhttp.RequirePermission("billing.close")).Post("/close", billHandler.ClosePeriods)
					br.With(identityhttp.RequirePermission("billing.read")).Get("/invoices", billHandler.ListAllInvoices)
				})
				priv.With(identityhttp.RequirePermission("payments.read")).Post("/invoices/{id}/pix-charge", payHandler.CreatePixCharge)
				priv.Route("/reports", func(rr chi.Router) {
					rr.With(identityhttp.RequirePermission("reports.read")).Get("/dashboard", reportsHandler.Dashboard)
					rr.With(identityhttp.RequirePermission("reports.read")).Get("/top-products", reportsHandler.TopProducts)
					rr.With(identityhttp.RequirePermission("reports.read")).Get("/inventory", reportsHandler.Inventory)
					rr.With(identityhttp.RequirePermission("reports.read")).Get("/forecast", forecastHandler.List)
					rr.With(identityhttp.RequirePermission("reports.read")).Post("/forecast/generate", forecastHandler.Generate)
				})
			})
		})

		if cfg.AppEnv == "development" {
			api.Post("/dev/pix/simulate/{chargeId}", payHandler.DevSimulate)
		}
	})

	return r
}

func AdminAudienceMiddleware(auth func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("X-App-Audience", "admin")
			auth(next).ServeHTTP(w, r)
		})
	}
}
