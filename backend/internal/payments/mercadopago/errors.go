package mercadopago

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// HTTPStatusError is returned for Mercado Pago API responses that should surface to operators.
type HTTPStatusError struct {
	Status       int
	Message      string
	MPRequestID  string
	ErrorCode    string
	ErrorMessage string
}

func (e HTTPStatusError) Error() string {
	return e.Message
}

type callError struct {
	status      int
	mpRequestID string
	body        []byte
}

func (e callError) Error() string {
	return fmt.Sprintf("mercado pago: HTTP %d: %s", e.status, truncateErrBody(e.body))
}

func parseMPErrorBody(body []byte) (code, message string) {
	if len(body) == 0 {
		return "", ""
	}
	var v struct {
		Error   string `json:"error"`
		Message string `json:"message"`
		Cause   []struct {
			Code        string `json:"code"`
			Description string `json:"description"`
		} `json:"cause"`
	}
	if err := json.Unmarshal(body, &v); err != nil {
		return "", truncateErrBody(body)
	}
	code = strings.TrimSpace(v.Error)
	message = strings.TrimSpace(v.Message)
	if code == "" && len(v.Cause) > 0 {
		code = strings.TrimSpace(v.Cause[0].Code)
		if message == "" {
			message = strings.TrimSpace(v.Cause[0].Description)
		}
	}
	return code, message
}

func httpStatusFromCall(err error, userMessage string) error {
	var call callError
	if !errors.As(err, &call) {
		return err
	}
	code, mpMsg := parseMPErrorBody(call.body)
	return HTTPStatusError{
		Status:       call.status,
		Message:      userMessage,
		MPRequestID:  call.mpRequestID,
		ErrorCode:    code,
		ErrorMessage: mpMsg,
	}
}
