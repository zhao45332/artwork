package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID    int64  `json:"user_id"`
	Role      int32  `json:"role"`
	SessionID string `json:"session_id"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	secretKey           []byte
	accessTokenDuration time.Duration
	refreshTokenTTL     time.Duration
}

func NewTokenManager(secret string, accessTokenDuration, refreshTokenTTL time.Duration) (*TokenManager, error) {
	trimmed := strings.TrimSpace(secret)
	if trimmed == "" {
		return nil, errors.New("jwt secret 不能为空")
	}
	if accessTokenDuration <= 0 {
		accessTokenDuration = 2 * time.Hour
	}
	if refreshTokenTTL <= 0 {
		refreshTokenTTL = 7 * 24 * time.Hour
	}
	return &TokenManager{
		secretKey:           []byte(trimmed),
		accessTokenDuration: accessTokenDuration,
		refreshTokenTTL:     refreshTokenTTL,
	}, nil
}

func (m *TokenManager) GenerateAccessToken(userID int64, role int32, sessionID string) (string, int64, error) {
	now := time.Now()
	expiresAt := now.Add(m.accessTokenDuration)
	claims := Claims{
		UserID:    userID,
		Role:      role,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			Subject:   sessionID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secretKey)
	if err != nil {
		return "", 0, err
	}
	return signed, int64(m.accessTokenDuration.Seconds()), nil
}

func (m *TokenManager) ParseAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		return m.secretKey, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("token 无效")
	}
	return claims, nil
}

func (m *TokenManager) GenerateRefreshToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func (m *TokenManager) RefreshTokenTTL() time.Duration {
	return m.refreshTokenTTL
}

type RefreshSession struct {
	UserID    int64     `json:"user_id"`
	Role      int32     `json:"role"`
	SessionID string    `json:"session_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

func EncodeRefreshSession(session *RefreshSession) (string, error) {
	data, err := json.Marshal(session)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func DecodeRefreshSession(value string) (*RefreshSession, error) {
	var session RefreshSession
	if err := json.Unmarshal([]byte(value), &session); err != nil {
		return nil, err
	}
	return &session, nil
}
