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
// - 这是对 `gorp new offline` 的联网补充方案。
// - 支持公开 GitHub Release + zip 模板包。
// - 模板包内部需要遵循和 offline 一致的目录约定：`templates/project/**`。
var (
	newReleaseRepo        string
	newReleaseTag         string
	newReleaseAsset       string
	newReleaseTemplate    string
	newReleasePreset      string
	newReleaseBackend     string
	newReleaseWithDB      bool
	newReleaseWithSwagger bool
	newReleaseWithAuth    bool
	newReleaseWithRBAC    bool
	newReleaseWithAdmin   bool
)

var newFromReleaseCmd = &cobra.Command{
	Use:   "from-release",
	Short: "Create a new project from GitHub Release",
	Long: `Create a new project by downloading starter templates from a GitHub Release asset.

Template options:
  - base          : minimal skeleton for custom structure
  - golayout      : standard single-service template
  - golayout-wire : single-service template with Wire assembly

Preset recommendation:
  - golayout-basic / golayout-enterprise: user-facing presets for the golayout template

Important:
  - When you specify --preset without --template, the template is automatically inferred.
  - For example: --preset golayout-enterprise automatically selects --template golayout.

If you are not sure which template to pick:
  - Use --template golayout or --template golayout-wire for from-release generation.
  - Multi-service Wire starter generation is currently recommended via offline --template multi-flat-wire.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !cmd.Flags().Changed("template") && newReleasePreset != "" {
			if inferred := templateFromPreset(newReleasePreset); inferred != "" {
				newReleaseTemplate = inferred
			}
		}
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

		presetProject := project
		if err := applyStarterPreset(newReleasePreset, &presetProject); err != nil {
			return err
		}
		if cmd.Flags().Changed("with-db") {
			project.WithDB = newReleaseWithDB
		} else {
			project.WithDB = presetProject.WithDB
		}
		if cmd.Flags().Changed("with-swagger") {
			project.WithSwagger = newReleaseWithSwagger
		} else {
			project.WithSwagger = presetProject.WithSwagger
		}
		if cmd.Flags().Changed("with-auth") {
			project.WithAuth = newReleaseWithAuth
		} else {
			project.WithAuth = presetProject.WithAuth
		}
		if cmd.Flags().Changed("with-rbac") {
			project.WithRBAC = newReleaseWithRBAC
		} else {
			project.WithRBAC = presetProject.WithRBAC
		}
		if cmd.Flags().Changed("with-admin") {
			project.WithAdmin = newReleaseWithAdmin
		} else {
			project.WithAdmin = presetProject.WithAdmin
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
		if _, err := fs.Stat(srcFS, "templates/project"); err != nil {
			return fmt.Errorf("release asset missing templates/project: %w", err)
		}

		data := buildScaffoldData(project)
		if err := renderTemplateProject(srcFS, "templates/project", folder, data); err != nil {
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
	newFromReleaseCmd.Flags().StringVar(&newReleaseTemplate, "template", starterTemplateBase, "starter template: base, golayout, golayout-wire")
	newFromReleaseCmd.Flags().StringVar(&newReleasePreset, "preset", "", "user preset: golayout-basic or golayout-enterprise")
	newFromReleaseCmd.Flags().StringVar(&newReleaseBackend, "backend", string(contract.RuntimeBackendGorm), "starter backend: gorm|ent")
	newFromReleaseCmd.Flags().BoolVar(&newReleaseWithDB, "with-db", true, "include DB sample and CRUD example")
	newFromReleaseCmd.Flags().BoolVar(&newReleaseWithSwagger, "with-swagger", true, "enable swagger config in generated starter")
	newFromReleaseCmd.Flags().BoolVar(&newReleaseWithAuth, "with-auth", true, "include auth/session starter")
	newFromReleaseCmd.Flags().BoolVar(&newReleaseWithRBAC, "with-rbac", true, "include Casbin RBAC starter")
	newFromReleaseCmd.Flags().BoolVar(&newReleaseWithAdmin, "with-admin", true, "include protected admin routes and RBAC management APIs")
}

func init() {
	newCmd.AddCommand(newFromReleaseCmd)
}
