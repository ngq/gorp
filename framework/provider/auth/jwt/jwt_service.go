// Package jwt provides JWT service implementation using HS256 algorithm.
// The implementation is a lightweight JWT service without external dependencies.
// Note: This implementation only supports HS256 algorithm.
//
// 本文件提供 JWT 服务实现，使用 HS256 算法。
// 此实现是轻量级 JWT 服务，无需外部依赖。
// 注意：此实现仅支持 HS256 算法。
package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	securitycontract "github.com/ngq/gorp/framework/contract/security"
)

// JWTService implements securitycontract.JWTService using HS256 algorithm.
//
// JWTService 使用 HS256 算法实现 securitycontract.JWTService 接口。
type JWTService struct {
	secret string // secret is the signing secret.
	//
	// secret 签名密钥。
	issuer string // issuer is the JWT issuer.
	//
	// issuer JWT 发行者。
	audience string // audience is the JWT audience.
	//
	// audience JWT 受众。
}

// NewJWTService creates a new JWT service instance.
//
// NewJWTService 创建新的 JWT 服务实例。
func NewJWTService(secret, issuer, audience string) *JWTService {
	return &JWTService{secret: strings.TrimSpace(secret), issuer: issuer, audience: audience}
}

// Sign generates a JWT token from the given claims.
// Core logic: Create header and payload JSON, encode with base64, sign with HMAC-SHA256.
// Eg:
//
// Sign 根据给定的声明生成 JWT token。
// 核心逻辑：创建 header 和 payload JSON，base64 编码，用 HMAC-SHA256 签名。
// Eg:
//
//	claims := jwtSvc.NewClaims(123, "user", "John", []string{"admin"}, 3600)
//	token, err := jwtSvc.Sign(claims)
func (s *JWTService) Sign(claims securitycontract.JWTClaims) (string, error) {
	if s.secret == "" {
		return "", errors.New("jwt secret is required")
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

// Verify validates a JWT token and returns the claims.
// Core logic: Split token into parts, verify signature, decode payload, check expiration.
// Returns error if token is invalid, signature mismatched, or expired.
//
// Verify 验证 JWT token 并返回声明。
// 核心逻辑：分割 token 为三部分，验证签名，解码 payload，检查过期时间。
// token 无效、签名不匹配或过期时返回错误。
func (s *JWTService) Verify(token string) (*securitycontract.JWTClaims, error) {
	if s.secret == "" {
		return nil, errors.New("jwt secret is required")
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token")
	}
	signingInput := parts[0] + "." + parts[1]
	h := hmac.New(sha256.New, []byte(s.secret))
	if _, err := h.Write([]byte(signingInput)); err != nil {
		return nil, errors.New("invalid token")
	}
	expectedSig := h.Sum(nil)
	gotSig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, errors.New("invalid token")
	}
	if !hmac.Equal(gotSig, expectedSig) {
		return nil, errors.New("invalid token")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errors.New("invalid token")
	}
	var claims securitycontract.JWTClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, errors.New("invalid token")
	}
	if claims.SubjectID == 0 || time.Now().Unix() >= claims.ExpiresAt {
		return nil, errors.New("invalid token")
	}
	return &claims, nil
}

// NewClaims creates a new JWTClaims with default TTL (24 hours if not specified).
// Parameters: subjectID, subjectType, subjectName, roles, ttlSeconds.
//
// NewClaims 创建新的 JWTClaims，默认 TTL 为 24 小时（未指定时）。
// 参数：主体 ID、主体类型、主体名称、角色列表、有效期秒数。
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
