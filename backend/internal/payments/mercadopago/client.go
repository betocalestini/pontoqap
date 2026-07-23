package mercadopago

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type apiClient struct {
	baseURL     string
	accessToken string
	http        *http.Client
	log         *slog.Logger
}

func newAPIClient(cfg Config, log *slog.Logger) *apiClient {
	return &apiClient{
		baseURL:     strings.TrimRight(cfg.BaseURL, "/"),
		accessToken: cfg.AccessToken,
		http: &http.Client{
			Timeout: cfg.RequestTimeout,
		},
		log: log,
	}
}

func (c *apiClient) postJSON(ctx context.Context, path string, idempotencyKey string, body any, dest any) (int, error) {
	start := time.Now()
	payload, err := json.Marshal(body)
	if err != nil {
		return 0, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if idempotencyKey != "" {
		req.Header.Set("X-Idempotency-Key", idempotencyKey)
	}

	res, err := c.http.Do(req)
	if err != nil {
		c.logAPICall("POST", path, 0, "", time.Since(start), true)
		return 0, fmt.Errorf("mercado pago: request failed: %w", err)
	}
	defer res.Body.Close()

	mpReqID := res.Header.Get("x-request-id")

	raw, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		c.logAPICall("POST", path, res.StatusCode, mpReqID, time.Since(start), true)
		return res.StatusCode, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		c.logAPICall("POST", path, res.StatusCode, mpReqID, time.Since(start), true)
		return res.StatusCode, callError{status: res.StatusCode, mpRequestID: mpReqID, body: append([]byte(nil), raw...)}
	}
	if dest != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, dest); err != nil {
			c.logAPICall("POST", path, res.StatusCode, mpReqID, time.Since(start), true)
			return res.StatusCode, fmt.Errorf("mercado pago: decode response: %w", err)
		}
	}
	c.logAPICall("POST", path, res.StatusCode, mpReqID, time.Since(start), false)
	return res.StatusCode, nil
}

func (c *apiClient) getJSON(ctx context.Context, path string, dest any) (int, error) {
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Accept", "application/json")

	res, err := c.http.Do(req)
	if err != nil {
		c.logAPICall("GET", path, 0, "", time.Since(start), true)
		return 0, fmt.Errorf("mercado pago: request failed: %w", err)
	}
	defer res.Body.Close()

	mpReqID := res.Header.Get("x-request-id")
	raw, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		c.logAPICall("GET", path, res.StatusCode, mpReqID, time.Since(start), true)
		return res.StatusCode, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		c.logAPICall("GET", path, res.StatusCode, mpReqID, time.Since(start), true)
		return res.StatusCode, callError{status: res.StatusCode, mpRequestID: mpReqID, body: append([]byte(nil), raw...)}
	}
	if dest != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, dest); err != nil {
			c.logAPICall("GET", path, res.StatusCode, mpReqID, time.Since(start), true)
			return res.StatusCode, fmt.Errorf("mercado pago: decode response: %w", err)
		}
	}
	c.logAPICall("GET", path, res.StatusCode, mpReqID, time.Since(start), false)
	return res.StatusCode, nil
}

func (c *apiClient) logAPICall(method, path string, status int, mpRequestID string, dur time.Duration, failed bool) {
	if c.log == nil {
		return
	}
	attrs := []any{
		slog.String("method", method),
		slog.String("path", path),
		slog.Int("status", status),
		slog.Int64("duration_ms", dur.Milliseconds()),
	}
	if mpRequestID != "" {
		attrs = append(attrs, slog.String("mp_request_id", mpRequestID))
	}
	if failed {
		c.log.Warn("mercado pago api call", attrs...)
	} else {
		c.log.Info("mercado pago api call", attrs...)
	}
}

func truncateErrBody(b []byte) string {
	const max = 512
	s := strings.TrimSpace(string(b))
	if len(s) > max {
		return s[:max] + "…"
	}
	return s
}

func expirationFromISO(iso string) time.Time {
	d, err := parseISODuration(iso)
	if err != nil || d <= 0 {
		return time.Now().Add(24 * time.Hour)
	}
	return time.Now().Add(d)
}

// parseISODuration supports a subset of ISO 8601 durations used for Pix (e.g. PT24H, PT1H, P2D).
func parseISODuration(iso string) (time.Duration, error) {
	iso = strings.TrimSpace(iso)
	if iso == "" {
		return 0, fmt.Errorf("empty duration")
	}
	if !strings.HasPrefix(iso, "P") {
		return 0, fmt.Errorf("invalid duration")
	}
	rest := iso[1:]
	var days int
	if i := strings.Index(rest, "D"); i >= 0 {
		if _, err := fmt.Sscanf(rest[:i], "%d", &days); err != nil {
			return 0, err
		}
		rest = rest[i+1:]
	}
	var hours, minutes int
	if strings.HasPrefix(rest, "T") {
		tpart := rest[1:]
		if i := strings.Index(tpart, "H"); i >= 0 {
			if _, err := fmt.Sscanf(tpart[:i], "%d", &hours); err != nil {
				return 0, err
			}
			tpart = tpart[i+1:]
		}
		if i := strings.Index(tpart, "M"); i >= 0 {
			if _, err := fmt.Sscanf(tpart[:i], "%d", &minutes); err != nil {
				return 0, err
			}
		}
	}
	d := time.Duration(days)*24*time.Hour + time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute
	if d == 0 {
		return 0, fmt.Errorf("zero duration")
	}
	return d, nil
}
