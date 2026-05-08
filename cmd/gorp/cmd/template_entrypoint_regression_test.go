package cmd

import (
	"os"
	"path/filepath"
	"testing"

	frameworktesting "github.com/ngq/gorp/framework/testing"
	"github.com/stretchr/testify/require"
)

func TestGoLayoutTemplateUsesApplicationRunEntrypoint(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "golayout-entry")
	data := buildScaffoldData(scaffoldInput{
		Name:            "golayout-entry",
		Module:          "example.com/golayout-entry",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})

	require.NoError(t, renderTemplateProject(projectTemplateFS, resolveOfflineTemplateRoot(starterTemplateGoLayout), projectDir, data))

	mainFile := filepath.Join(projectDir, "cmd", "app", "main.go")
	content, err := os.ReadFile(mainFile)
	require.NoError(t, err)
	text := string(content)

	require.Contains(t, text, "gorp.Run(")
	require.Contains(t, text, "gorp.HTTP()")
	require.Contains(t, text, "gorp.WithMicroserviceMode()")
	require.Contains(t, text, "gorp.WithMigrate(migrate)")
	require.Contains(t, text, "gorp.WithSetup(setup)")
	require.Contains(t, text, "func migrate(rt *gorp.HTTPRuntime) error")
	require.Contains(t, text, "return rt.DB.AutoMigrate(&data.DemoPO{})")
	require.NotContains(t, text, "gorp.BootHTTPService(")
	require.NotContains(t, text, "frameworkbootstrap.")
	require.NotContains(t, text, "rt.Container")
	require.NotContains(t, text, "cmd.Execute()")
	require.Contains(t, text, `apphttp "example.com/golayout-entry/app/http"`)
}

func TestGoLayoutTemplateRoutesDoNotDuplicateHealthz(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "golayout-routes")
	data := buildScaffoldData(scaffoldInput{
		Name:            "golayout-routes",
		Module:          "example.com/golayout-routes",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})

	require.NoError(t, renderTemplateProject(projectTemplateFS, resolveOfflineTemplateRoot(starterTemplateGoLayout), projectDir, data))

	routesFile := filepath.Join(projectDir, "app", "http", "routes.go")
	content, err := os.ReadFile(routesFile)
	require.NoError(t, err)
	text := string(content)

	require.Contains(t, text, `api := r.Group("/api/v1")`)
	require.NotContains(t, text, `engine.GET("/healthz"`)
	require.NotContains(t, text, "framework/contract")
}

func TestGoLayoutTemplateDoesNotExposeKernelImports(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "golayout-boundary")
	data := buildScaffoldData(scaffoldInput{
		Name:            "golayout-boundary",
		Module:          "example.com/golayout-boundary",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})

	require.NoError(t, renderTemplateProject(projectTemplateFS, resolveOfflineTemplateRoot(starterTemplateGoLayout), projectDir, data))

	var files []string
	require.NoError(t, filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	}))

	for _, path := range files {
		content, err := os.ReadFile(path)
		require.NoError(t, err)
		text := string(content)
		require.NotContains(t, text, "framework/bootstrap", path)
		require.NotContains(t, text, "framework/container", path)
		require.NotContains(t, text, "framework/contract", path)
		require.NotContains(t, text, "framework/provider", path)
		require.NotContains(t, text, "BootHTTPService", path)
		require.NotContains(t, text, "AutoMigrateModels", path)
	}
}

func TestReleaseGoLayoutTemplateDoesNotExposeKernelImports(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "release-golayout-boundary")
	data := buildScaffoldData(scaffoldInput{
		Name:            "release-golayout-boundary",
		Module:          "example.com/release-golayout-boundary",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})

	srcFS, srcRoot := releaseTemplateSource(starterTemplateGoLayout)
	require.NoError(t, renderTemplateProject(srcFS, srcRoot, projectDir, data))

	var files []string
	require.NoError(t, filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	}))

	for _, path := range files {
		content, err := os.ReadFile(path)
		require.NoError(t, err)
		text := string(content)
		require.NotContains(t, text, "framework/bootstrap", path)
		require.NotContains(t, text, "framework/container", path)
		require.NotContains(t, text, "framework/contract", path)
		require.NotContains(t, text, "framework/provider", path)
		require.NotContains(t, text, "BootHTTPService", path)
		require.NotContains(t, text, "AutoMigrateModels", path)
	}
}

