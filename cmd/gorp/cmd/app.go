package cmd

import "github.com/spf13/cobra"

var appCmd = &cobra.Command{
	Use:   "app",
	Short: "Application management",
}

func init() {
	appCmd.AddCommand(appStartCmd)
}
