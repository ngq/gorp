// Application scenarios:
// - Define the runtime cron scheduling contract used by bootstrap and providers.
// - Support both anonymous and named jobs with start/stop lifecycle control.
// - Keep cron execution context-aware for framework-managed background tasks.
//
// 适用场景：
// - 定义 bootstrap 和 provider 使用的运行时 cron 调度契约。
// - 支持匿名任务和具名任务，并提供启动/停止生命周期控制。
// - 让框架管理的后台任务具备 context 感知能力。
package runtime

import "context"

// CronKey is the container key for the cron capability.
//
// CronKey 是 cron 能力的容器键。
const CronKey = "framework.cron"

// Cron defines the runtime cron scheduling capability.
//
// Cron 定义运行时 cron 调度能力。
type Cron interface {
	// Add registers a cron job.
	//
	// Add 注册一个 cron 任务。
	Add(spec string, fn func(ctx context.Context) error) (entryID int, err error)

	// AddNamed registers a named cron job.
	//
	// AddNamed 注册一个具名 cron 任务。
	AddNamed(name, spec string, fn func(ctx context.Context) error) (entryID int, err error)

	// Start starts the cron scheduler.
	//
	// Start 启动 cron 调度器。
	Start()

	// Stop stops the cron scheduler and returns a context for shutdown coordination.
	//
	// Stop 停止 cron 调度器，并返回一个用于关闭协调的 context。
	Stop() context.Context
}
