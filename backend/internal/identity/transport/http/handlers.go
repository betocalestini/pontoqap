package http

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/store-platform/store/internal/identity"
	"github.com/store-platform/store/internal/platform/config"
	"github.com/store-platform/store/internal/platform/httpx"
)

type Handler struct {
	svc    *identity.Service
	verify *identity.VerificationService
	cfg    config.SessionConfig
	secure config.SecurityConfig
}

func NewHandler(svc *identity.Service, verify *identity.VerificationService, sessionCfg config.SessionConfig, sec config.SecurityConfig) *Handler {
	return &Handler{svc: svc, verify: verify, cfg: sessionCfg, secure: sec}
}

func (h *Handler) Routes(r chi.Router) {
	r.Post("/login", h.login)
	r.Get("/verify-email", h.verifyEmail)
	r.Post("/resend-verification", h.resendVerification)
}

func (h *Handler) AuthenticatedRoutes(r chi.Router) {
	r.Post("/logout", h.logout)
	r.Get("/me", h.me)
	r.Post("/mfa/setup", h.mfaSetup)
	r.Post("/mfa/verify", h.mfaVerify)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Audience string `json:"audience"`
	MFACode  string `json:"mfa_code"`
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Corpo da requisição inválido")
		return
	}
	if req.Audience == "" {
		req.Audience = audienceFromHeader(r)
	}
	res, err := h.svc.Login(r.Context(), identity.LoginInput{
		Email:     strings.TrimSpace(req.Email),
		Password:  req.Password,
		Audience:  req.Audience,
		MFACode:   req.MFACode,
		IP:        clientIP(r),
		UserAgent: r.UserAgent(),
	})
	if err != nil {
		writeAppError(w, err)
		return
	}
	if res.MFARequired {
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"mfa_required": true})
		return
	}
	if req.Audience != "admin" {
		setSessionCookie(w, h.cookieName(req.Audience), res.SessionToken, h.cookieTTL(req.Audience), h.secure.CookieSecure)
	}
	payload := mapUser(res.User)
	if req.Audience == "admin" {
		payload["access_token"] = res.AccessToken
		payload["expires_at"] = res.ExpiresAt.UTC().Format(time.RFC3339)
	}
	if req.Audience == "admin" && h.secure.AdminMFARequired && !res.User.User.MFAEnabled {
		payload["mfa_setup_required"] = true
	}
	httpx.WriteJSON(w, http.StatusOK, payload)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	aud := audienceFromHeader(r)
	if aud == "admin" {
		if bearer := bearerToken(r); bearer != "" {
			_ = h.svc.LogoutAdminBearer(r.Context(), bearer)
		}
		clearSessionCookie(w, h.cookieName(aud), h.secure.CookieSecure)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	token := readSessionCookie(r, h.cookieName(aud))
	if token != "" {
		_ = h.svc.Logout(r.Context(), token)
	}
	clearSessionCookie(w, h.cookieName(aud), h.secure.CookieSecure)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r.Context())
	if user == nil {
		httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Não autenticado")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, mapUser(*user))
}

func (h *Handler) mfaSetup(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r.Context())
	if user == nil {
		httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Não autenticado")
		return
	}
	secret, uri, err := h.svc.SetupMFA(r.Context(), user.User.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Falha ao configurar MFA")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"secret": secret, "otpauth_url": uri})
}

