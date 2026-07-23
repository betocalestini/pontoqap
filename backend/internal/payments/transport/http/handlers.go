package http

import (
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/store-platform/store/internal/audit"
	"github.com/store-platform/store/internal/payments"
	identityhttp "github.com/store-platform/store/internal/identity/transport/http"
	"github.com/store-platform/store/internal/platform/httpx"
)

type Handler struct {
	svc               *payments.Service
	audit             *audit.Service
	log               *slog.Logger
	detailedPixErrors bool
	webhookDebug      bool
}

func NewHandler(svc *payments.Service, auditSvc *audit.Service, log *slog.Logger, detailedPixErrors, webhookDebug bool) *Handler {
	if log == nil {
		log = slog.Default()
	}
	return &Handler{svc: svc, audit: auditSvc, log: log, detailedPixErrors: detailedPixErrors, webhookDebug: webhookDebug}
}

func (h *Handler) AdminRoutes(r chi.Router) {
	r.Post("/invoices/{id}/pix-charge", h.CreatePixCharge)
	r.Post("/payment-charges/{id}/sync", h.SyncPaymentCharge)
}

func (h *Handler) MeRoutes(r chi.Router) {
	r.Post("/invoices/{id}/pix-charge", h.CreatePixCharge)
	r.Get("/installments/{installmentId}/pix-charge", h.GetInstallmentPixCharge)
	r.Post("/installments/{installmentId}/pix-charge", h.CreateInstallmentPixCharge)
}

func (h *Handler) SyncPaymentCharge(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID inválido")
		return
	}
	if err := h.svc.SyncMercadoPagoChargeByID(r.Context(), id); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) WebhookRoutes(r chi.Router) {
	r.Post("/pix", h.WebhookPix)
	r.Post("/mercado-pago/orders", h.WebhookMercadoPagoOrders)
}

func (h *Handler) CreatePixCharge(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID inválido")
		return
	}
	charge, err := h.svc.CreateOrReusePixCharge(r.Context(), id)
	if err != nil {
		h.writePixChargeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, charge)
}

