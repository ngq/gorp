package service

import "github.com/ngq/gorp/framework/contract"

// AutoMigrate 执行业务模型的自动迁移。
//
// 中文说明：
// - 当前母仓不再内置业务模型示例，因此默认不执行任何业务表迁移；
// - 真正的业务模型迁移由外部案例或生成项目负责；
// - 保留这个函数是为了不破坏当前 runtime provider 启动骨架。
func AutoMigrate(c contract.Container) error {
	_ = c
	return nil
}