func TestMultiFlatWireTemplateUsesApplicationEntrypointAndKeepsWireInCmdLayer(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "multi-flat-wire-entry")
	data := buildScaffoldData(scaffoldInput{
		Name:            "multi-flat-wire-entry",
		Module:          "example.com/multi-flat-wire-entry",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})

	require.NoError(t, renderTemplateProject(projectTemplateFS, resolveOfflineTemplateRoot(starterTemplateMultiFlatWire), projectDir, data))

	mainFile := filepath.Join(projectDir, "services", "user", "cmd", "main.go")
	content, err := os.ReadFile(mainFile)
	require.NoError(t, err)
	text := string(content)
	require.Contains(t, text, "gorp.Run(")
	require.Contains(t, text, "gorp.HTTP()")
	require.Contains(t, text, "gorp.WithMicroserviceMode()")
	require.Contains(t, text, "gorp.WithMigrate(migrate)")
	require.Contains(t, text, "gorp.WithSetup(setup)")
	require.Contains(t, text, "func migrate(rt *gorp.HTTPRuntime) error")
	require.Contains(t, text, "return rt.DB.AutoMigrate(&userdata.UserPO{})")
	require.NotContains(t, text, "frameworkbootstrap.BootHTTPService(")
	require.NotContains(t, text, "frameworkbootstrap.AutoMigrateModels(")

	wireFile := filepath.Join(projectDir, "services", "user", "cmd", "wire.go")
	wireContent, err := os.ReadFile(wireFile)
	require.NoError(t, err)
	wireText := string(wireContent)
	require.Contains(t, wireText, "package main")
	require.Contains(t, wireText, "func wireUserServices(db *gorm.DB)")
}

func TestReleaseGoLayoutTemplateUsesApplicationRunEntrypoint(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "release-golayout-entry")
	data := buildScaffoldData(scaffoldInput{
		Name:            "release-golayout-entry",
		Module:          "example.com/release-golayout-entry",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})

	srcFS, srcRoot := releaseTemplateSource(starterTemplateGoLayout)
	require.NoError(t, renderTemplateProject(srcFS, srcRoot, projectDir, data))

	mainFile := filepath.Join(projectDir, "cmd", "app", "main.go")
	content, err := os.ReadFile(mainFile)
	require.NoError(t, err)
	text := string(content)

	require.Contains(t, text, "gorp.Run(")
	require.Contains(t, text, "gorp.HTTP()")
	require.Contains(t, text, "gorp.WithMicroserviceMode()")
	require.Contains(t, text, "gorp.WithMigrate(migrate)")
	require.Contains(t, text, "gorp.WithSetup(setup)")
	require.Contains(t, text, "func migrate(rt *gorp.HTTPRuntime) error")
	require.Contains(t, text, "return rt.DB.AutoMigrate(&data.DemoPO{})")
	require.NotContains(t, text, "frameworkbootstrap.BootHTTPService(")
	require.NotContains(t, text, "cmd.Execute()")
	require.Contains(t, text, `apphttp "example.com/release-golayout-entry/app/http"`)
}

func TestReleaseGoLayoutTemplateRoutesDoNotDuplicateHealthz(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "release-golayout-routes")
	data := buildScaffoldData(scaffoldInput{
		Name:            "release-golayout-routes",
		Module:          "example.com/release-golayout-routes",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})

	srcFS, srcRoot := releaseTemplateSource(starterTemplateGoLayout)
	require.NoError(t, renderTemplateProject(srcFS, srcRoot, projectDir, data))

	routesFile := filepath.Join(projectDir, "app", "http", "routes.go")
	content, err := os.ReadFile(routesFile)
	require.NoError(t, err)
	text := string(content)

	require.Contains(t, text, `api := r.Group("/api/v1")`)
	require.NotContains(t, text, `engine.GET("/healthz"`)
}

func TestReleaseProjectTemplateUsesApplicationRunEntrypoint(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "release-project-entry")
	data := buildScaffoldData(scaffoldInput{
		Name:            "release-project-entry",
		Module:          "example.com/release-project-entry",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})

	srcFS, srcRoot := releaseTemplateSource(starterTemplateBase)
	require.NoError(t, renderTemplateProject(srcFS, srcRoot, projectDir, data))

	mainFile := filepath.Join(projectDir, "cmd", "app", "main.go")
	content, err := os.ReadFile(mainFile)
	require.NoError(t, err)
	text := string(content)

	require.Contains(t, text, "gorp.Run(")
	require.Contains(t, text, "gorp.HTTP()")
	require.Contains(t, text, "gorp.WithMicroserviceMode()")
	require.Contains(t, text, "gorp.WithSetup(setup)")
	require.NotContains(t, text, "frameworkbootstrap.BootHTTPService(")
	require.NotContains(t, text, "frameworkbootstrap.")
	require.Contains(t, text, `apphttp "example.com/release-project-entry/app/http"`)

	readme, err := os.ReadFile(filepath.Join(projectDir, "README.md"))
	require.NoError(t, err)
	require.Contains(t, string(readme), "WithMicroserviceMode")
	require.Contains(t, string(readme), "governance.disable")
	require.Contains(t, string(readme), "governance.providers")

	configFile, err := os.ReadFile(filepath.Join(projectDir, "config", "app.yaml"))
	require.NoError(t, err)
	require.Contains(t, string(configFile), "governance:")
	require.Contains(t, string(configFile), "disable: []")
	require.Contains(t, string(configFile), "providers: {}")
}

