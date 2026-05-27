package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

const jwtAlgorithm = "HS256"

var ErrInvalidToken = errors.New("invalid token")

type Config struct {
	JWTSecret        string
	JWTIssuer        string
	JWTAudience      []string
	TelegramBotToken string
	TokenTTL         time.Duration
}

type TokenManager struct {
	secret   []byte
	issuer   string
	audience []string
	ttl      time.Duration
	now      func() time.Time
}

type Claims struct {
	Issuer   string   `json:"iss"`
	Audience []string `json:"aud"`
	Subject  string   `json:"sub"`
	Role     string   `json:"role"`
	Provider string   `json:"provider"`
	Expiry   int64    `json:"exp"`
	IssuedAt int64    `json:"iat"`
}

func NewTokenManager(cfg Config) *TokenManager {
	cfg = cfg.Normalized()
	return &TokenManager{
		secret:   []byte(cfg.JWTSecret),
		issuer:   cfg.JWTIssuer,
		audience: cfg.JWTAudience,
		ttl:      cfg.TokenTTL,
		now:      time.Now,
	}
}

func (cfg Config) Normalized() Config {
	ttl := cfg.TokenTTL
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	if cfg.JWTIssuer == "" {
		cfg.JWTIssuer = "email-triage-system"
	}
	if len(cfg.JWTAudience) == 0 {
		cfg.JWTAudience = []string{"email-triage-system", "pg-ops-console"}
	}
	cfg.TokenTTL = ttl
	return cfg
}

func (m *TokenManager) Issue(principal Principal) (string, error) {
	if len(m.secret) == 0 {
		return "", errors.New("jwt secret is not configured")
	}
	now := m.now().UTC()
	claims := Claims{
		Issuer:   m.issuer,
		Audience: m.audience,
		Subject:  principal.UserID,
		Role:     principal.Role,
		Provider: principal.Provider,
		Expiry:   now.Add(m.ttl).Unix(),
		IssuedAt: now.Unix(),
	}
	header := map[string]string{"alg": jwtAlgorithm, "typ": "JWT"}

	encodedHeader, err := encodeJSONSegment(header)
	if err != nil {
		return "", err
	}
	encodedClaims, err := encodeJSONSegment(claims)
	if err != nil {
		return "", err
	}

	signingInput := encodedHeader + "." + encodedClaims
	signature := signHS256(m.secret, signingInput)
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func (m *TokenManager) Verify(token string, expectedAudience string) (Principal, error) {
	if len(m.secret) == 0 {
		return Principal{}, errors.New("jwt secret is not configured")
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Principal{}, ErrInvalidToken
	}

	signingInput := parts[0] + "." + parts[1]
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return Principal{}, ErrInvalidToken
	}
	if !hmac.Equal(signature, signHS256(m.secret, signingInput)) {
		return Principal{}, ErrInvalidToken
	}

	var header struct {
		Algorithm string `json:"alg"`
	}
	if err := decodeJSONSegment(parts[0], &header); err != nil || header.Algorithm != jwtAlgorithm {
		return Principal{}, ErrInvalidToken
	}

	var claims Claims
	if err := decodeJSONSegment(parts[1], &claims); err != nil {
		return Principal{}, ErrInvalidToken
	}
	if claims.Issuer != m.issuer || claims.Subject == "" || claims.Role == "" {
		return Principal{}, ErrInvalidToken
	}
	if claims.Expiry <= m.now().UTC().Unix() {
		return Principal{}, ErrInvalidToken
	}
	if expectedAudience != "" && !slices.Contains(claims.Audience, expectedAudience) {
		return Principal{}, ErrInvalidToken
	}
	if claims.Role != "user" && claims.Role != "admin" {
		return Principal{}, ErrInvalidToken
	}

	return Principal{
		UserID:   claims.Subject,
		Role:     claims.Role,
		Provider: claims.Provider,
	}, nil
}

func encodeJSONSegment(value any) (string, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func decodeJSONSegment(segment string, value any) error {
	b, err := base64.RawURLEncoding.DecodeString(segment)
	if err != nil {
		return fmt.Errorf("%w: malformed segment", ErrInvalidToken)
	}
	if err := json.Unmarshal(b, value); err != nil {
		return fmt.Errorf("%w: malformed json", ErrInvalidToken)
	}
	return nil
}

func signHS256(secret []byte, input string) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(input))
	return mac.Sum(nil)
}
