package logging

import "strings"

// MaskEmail reduz identificação em logs (ex.: u***@example.com).
func MaskEmail(email string) string {
	email = strings.TrimSpace(email)
	at := strings.LastIndex(email, "@")
	if at <= 0 {
		return "***"
	}
	local := email[:at]
	domain := email[at:]
	if len(local) <= 1 {
		return "*" + domain
	}
	return local[:1] + "***" + domain
}
