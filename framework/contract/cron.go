package contract

import "context"

const CronKey = "framework.cron"

type Cron interface {
	// Add 注册一个定时任务
	Add(spec string, fn func(ctx context.Context) error) (entryID int, err error)
	// AddNamed 注册一个带名称的定时任务（用于指标收集）
	AddNamed(name, spec string, fn func(ctx context.Context) error) (entryID int, err error)
	Start()
	Stop() context.Context
}
