package container

import "github.com/ngq/gorp/framework/contract"

// MustMakeJWTService 返回当前容器中的业务 JWT 服务。
//
// 中文说明：
// - 用于业务项目快速接入 framework 级 auth/JWT 最小骨架；
// - 失败直接 panic，适合启动阶段或框架内部约束调用；
// - 业务层如需可恢复错误，优先使用 Make(contract.AuthJWTKey)。
func MustMakeJWTService(c contract.Container) contract.JWTService {
	v := c.MustMake(contract.AuthJWTKey)
	return v.(contract.JWTService)
}
