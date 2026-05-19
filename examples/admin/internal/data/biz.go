package data

import (
	"admin/internal/biz"
)

// NewBiz 创建业务层（依赖注入）。
func NewBiz(data *Data) *biz.Biz {
	return biz.NewBiz(
		NewDemoRepo(data),
		NewUserRepo(data),
		NewRoleRepo(data),
	)
}
