// backend/internal/token/jwt.go

package token

import (
	"fmt"
	"github.com/Culturae-org/culturae/internal/model"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	secret string
}

func NewJWTService(secret string) *JWTService {
	return &JWTService{
		secret: secret,
	}
}

func (js *JWTService) GenerateAccessToken(session *model.Session, duration time.Duration) (string, error) {
	roles := []string{session.User.Role}
	claims := &model.SessionTokenClaims{
		USN:   session.User.Username,
		SID:   session.ID.String(),
		Roles: roles,
		VRS:   session.Variables,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			Subject:   session.User.PublicID,
			ID:        session.TokenID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(js.secret))
}

func (js *JWTService) ValidateToken(tokenStr string) (*model.SessionTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &model.SessionTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(js.secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*model.SessionTokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func (js *JWTService) ExtractTokenID(tokenStr string) (string, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, &model.SessionTokenClaims{})
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*model.SessionTokenClaims); ok {
		return claims.ID, nil
	}

	return "", fmt.Errorf("invalid token format")
}

