package security

const (
	// AuthJWTKey 是业务 JWT 服务在容器中的绑定 key。
	AuthJWTKey = "framework.auth.jwt"
)

// JWTClaims 表示业务 JWT 的通用 claims。
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

// JWTService 定义业务侧 JWT 最小契约。
//
// 中文说明：
// - 这是给业务项目直接使用的最小认证能力；
// - 不承担 session 存储，只负责 JWT 签发与校验；
// - 它与 service_auth.go 中的“服务间身份认证”是两条独立主线；
// - 可被 gin middleware、HTTP handler、业务 service 复用。
type JWTService interface {
	Sign(claims JWTClaims) (string, error)
	Verify(token string) (*JWTClaims, error)
	NewClaims(subjectID int64, subjectType, subjectName string, roles []string, ttlSeconds int64) JWTClaims
}