func TestGoLayoutDockerfileRunsProjectBinaryDirectly(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "golayout-docker")
	data := buildScaffoldData(scaffoldInput{
		Name:            "golayout-docker",
		Module:          "example.com/golayout-docker",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})

	require.NoError(t, renderTemplateProject(projectTemplateFS, resolveOfflineTemplateRoot(starterTemplateGoLayout), projectDir, data))

	dockerfile, err := os.ReadFile(filepath.Join(projectDir, "Dockerfile"))
	require.NoError(t, err)
	text := string(dockerfile)
	require.Contains(t, text, `ENTRYPOINT ["./app"]`)
	require.NotContains(t, text, `"./app", "app", "start"`)
}

func TestGoLayoutDocsAndDeployAssetsFollowApplicationAndProjectScopedNaming(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "Demo_App")
	data := buildScaffoldData(scaffoldInput{
		Name:            "Demo_App",
		Module:          "example.com/demo-app",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})

	require.NoError(t, renderTemplateProject(projectTemplateFS, resolveOfflineTemplateRoot(starterTemplateGoLayout), projectDir, data))

	readme, err := os.ReadFile(filepath.Join(projectDir, "README.md"))
	require.NoError(t, err)
	readmeText := string(readme)
	require.Contains(t, readmeText, "gorp.Run")
	require.Contains(t, readmeText, "WithMicroserviceMode")
	require.Contains(t, readmeText, "governance.disable")
	require.Contains(t, readmeText, "governance.providers")
	require.Contains(t, readmeText, "docs/deploy.md")
	require.Contains(t, readmeText, "demo-app")
	require.NotContains(t, readmeText, "framework/bootstrap")

	configFile, err := os.ReadFile(filepath.Join(projectDir, "config", "app.yaml"))
	require.NoError(t, err)
	configText := string(configFile)
	require.Contains(t, configText, "governance:")
	require.Contains(t, configText, "disable: []")
	require.Contains(t, configText, "providers: {}")

	deployDoc, err := os.ReadFile(filepath.Join(projectDir, "docs", "deploy.md"))
	require.NoError(t, err)
	deployText := string(deployDoc)
	require.Contains(t, deployText, "docker build -t demo-app:latest .")
	require.Contains(t, deployText, "demo-app-config")
	require.Contains(t, deployText, "demo-app-secret")
	require.Contains(t, deployText, "demo-app-dev")
	require.Contains(t, deployText, "demo-app-staging")

	deployment, err := os.ReadFile(filepath.Join(projectDir, "deploy", "kubernetes", "base", "deployment.yaml"))
	require.NoError(t, err)
	deploymentText := string(deployment)
	require.Contains(t, deploymentText, "name: demo-app")
	require.Contains(t, deploymentText, "image: your-registry/demo-app:latest")
	require.NotContains(t, deploymentText, "gorp-service")

	serviceDoc, err := os.ReadFile(filepath.Join(projectDir, "deploy", "kubernetes", "base", "service.yaml"))
	require.NoError(t, err)
	require.Contains(t, string(serviceDoc), "name: demo-app")
	require.NotContains(t, string(serviceDoc), "gorp-service")
}

func TestReleaseGoLayoutDocsAndDeployAssetsFollowApplicationAndProjectScopedNaming(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "Demo_App")
	data := buildScaffoldData(scaffoldInput{
		Name:            "Demo_App",
		Module:          "example.com/demo-app",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})

	srcFS, srcRoot := releaseTemplateSource(starterTemplateGoLayout)
	require.NoError(t, renderTemplateProject(srcFS, srcRoot, projectDir, data))

	readme, err := os.ReadFile(filepath.Join(projectDir, "README.md"))
	require.NoError(t, err)
	readmeText := string(readme)
	require.Contains(t, readmeText, "gorp.Run")
	require.Contains(t, readmeText, "WithMicroserviceMode")
	require.Contains(t, readmeText, "governance.disable")
	require.Contains(t, readmeText, "governance.providers")
	require.Contains(t, readmeText, "docs/deploy.md")
	require.Contains(t, readmeText, "demo-app")
	require.NotContains(t, readmeText, "framework/bootstrap")

	configFile, err := os.ReadFile(filepath.Join(projectDir, "config", "app.yaml"))
	require.NoError(t, err)
	configText := string(configFile)
	require.Contains(t, configText, "governance:")
	require.Contains(t, configText, "disable: []")
	require.Contains(t, configText, "providers: {}")

	deployDoc, err := os.ReadFile(filepath.Join(projectDir, "docs", "deploy.md"))
	require.NoError(t, err)
	deployText := string(deployDoc)
	require.Contains(t, deployText, "docker build -t demo-app:latest .")
	require.Contains(t, deployText, "demo-app-config")
	require.Contains(t, deployText, "demo-app-secret")
	require.Contains(t, deployText, "demo-app-dev")
	require.Contains(t, deployText, "demo-app-staging")

	deployment, err := os.ReadFile(filepath.Join(projectDir, "deploy", "kubernetes", "base", "deployment.yaml"))
	require.NoError(t, err)
	deploymentText := string(deployment)
	require.Contains(t, deploymentText, "name: demo-app")
	require.Contains(t, deploymentText, "image: your-registry/demo-app:latest")
	require.NotContains(t, deploymentText, "gorp-service")
}
