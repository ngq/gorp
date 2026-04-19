package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ngq/gorp/framework/contract"
)

// JWTService 是 framework 级业务 JWT 最小实现。
//
// 中文说明：
// - 这里承接真正的业务 JWT 实现边界；
// - 面向最终用户、后台用户、客户等业务主体；
// - 与 serviceauth/token 的服务间认证实现分离。
// - 旧的 serviceauth/token 同名能力后续仅保留兼容壳。
type JWTService struct {
	secret   string
	issuer   string
	audience string
}

// NewJWTService 创建业务 JWT 服务。
//
// 中文说明：
// - secret 为签名密钥；
// - issuer / audience 为业务 JWT 的基础语义字段；
// - 返回值实现 contract.JWTService，可供 provider、业务 service、middleware 统一复用。
func NewJWTService(secret, issuer, audience string) *JWTService {
	return &JWTService{secret: strings.TrimSpace(secret), issuer: issuer, audience: audience}
}

// Sign 对业务 claims 执行签名。
func (s *JWTService) Sign(claims contract.JWTClaims) (string, error) {
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

// Verify 校验业务 JWT 并返回 claims。
func (s *JWTService) Verify(token string) (*contract.JWTClaims, error) {
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
	var claims contract.JWTClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("invalid token")
	}
	if claims.SubjectID == 0 || time.Now().Unix() >= claims.ExpiresAt {
		return nil, fmt.Errorf("invalid token")
	}
	return &claims, nil
}

// NewClaims 创建带默认 TTL 的业务 claims。
func (s *JWTService) NewClaims(subjectID int64, subjectType, subjectName string, roles []string, ttlSeconds int64) contract.JWTClaims {
	now := time.Now()
	if ttlSeconds <= 0 {
		ttlSeconds = 86400
	}
	return contract.JWTClaims{
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
