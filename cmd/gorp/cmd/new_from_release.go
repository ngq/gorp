package cmd

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/ngq/gorp/framework/contract"
	"github.com/spf13/cobra"
)

// newFromReleaseCmd 从 GitHub Release 下载模板包并创建项目。
//
// 中文说明：
// - 这是对 `gorp new` 的联网补充方案。
// - 支持公开 GitHub Release + zip 模板包。
// - 模板包内部需要遵循和内置模板一致的目录约定：`templates/project/**`。
var (
	newReleaseRepo        string
	newReleaseTag         string
	newReleaseAsset       string
	newReleaseTemplate    string
	newReleaseBackend     string
	newReleaseWithDB      bool
	newReleaseWithSwagger bool
)

var newFromReleaseCmd = &cobra.Command{
	Use:   "from-release",
	Short: "Create a project from published release assets",
	Long: `Create a project from published GitHub Release starter assets.

This is a supported supplemental path.
Most users should start with 'gorp new', and only use from-release when they specifically need published release assets or fixed-version starter delivery.

Recommended path today:
  - Use bare 'gorp new' for the default single-service quick start.
  - Use positional intents on 'gorp new' for wire / multi / multi-wire.
  - Use from-release only when you specifically need published release assets.

Template options in the current release path:
  - base            : minimal skeleton for custom structure
  - golayout        : standard single-service template
  - golayout-wire   : advanced single-service template with Wire assembly
  - multi-flat      : standard multi-service template
  - multi-flat-wire : advanced multi-service template with Wire assembly

Important:
  - Use --template to choose the release template explicitly.
  - The release path now supports the public starter templates carried by release assets.

If you want the latest embedded matrix from the current CLI build:
  - Use 'gorp new' directly.
If you want fixed-version delivery:
  - Use 'gorp new from-release' with a pinned release tag.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateReleaseStarterTemplate(newReleaseTemplate); err != nil {
			return err
		}
		if err := validateGitHubRepo(newReleaseRepo); err != nil {
			return err
		}
		if strings.TrimSpace(newReleaseTag) == "" {
			return fmt.Errorf("tag is required")
		}
		if strings.TrimSpace(newReleaseAsset) == "" {
			newReleaseAsset = defaultReleaseTemplateAsset(newReleaseTemplate)
		}
		if err := validateAssetName(newReleaseAsset); err != nil {
			return err
		}

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		in := bufio.NewReader(cmd.InOrStdin())
		project, name, err := promptProjectInput(in, cmd.OutOrStdout(), "", false)
		if err != nil {
			return err
		}
		if newReleaseBackend == "" {
			newReleaseBackend = string(contract.RuntimeBackendGorm)
		}
		project.Backend = newReleaseBackend

		if cmd.Flags().Changed("with-db") {
			project.WithDB = newReleaseWithDB
		}
		if cmd.Flags().Changed("with-swagger") {
			project.WithSwagger = newReleaseWithSwagger
		}
		folder, err := prepareScaffoldTargetDir(cwd, name)
		if err != nil {
			return err
		}

		url := buildGitHubReleaseAssetURL(newReleaseRepo, newReleaseTag, newReleaseAsset)
		fmt.Fprintf(cmd.OutOrStdout(), "downloading template[%s]: %s\n", templateDisplayName(newReleaseTemplate), url)
		zipBytes, err := downloadReleaseAsset(url)
		if err != nil {
			return err
		}
		srcFS, err := newZipFSFromBytes(zipBytes)
		if err != nil {
			return err
		}
		templateRoot := resolveReleaseTemplateRoot(newReleaseTemplate)
		if _, err := fs.Stat(srcFS, templateRoot); err != nil {
			return fmt.Errorf("release asset missing %s: %w", templateRoot, err)
		}

		data := buildScaffoldData(project)
		if err := renderTemplateProject(srcFS, templateRoot, folder, data); err != nil {
			return err
		}

		printScaffoldNext(cmd.OutOrStdout(), folder)
		return nil
	},
}

func init() {
	newFromReleaseCmd.Flags().StringVar(&newReleaseRepo, "repo", "<owner>/<repo>", "GitHub repository (owner/repo)")
	newFromReleaseCmd.Flags().StringVar(&newReleaseTag, "tag", "latest", "Release tag (or 'latest')")
	newFromReleaseCmd.Flags().StringVar(&newReleaseAsset, "asset", "", "Release asset file name (default depends on --template)")
	newFromReleaseCmd.Flags().StringVar(&newReleaseTemplate, "template", starterTemplateGoLayout, "release starter template: base, golayout, golayout-wire, multi-flat, multi-flat-wire")
	newFromReleaseCmd.Flags().StringVar(&newReleaseBackend, "backend", string(contract.RuntimeBackendGorm), "starter backend: gorm|ent")
	newFromReleaseCmd.Flags().BoolVar(&newReleaseWithDB, "with-db", true, "include DB sample and CRUD example")
	newFromReleaseCmd.Flags().BoolVar(&newReleaseWithSwagger, "with-swagger", true, "enable swagger config in generated starter")
}

func init() {
	newCmd.AddCommand(newFromReleaseCmd)
}
