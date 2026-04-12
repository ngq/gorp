package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var cronListCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered cron jobs (no built-in demo jobs in mother repo)",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(cmd.OutOrStdout(), "(no built-in jobs; use external examples to demonstrate cron capabilities)")
		return nil
	},
}
