package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/cobra"
)

// openapiCmd 是 OpenAPI3 相关工具命令组。
//
// 中文说明：
// - 当前主要承担"把 swagger2 转成 openapi3"这一补充步骤。
// - 它依赖 `swagger gen` 先产出 docs/swagger.json。
var openapiCmd = &cobra.Command{
	Use:   "openapi",
	Short: "OpenAPI conversion tools",
}

var openapiGenCmd = &cobra.Command{
	Use:   "gen",
	Short: "Convert swagger2 (docs/swagger.json) to openapi3 (docs/openapi.json)",
	RunE: func(cmd *cobra.Command, args []string) error {
		swaggerPath := filepath.Join("docs", "swagger.json")
		b, err := os.ReadFile(swaggerPath)
		if err != nil {
			return fmt.Errorf("read %s: %w", swaggerPath, err)
		}

		var doc2 openapi2.T
		if err := json.Unmarshal(b, &doc2); err != nil {
			return fmt.Errorf("parse swagger2 json: %w", err)
		}

		doc3, err := openapi2conv.ToV3(&doc2)
		if err != nil {
			return fmt.Errorf("convert to openapi3: %w", err)
		}
		// validate (some converted docs may omit fields and fail strict validation)
		// We keep the validation warning non-fatal to allow gradual adoption.
		if err := doc3.Validate(openapi3.NewLoader().Context); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: openapi3 validation failed: %v\n", err)
		}

		outPath := filepath.Join("docs", "openapi.json")
		out, err := json.MarshalIndent(doc3, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(outPath, out, 0o644); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "openapi3 generated at %s\n", outPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(openapiCmd)
	openapiCmd.AddCommand(openapiGenCmd)
}