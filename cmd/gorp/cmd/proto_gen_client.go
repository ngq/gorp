package cmd

import (
	"context"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	"github.com/spf13/cobra"
)

// protoGenClientCmd generates typed RPC client wrapper from proto file.
//
// protoGenClientCmd 从 proto 文件生成类型化 RPC 客户端 wrapper。
var protoGenClientCmd = &cobra.Command{
	Use:   "gen-client",
	Short: "Generate typed RPC client wrapper from proto",
	Long: `gorp proto gen-client generates a typed RPC client wrapper from a proto file.

The generated wrapper provides:
- Type-safe method calls matching proto service methods
- Integration with gorp's RPCClient (with governance middleware)
- Clean API without manual Call() with service/method strings

Example:
  gorp proto gen-client -f api/proto/user.proto -o internal/client/user_client.go

The generated client can be used like:
  client := NewUserClient(rpcClient)
  resp, err := client.GetUser(ctx, &pb.GetUserRequest{Id: 1})`,
	RunE: runProtoGenClient,
}

// genClientProtoFile is the proto file path.
var genClientProtoFile string

// genClientOutputFile is the output Go file path.
var genClientOutputFile string

// genClientPackageName is the Go package name.
var genClientPackageName string

// genClientServiceName is the specific service to generate.
var genClientServiceName string

// genClientPrefix is the client struct name prefix.
var genClientPrefix string

// genClientImportPaths are proto import paths.
var genClientImportPaths []string

// genClientGovernance indicates governance integration.
var genClientGovernance bool

func init() {
	protoCmd.AddCommand(protoGenClientCmd)

	protoGenClientCmd.Flags().StringVarP(&genClientProtoFile, "proto-file", "f", "", "Proto file path (required)")
	protoGenClientCmd.Flags().StringVarP(&genClientOutputFile, "output", "o", "", "Output Go file path (required)")
	protoGenClientCmd.Flags().StringVar(&genClientPackageName, "package", "", "Go package name (default: derived from output directory)")
	protoGenClientCmd.Flags().StringVar(&genClientServiceName, "service", "", "Service name to generate (default: all services)")
	protoGenClientCmd.Flags().StringVar(&genClientPrefix, "prefix", "", "Client struct name prefix (default: service name without 'Service' suffix)")
	protoGenClientCmd.Flags().StringSliceVarP(&genClientImportPaths, "import", "I", nil, "Proto import paths")
	protoGenClientCmd.Flags().BoolVar(&genClientGovernance, "governance", false, "Include governance integration comments")

	protoGenClientCmd.MarkFlagRequired("proto-file")
	protoGenClientCmd.MarkFlagRequired("output")
}

func runProtoGenClient(cmd *cobra.Command, args []string) error {
	opts := integrationcontract.ClientGenOptions{
		ProtoFile:     genClientProtoFile,
		OutputFile:    genClientOutputFile,
		PackageName:   genClientPackageName,
		ImportPaths:   genClientImportPaths,
		ServiceName:   genClientServiceName,
		ClientPrefix:  genClientPrefix,
		UseGovernance: genClientGovernance,
	}

	generator, err := createProtoGenerator(false)
	if err != nil {
		printProtoError("create generator", err)
		return err
	}

	if err := generator.GenClient(context.Background(), opts); err != nil {
		printProtoError("gen-client", err)
		return err
	}

	printProtoSuccess("gen-client", genClientOutputFile)
	return nil
}
