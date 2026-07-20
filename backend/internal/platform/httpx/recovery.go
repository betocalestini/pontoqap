package httpx

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

func RecoveryMiddleware(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic recovered", "error", rec, "stack", string(debug.Stack()))
					WriteJSON(w, http.StatusInternalServerError, map[string]string{
						"code":    "INTERNAL_ERROR",
						"message": "Erro interno do servidor",
					})
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
