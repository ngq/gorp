package runtime

import "context"

const CronKey = "framework.cron"

type Cron interface {
	Add(spec string, fn func(ctx context.Context) error) (entryID int, err error)
	AddNamed(name, spec string, fn func(ctx context.Context) error) (entryID int, err error)
	Start()
	Stop() context.Context
}