func (h *Handler) mfaVerify(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r.Context())
	if user == nil {
		httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Não autenticado")
		return
	}
	var body struct {
		Code string `json:"code"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Código inválido")
		return
	}
	if err := h.svc.VerifyMFA(r.Context(), user.User.ID, body.Code); err != nil {
		writeAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) verifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Token obrigatório")
		return
	}
	if h.verify == nil {
		httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Verificação indisponível")
		return
	}
	if err := h.verify.VerifyEmail(r.Context(), token); err != nil {
		writeAppError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "verified"})
}

func (h *Handler) resendVerification(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil || body.Email == "" {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "E-mail obrigatório")
		return
	}
	if h.verify != nil {
		_ = h.verify.ResendVerification(r.Context(), strings.TrimSpace(body.Email))
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) cookieName(audience string) string {
	if audience == "admin" {
		return h.cfg.AdminCookie
	}
	return h.cfg.StoreCookie
}

func (h *Handler) cookieTTL(audience string) int {
	if audience == "admin" {
		return int(h.cfg.AdminTTL.Seconds())
	}
	return int(h.cfg.StoreTTL.Seconds())
}

func audienceFromHeader(r *http.Request) string {
	switch strings.ToLower(r.Header.Get("X-App-Audience")) {
	case "admin":
		return "admin"
	default:
		return "store"
	}
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	return strings.Split(r.RemoteAddr, ":")[0]
}

func mapUser(u identity.AuthUser) map[string]any {
	out := map[string]any{
		"id":          u.User.ID,
		"name":        u.User.Name,
		"email":       u.User.Email,
		"roles":       u.Roles,
		"permissions": u.Permissions,
		"mfa_enabled": u.User.MFAEnabled,
	}
	if u.CustomerID != nil {
		out["customer_id"] = u.CustomerID
	}
	return out
}

func writeAppError(w http.ResponseWriter, err error) {
	if ae := identity.AsAppError(err); ae != nil {
		httpx.WriteError(w, ae.Status, ae.Code, ae.Message)
		return
	}
	httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Erro interno")
}

type ctxKey int

const userCtxKey ctxKey = 1

func UserFromContext(ctx context.Context) *identity.AuthUser {
	if v, ok := ctx.Value(userCtxKey).(*identity.AuthUser); ok {
		return v
	}
	return nil
}

// ContextWithAuthUser injeta o usuário autenticado (útil em testes HTTP).
func ContextWithAuthUser(ctx context.Context, user *identity.AuthUser) context.Context {
	return context.WithValue(ctx, userCtxKey, user)
}

func AuthMiddleware(svc *identity.Service, sessionCfg config.SessionConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			aud := audienceFromHeader(r)
			if aud == "admin" {
				if bearer := bearerToken(r); bearer != "" {
					user, err := svc.AuthenticateAdminBearer(r.Context(), bearer)
					if err == nil {
						ctx := context.WithValue(r.Context(), userCtxKey, user)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
				}
				next.ServeHTTP(w, r)
				return
			}
			cookieName := sessionCfg.StoreCookie
			token := readSessionCookie(r, cookieName)
			if token == "" {
				next.ServeHTTP(w, r)
				return
			}
			user, err := svc.AuthenticateSession(r.Context(), token, aud)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			ctx := context.WithValue(r.Context(), userCtxKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if UserFromContext(r.Context()) == nil {
			httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Não autenticado")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireAdminMFA bloqueia rotas administrativas até o TOTP estar ativo.
func RequireAdminMFA(required bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !required {
				next.ServeHTTP(w, r)
				return
			}
			user := UserFromContext(r.Context())
			if user == nil {
				httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Não autenticado")
				return
			}
			if !user.User.MFAEnabled {
				httpx.WriteError(w, http.StatusForbidden, "MFA_REQUIRED", "Configure MFA antes de acessar o painel")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequirePermission(code string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !HasPermission(r.Context(), code) {
				user := UserFromContext(r.Context())
				if user == nil {
					httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Não autenticado")
					return
				}
				httpx.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Permissão insuficiente")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// HasPermission reports whether the authenticated user has a permission code.
func HasPermission(ctx context.Context, code string) bool {
	user := UserFromContext(ctx)
	if user == nil {
		return false
	}
	for _, p := range user.Permissions {
		if p == code {
			return true
		}
	}
	return false
}

func setSessionCookie(w http.ResponseWriter, name, value string, maxAge int, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearSessionCookie(w http.ResponseWriter, name string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func readSessionCookie(r *http.Request, name string) string {
	c, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return c.Value
}

func bearerToken(r *http.Request) string {
	const prefix = "Bearer "
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, prefix) {
		return ""
	}
	return strings.TrimSpace(h[len(prefix):])
}
