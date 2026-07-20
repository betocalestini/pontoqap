package app

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/store-platform/store/internal/audit"
	"github.com/store-platform/store/internal/billing"
	"github.com/store-platform/store/internal/catalog"
	cataloghttp "github.com/store-platform/store/internal/catalog/transport/http"
	"github.com/store-platform/store/internal/customers"
	customershttp "github.com/store-platform/store/internal/customers/transport/http"
	"github.com/store-platform/store/internal/identity"
	identityhttp "github.com/store-platform/store/internal/identity/transport/http"
	"github.com/store-platform/store/internal/inventory"
	inventoryhttp "github.com/store-platform/store/internal/inventory/transport/http"
	"github.com/store-platform/store/internal/platform/config"
	"github.com/store-platform/store/internal/platform/database"
	"github.com/store-platform/store/internal/platform/httpx"
	"github.com/store-platform/store/internal/sales"
	saleshttp "github.com/store-platform/store/internal/sales/transport/http"
)

// NewRouter monta o roteador HTTP da API (usado por cmd/api e testes E2E).
func NewRouter(cfg config.Config, pool *pgxpool.Pool, idSvc *identity.Service, logger *slog.Logger) http.Handler {
	catalogHandler := cataloghttp.NewHandler(catalog.NewService(pool))
	customersHandler := customershttp.NewHandler(customers.NewService(pool))
	invSvc := inventory.NewService(pool)
	invHandler := inventoryhttp.NewHandler(invSvc)
	salesHandler := saleshttp.NewHandler(sales.NewService(pool, invSvc, billing.NewService(pool)))
	idHandler := identityhttp.NewHandler(idSvc, cfg.Session, cfg.Security)

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
