package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	securitycontract "github.com/ngq/gorp/framework/contract/security"
)

type JWTService struct {
	secret   string
	issuer   string
	audience string
}

func NewJWTService(secret, issuer, audience string) *JWTService {
	return &JWTService{secret: strings.TrimSpace(secret), issuer: issuer, audience: audience}
}

func (s *JWTService) Sign(claims securitycontract.JWTClaims) (string, error) {
	if s.secret == "" {
		return "", fmt.Errorf("jwt secret is required")
	}
	headerJSON, err := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		return "", err
	}
	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	headerPart := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadPart := base64.RawURLEncoding.EncodeToString(payloadJSON)
	signingInput := headerPart + "." + payloadPart
	h := hmac.New(sha256.New, []byte(s.secret))
	if _, err := h.Write([]byte(signingInput)); err != nil {
		return "", err
	}
	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	return signingInput + "." + signature, nil
}

func (s *JWTService) Verify(token string) (*securitycontract.JWTClaims, error) {
	if s.secret == "" {
		return nil, fmt.Errorf("jwt secret is required")
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token")
	}
	signingInput := parts[0] + "." + parts[1]
	h := hmac.New(sha256.New, []byte(s.secret))
	if _, err := h.Write([]byte(signingInput)); err != nil {
		return nil, fmt.Errorf("invalid token")
	}
	expectedSig := h.Sum(nil)
	gotSig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid token")
	}
	if !hmac.Equal(gotSig, expectedSig) {
		return nil, fmt.Errorf("invalid token")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid token")
	}
	var claims securitycontract.JWTClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("invalid token")
	}
	if claims.SubjectID == 0 || time.Now().Unix() >= claims.ExpiresAt {
		return nil, fmt.Errorf("invalid token")
	}
	return &claims, nil
}

func (s *JWTService) NewClaims(subjectID int64, subjectType, subjectName string, roles []string, ttlSeconds int64) securitycontract.JWTClaims {
	now := time.Now()
	if ttlSeconds <= 0 {
		ttlSeconds = 86400
	}
	return securitycontract.JWTClaims{
		SubjectID:   subjectID,
		SubjectType: subjectType,
		SubjectName: subjectName,
		Roles:       roles,
		IssuedAt:    now.Unix(),
		ExpiresAt:   now.Add(time.Duration(ttlSeconds) * time.Second).Unix(),
		Issuer:      s.issuer,
		Audience:    s.audience,
	}
}
