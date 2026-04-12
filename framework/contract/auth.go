package contract

const (
	// AuthJWTKey 是业务 JWT 服务在容器中的绑定 key。
	//
	// 中文说明：
	// - 面向业务项目的最小 JWT 骨架；
	// - 用于签发/校验最终用户或后台用户的 JWT；
	// - 与 ServiceAuthKey（服务间认证）区分开。
	AuthJWTKey = "framework.auth.jwt"
)

// JWTClaims 表示业务 JWT 的通用 claims。
//
// 中文说明：
// - SubjectID: 当前主体 ID；
// - SubjectType: 主体类型（如 user/admin/customer）；
// - SubjectName: 当前主体名（可选）；
// - Roles: 当前主体角色列表（可选）；
// - ExpiresAt / IssuedAt / Issuer / Audience: 基础 JWT 语义字段。
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

// JWTService 定义业务 JWT 最小骨架。
//
// 中文说明：
// - 这是给业务项目直接使用的最小认证能力；
// - 不承担 session 存储，只负责 JWT 签发与校验；
// - 可被 gin middleware、HTTP handler、业务 service 复用。
type JWTService interface {
	Sign(claims JWTClaims) (string, error)
	Verify(token string) (*JWTClaims, error)
	NewClaims(subjectID int64, subjectType, subjectName string, roles []string, ttlSeconds int64) JWTClaims
}
