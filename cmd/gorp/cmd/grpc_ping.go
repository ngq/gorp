package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	grpcPingAddr string
	grpcPingMsg  string
)

var grpcPingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Reserved command: use external examples to verify gRPC services",
	RunE: func(cmd *cobra.Command, args []string) error {
		_ = grpcPingAddr
		_ = grpcPingMsg
		return fmt.Errorf("the mother repo no longer includes a built-in gRPC ping demo; please use external examples to verify gRPC capabilities")
	},
}

func init() {
	grpcCmd.AddCommand(grpcPingCmd)
	grpcPingCmd.Flags().StringVar(&grpcPingAddr, "addr", ":9090", "gRPC server address")
	grpcPingCmd.Flags().StringVar(&grpcPingMsg, "msg", "hello", "message to send")
}
