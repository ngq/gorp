package cmd

import (
	"github.com/spf13/cobra"
)

var cronCmd = &cobra.Command{
	Use:   "cron",
	Short: "Cron worker management",
}

func init() {
	rootCmd.AddCommand(cronCmd)
}
