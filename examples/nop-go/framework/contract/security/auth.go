// Application scenarios:
// - Define the business-facing JWT authentication contract used by handlers and services.
// - Keep JWT signing, verification, and default-claims construction provider-neutral.
// - Separate user/business JWT semantics from service-to-service authentication semantics.
//
// 适用场景：
// - 定义 handler 和 service 直接使用的业务侧 JWT 认证契约。
// - 在 provider 中立的前提下统一 JWT 签发、校验和默认 claims 构建语义。
// - 将用户/业务 JWT 语义与服务间身份认证语义分离开来。
package security

const (
	// AuthJWTKey is the container key for the business JWT service.
	//
	// AuthJWTKey 是业务侧 JWT 服务的容器键。
	AuthJWTKey = "framework.auth.jwt"
)

// JWTClaims describes the shared business JWT claims model.
//
// JWTClaims 描述共享的业务 JWT claims 模型。
type JWTClaims struct {
	SubjectID   int64
	SubjectType string
	SubjectName string
	Roles       []string
	ExpiresAt   int64
	IssuedAt    int64
	Issuer      string
	Audience    string
}

// JWTService defines the minimal business JWT contract.
//
// JWTService 定义最小业务 JWT 契约。
//
// 中文说明：
// - 这是给业务项目直接使用的最小认证能力。
// - 不承载 session 存储，只负责 JWT 签发与校验。
// - 它与 service_auth.go 中的“服务间身份认证”是两条独立主线。
// - 可被 gin middleware、HTTP handler 和业务 service 复用。
type JWTService interface {
	// Sign signs one set of JWT claims.
	//
	// Sign 对一组 JWT claims 进行签发。
	Sign(claims JWTClaims) (string, error)

	// Verify verifies one JWT token and returns claims.
	//
	// Verify 校验一个 JWT token 并返回 claims。
	Verify(token string) (*JWTClaims, error)

	// NewClaims builds one standard JWT claims object.
	//
	// NewClaims 构建一份标准 JWT claims。
	NewClaims(subjectID int64, subjectType, subjectName string, roles []string, ttlSeconds int64) JWTClaims
}
