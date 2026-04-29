package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/ngq/gorp/framework/goroutine"

	"github.com/spf13/cobra"
)

// safeGoCmd 演示 framework/goroutine 中 SafeGoAndWait 的基本用法。
//
// 中文说明：
// - 这是一个示例/调试命令，不是正式业务命令。
// - 它通过一个成功任务 + 一个失败任务，演示并发执行与错误汇总语义。
var safeGoCmd = &cobra.Command{
	Use:    "safego",
	Short:  "Demo SafeGo/SafeGoAndWait",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, c, err := bootstrap()
		if err != nil {
			return err
		}

		start := time.Now()
		err = goroutine.SafeGoAndWait(cmd.Context(), c,
			func(ctx context.Context) error {
				time.Sleep(50 * time.Millisecond)
				return nil
			},
			func(ctx context.Context) error {
				return fmt.Errorf("demo error")
			},
		)
		_ = start
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "error: %v\n", err)
			return nil
		}
		fmt.Fprintln(cmd.OutOrStdout(), "ok")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(safeGoCmd)
}
