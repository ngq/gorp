package cron

import (
	"context"

	frameworkcontainer "github.com/ngq/gorp/framework/container"
	"github.com/ngq/gorp/framework/contract"
)

// JobDef 定义了一个 cron 任务的最小描述。
//
// 中文说明：
// - Name：任务名称，主要用于日志/排错；
// - Spec：cron 表达式；
// - Handler：真正执行任务逻辑的函数。
//
// 采用”静态注册表”方案：
// - 项目 cron worker 启动时，会直接从 `JobDefs()` 取出所有任务并注册；
// - 后续可扩展为动态模块发现机制，支持插件化任务注册。
type JobDef struct {
	Name    string
	Spec    string
	Handler func(ctx context.Context, c contract.Container) error
}

// JobDefs 返回当前项目需要注册的所有 cron 任务。
//
// 中文说明：
// - 当前只有一个 `demo_heartbeat`，主要用于演示 cron worker 路径是否打通；
// - 后续真实业务任务可以继续按这个结构追加；
// - 这里的任务定义不直接启动，真正的启动动作发生在项目自己的 cron worker 启动链路中。
func JobDefs() []JobDef {
	return []JobDef{
		{
			Name: "demo_heartbeat",
			Spec: "*/5 * * * * *",
			Handler: func(ctx context.Context, c contract.Container) error {
				// 这个 demo 任务每 5 秒打一条日志，
				// 主要用于验证：cron worker 已启动、任务已注册、container 依赖可正常解析。
				l, err := frameworkcontainer.MakeLogger(c)
				if err != nil {
					return err
				}
				l.Info("cron heartbeat")
				return nil
			},
		},
	}
}
