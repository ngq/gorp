// Application scenarios:
// - Define the runtime cron scheduling contract used by bootstrap and providers.
// - Support both anonymous and named jobs with start/stop lifecycle control.
// - Keep cron execution context-aware for framework-managed background tasks.
// - Expose job introspection for diagnostic endpoints (e.g. /debug/cron).
//
// 适用场景：
// - 定义 bootstrap 和 provider 使用的运行时 cron 调度契约。
// - 支持匿名任务和具名任务，并提供启动/停止生命周期控制。
// - 让框架管理的后台任务具备 context 感知能力。
// - 暴露任务内省能力，供诊断端点（如 /debug/cron）使用。
package runtime

import (
	"context"
	"time"
)

// CronKey is the container key for the cron capability.
//
// CronKey 是 cron 能力的容器键。
const CronKey = "framework.cron"

// CronJobStatus represents the execution status of a cron job.
//
// CronJobStatus 表示 cron 任务的执行状态。
type CronJobStatus string

const (
	// CronJobStatusSuccess indicates the last execution succeeded.
	//
	// CronJobStatusSuccess 表示上次执行成功。
	CronJobStatusSuccess CronJobStatus = "success"

	// CronJobStatusError indicates the last execution failed.
	//
	// CronJobStatusError 表示上次执行失败。
	CronJobStatusError CronJobStatus = "error"

	// CronJobStatusPending indicates the job has not yet executed.
	//
	// CronJobStatusPending 表示任务尚未执行过。
	CronJobStatusPending CronJobStatus = "pending"

	// CronJobStatusRunning indicates the job is currently executing.
	//
	// CronJobStatusRunning 表示任务正在执行中。
	CronJobStatusRunning CronJobStatus = "running"
)

// CronJobEntry describes a registered cron job and its recent execution state.
//
// CronJobEntry 描述一个已注册的 cron 任务及其最近执行状态。
type CronJobEntry struct {
	// ID is the unique entry identifier assigned by the scheduler.
	//
	// ID 是调度器分配的唯一任务标识。
	ID int `json:"id"`

	// Name is the human-readable job name. Empty for anonymous jobs.
	//
	// Name 是任务的可读名称。匿名任务为空。
	Name string `json:"name"`

	// Spec is the cron expression (e.g. "0 */5 * * * *").
	//
	// Spec 是 cron 表达式（如 "0 */5 * * * *"）。
	Spec string `json:"spec"`

	// Status is the last known execution status.
	//
	// Status 是最近一次执行的状态。
	Status CronJobStatus `json:"status"`

	// LastRunTime is the time of the last execution, zero if never run.
	//
	// LastRunTime 是上次执行时间，从未执行则为零值。
	LastRunTime time.Time `json:"last_run_time,omitempty"`

	// NextRunTime is the time of the next scheduled execution, zero if stopped.
	//
	// NextRunTime 是下次计划执行时间，已停止则为零值。
	NextRunTime time.Time `json:"next_run_time,omitempty"`

	// LastError is the error from the last execution, nil if succeeded or never run.
	//
	// LastError 是上次执行的错误，成功或从未执行则为 nil。
	LastError string `json:"last_error,omitempty"`

	// LastDuration is the duration of the last execution, zero if never run.
	//
	// LastDuration 是上次执行的耗时，从未执行则为零值。
	LastDuration time.Duration `json:"last_duration,omitempty"`
}

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

	// Jobs returns information about all registered cron jobs, including
	// schedule, last/next run time, and execution status.
	//
	// Jobs 返回所有已注册 cron 任务的信息，包括调度表达式、上次/下次执行时间及执行状态。
	Jobs() []CronJobEntry
}