func (h *Handler) GetInstallmentPixCharge(w http.ResponseWriter, r *http.Request) {
	user := identityhttp.UserFromContext(r.Context())
	if user == nil || user.CustomerID == nil {
		httpx.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Cliente necessário")
		return
	}
	installmentID, err := uuid.Parse(chi.URLParam(r, "installmentId"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID inválido")
		return
	}
	charge, err := h.svc.GetPendingPixChargeForInstallment(r.Context(), installmentID, *user.CustomerID)
	if err != nil {
		if errors.Is(err, payments.ErrPixChargeNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Nenhum Pix pendente para esta parcela")
			return
		}
		h.writePixChargeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, charge)
}

func (h *Handler) CreateInstallmentPixCharge(w http.ResponseWriter, r *http.Request) {
	user := identityhttp.UserFromContext(r.Context())
	if user == nil || user.CustomerID == nil {
		httpx.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Cliente necessário")
		return
	}
	installmentID, err := uuid.Parse(chi.URLParam(r, "installmentId"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID inválido")
		return
	}
	charge, err := h.svc.CreateOrReusePixChargeForInstallment(r.Context(), installmentID, *user.CustomerID)
	if err != nil {
		h.writePixChargeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, charge)
}

func (h *Handler) WebhookPix(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "payload inválido")
		return
	}
	sig := r.Header.Get("X-Webhook-Signature")
	result, err := h.svc.ProcessWebhook(r.Context(), body, sig)
	reqID := httpx.RequestIDFromContext(r.Context())
	if err != nil {
		if errors.Is(err, payments.ErrInvalidWebhookSignature) {
			h.log.Warn("sandbox pix webhook rejected",
				slog.String("request_id", reqID),
				slog.String("reason", "invalid_signature"),
			)
			httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		h.log.Warn("sandbox pix webhook rejected",
			slog.String("request_id", reqID),
			slog.String("reason", err.Error()),
		)
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	if result.Ignored {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	h.log.Info("sandbox pix webhook received",
		slog.String("request_id", reqID),
		slog.String("external_event_id", result.ExternalEventID),
		slog.String("event_type", result.EventType),
		slog.Bool("duplicate", result.Duplicate),
		slog.Bool("settled", result.Settled),
		slog.String("invoice_id", result.InvoiceID.String()),
	)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) WebhookMercadoPagoOrders(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "payload inválido")
		return
	}
	mpRequestID := r.Header.Get("x-request-id")
	xSignature := r.Header.Get("x-signature")
	dataIDQuery := r.URL.Query().Get("data.id")
	if h.webhookDebug {
		cap := payments.BuildMercadoPagoWebhookCapture(
			r.URL.RawQuery,
			dataIDQuery,
			xSignature,
			mpRequestID,
			r.Host,
			r.ContentLength,
			body,
		)
		h.log.Info("mercado pago webhook debug capture",
			slog.String("raw_query", cap.RawQuery),
			slog.String("data_id_query", cap.DataIDQuery),
			slog.String("data_id_body", cap.DataIDBody),
			slog.String("x_signature", cap.XSignature),
			slog.String("x_request_id", cap.XRequestID),
			slog.String("host", cap.Host),
			slog.Int64("content_length", cap.ContentLength),
			slog.String("application_id", cap.ApplicationID),
			slog.String("capture_path", payments.DefaultMercadoPagoWebhookCapturePath),
		)
		if err := payments.WriteMercadoPagoWebhookCapture(payments.DefaultMercadoPagoWebhookCapturePath, cap); err != nil {
			h.log.Warn("mercado pago webhook debug capture write failed", slog.String("error", err.Error()))
		}
	}
	result, err := h.svc.ReceiveMercadoPagoOrderWebhook(r.Context(),
		xSignature,
		mpRequestID,
		dataIDQuery,
		body,
	)
	reqID := httpx.RequestIDFromContext(r.Context())
	if err != nil {
		if errors.Is(err, payments.ErrInvalidWebhookSignature) {
			h.log.Warn("mercado pago webhook rejected",
				slog.String("request_id", reqID),
				slog.String("mp_request_id", mpRequestID),
				slog.String("data_id_query", dataIDQuery),
				slog.String("reason", "invalid_signature"),
			)
			httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "assinatura inválida")
			return
		}
		h.log.Warn("mercado pago webhook rejected",
			slog.String("request_id", reqID),
			slog.String("mp_request_id", mpRequestID),
			slog.String("reason", err.Error()),
		)
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	h.log.Info("mercado pago webhook received",
		slog.String("request_id", reqID),
		slog.String("mp_request_id", mpRequestID),
		slog.String("order_id", result.OrderID),
		slog.String("external_event_id", result.EventID),
		slog.String("event_type", result.EventType),
		slog.Bool("duplicate", !result.Inserted),
	)
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// DevSimulate expõe confirmação sandbox (somente desenvolvimento).
func (h *Handler) DevSimulate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "chargeId"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ID inválido")
		return
	}
	if err := h.svc.SimulateSandboxPayment(r.Context(), id); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) writePixChargeError(w http.ResponseWriter, r *http.Request, err error) {
	ref := httpx.RequestIDFromContext(r.Context())
	msg := "Não foi possível gerar o Pix neste momento. Tente novamente ou procure o atendimento."
	if ref != "" {
		msg += " Código de referência: " + ref
	}
	if h.detailedPixErrors {
		msg += " (" + err.Error() + ")"
	}
	h.log.Warn("pix charge rejected for client", slog.String("request_id", ref), slog.String("reason", err.Error()))
	httpx.WriteError(w, http.StatusBadRequest, "PIX_UNAVAILABLE", msg)
}
