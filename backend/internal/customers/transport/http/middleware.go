package http

import (
	"net/http"

	"github.com/store-platform/store/internal/customers"
	identityhttp "github.com/store-platform/store/internal/identity/transport/http"
	"github.com/store-platform/store/internal/platform/httpx"
)

// RequireStoreCustomerActive blocks /me/* for blocked or non-approved store customers.
func RequireStoreCustomerActive(svc *customers.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := identityhttp.UserFromContext(r.Context())
			if user == nil || user.CustomerID == nil {
				next.ServeHTTP(w, r)
				return
			}
			cust, err := svc.GetByID(r.Context(), *user.CustomerID)
			if err != nil || cust == nil {
				httpx.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Perfil de cliente necessário")
				return
			}
			if err := svc.EnsureNotBlocked(cust); err != nil {
				writeCustomerAppError(w, err)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func writeCustomerAppError(w http.ResponseWriter, err error) {
	if ae := customers.AsAppError(err); ae != nil {
		httpx.WriteError(w, ae.Status, ae.Code, ae.Message)
		return
	}
	httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno")
}
