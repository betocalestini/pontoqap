package security

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const AdminJWTAudience = "admin"

type AdminClaims struct {
	jwt.RegisteredClaims
	SID string `json:"sid"`
}

func IssueAdminToken(secret string, userID, sessionID uuid.UUID, expiresAt time.Time) (string, error) {
	if secret == "" {
		return "", fmt.Errorf("jwt secret required")
	}
	now := time.Now()
	claims := AdminClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			Audience:  jwt.ClaimStrings{AdminJWTAudience},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		SID: sessionID.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ParseAdminToken(secret, raw string) (userID, sessionID uuid.UUID, err error) {
	if secret == "" {
		return uuid.Nil, uuid.Nil, fmt.Errorf("jwt secret required")
	}
	parsed, err := jwt.ParseWithClaims(raw, &AdminClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	}, jwt.WithAudience(AdminJWTAudience))
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	claims, ok := parsed.Claims.(*AdminClaims)
	if !ok || !parsed.Valid {
		return uuid.Nil, uuid.Nil, fmt.Errorf("invalid token")
	}
	userID, err = uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	sessionID, err = uuid.Parse(claims.SID)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	return userID, sessionID, nil
}
