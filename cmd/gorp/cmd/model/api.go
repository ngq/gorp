package model

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/ngq/gorp/framework/contract"

	"github.com/spf13/cobra"
)

var (
	modelAPITable      string
	modelAPIOut        string // http module output dir
	modelAPIServiceOut string
	modelAPIModelOut   string
	modelAPIBackend    string
	modelAPIForce      bool
)

// modelAPICmd 基于数据库表生成 model/service/http 三层 CRUD 骨架。
//
// 中文说明：
// - 这是比 `model gen` 更完整的一条脚手架链路。
// - 它会同时生成：
//   1. app/model/... 模型
//   2. app/service/... CRUD 服务
//   3. app/http/module/... HTTP 接口骨架
// - 生成后通常还需要手动把 routes 注册到 app/http/routes.go。
// - 当前阶段 `--backend` 已支持 `gorm|ent`：
//   - `gorm`：生成当前可直接使用的 GORM CRUD 骨架；
//   - `ent`：生成最小占位骨架，不含 gorm import / gorm tag / `*gorm.DB`，用于后续接入 ent repository/client。
var modelAPICmd = &cobra.Command{
	Use:   "api",
	Short: "Generate CRUD API skeleton from DB table",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, c, err := bootstrap()
		if err != nil {
			return err
		}
		insAny, err := c.Make(contract.DBInspectorKey)
		if err != nil {
			return err
		}
		ins := insAny.(contract.DBInspector)

		if modelAPITable == "" {
			return fmt.Errorf("--table is required")
		}
		backend := contract.NormalizeBackendName(modelAPIBackend)
		if backend != contract.RuntimeBackendGorm && backend != contract.RuntimeBackendEnt {
			return fmt.Errorf("unsupported --backend: %s", modelAPIBackend)
		}
		cols, err := ins.Columns(cmd.Context(), modelAPITable)
		if err != nil {
			return err
		}

		httpOutDir := modelAPIOut
		if httpOutDir == "" {
			httpOutDir = filepath.Join("app", "http", "module", "generated")
		}
		serviceOutDir := modelAPIServiceOut
		if serviceOutDir == "" {
			serviceOutDir = filepath.Join("app", "service", "generated")
		}
		modelOutDir := modelAPIModelOut
		if modelOutDir == "" {
			modelOutDir = filepath.Join("app", "model", "generated")
		}
		for _, d := range []string{httpOutDir, serviceOutDir, modelOutDir} {
			if err := os.MkdirAll(d, 0o755); err != nil {
				return err
			}
		}

		name := toGoName(modelAPITable)
		pkg := strings.ToLower(name)

		pkName := "id"
		for _, col := range cols {
			if col.PrimaryKey {
				pkName = col.Name
				break
			}
		}

		data := struct {
			Table     string
			Name      string
			Pkg       string
			Cols      []contract.Column
			PKName    string
			PKGoName  string
			PKGoType  string
			ModelPkg  string
			SvcPkg    string
			SvcName   string
			Force     bool
			Backend   string
			HTTPGroup string
			Module    string
		}{
			Table:     modelAPITable,
			Name:      name,
			Pkg:       pkg,
			Cols:      cols,
			PKName:    pkName,
			PKGoName:  toGoName(pkName),
			PKGoType:  goTypeFromCols(cols, pkName),
			ModelPkg:  "generated",
			SvcPkg:    "generated",
			SvcName:   name + "Service",
			Force:     modelAPIForce,
			Backend:   string(backend),
			HTTPGroup: "/" + pkg,
			Module:    detectModulePath(),
		}

		tpl := template.Must(template.New("root").Funcs(template.FuncMap{
			"toGoName": toGoName,
			"goType":   sqliteTypeToGo,
		}).Parse(""))

		modelTpl := apiModelTpl
		serviceTplSrc := serviceTpl
		httpFiles := map[string]string{
			"routes.go":     routesTpl,
			"api_create.go": createTpl,
			"api_update.go": updateTpl,
			"api_get.go":    getTpl,
			"api_list.go":   listTpl,
			"api_delete.go": deleteTpl,
		}
		if backend == contract.RuntimeBackendEnt {
			modelTpl = entAPIModelTpl
			serviceTplSrc = entServiceTpl
			httpFiles = map[string]string{
				"routes.go":     entRoutesTpl,
				"api_create.go": entCreateTpl,
				"api_update.go": entUpdateTpl,
				"api_get.go":    entGetTpl,
				"api_list.go":   entListTpl,
				"api_delete.go": entDeleteTpl,
			}
		}

		modelFile := filepath.Join(modelOutDir, pkg+".go")
		if err := writeTemplateFile(tpl, modelFile, modelTpl, data, modelAPIForce); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "generated: %s\n", modelFile)

		serviceFile := filepath.Join(serviceOutDir, pkg+".go")
		if err := writeTemplateFile(tpl, serviceFile, serviceTplSrc, data, modelAPIForce); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "generated: %s\n", serviceFile)

		targetDir := filepath.Join(httpOutDir, pkg)
		if err := os.MkdirAll(targetDir, 0o755); err != nil {
			return err
		}
		for fname, src := range httpFiles {
			fpath := filepath.Join(targetDir, fname)
			if err := writeTemplateFile(tpl, fpath, src, data, modelAPIForce); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "generated: %s\n", fpath)
		}

		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout(), "Next step: register the generated module routes in app/http/routes.go")
		fmt.Fprintf(cmd.OutOrStdout(), "Example: generated/%s.RegisterRoutes(engine, c)\n", pkg)
		return nil
	},
}

func init() {
	modelAPICmd.Flags().StringVar(&modelAPITable, "table", "", "table name")
	modelAPICmd.Flags().StringVar(&modelAPIOut, "out", "", "output directory for http module")
	modelAPICmd.Flags().StringVar(&modelAPIServiceOut, "service-out", "", "output directory for app service code")
	modelAPICmd.Flags().StringVar(&modelAPIModelOut, "model-out", "", "output directory for app model code")
	modelAPICmd.Flags().StringVar(&modelAPIBackend, "backend", string(contract.RuntimeBackendGorm), "generation backend: gorm|ent")
	modelAPICmd.Flags().BoolVar(&modelAPIForce, "force", false, "overwrite existing files")
}

func goTypeFromCols(cols []contract.Column, name string) string {
	for _, c := range cols {
		if c.Name == name {
			return sqliteTypeToGo(c.Type)
		}
	}
	return "int64"
}

func detectModulePath() string {
	content, err := os.ReadFile("go.mod")
	if err != nil {
		return "github.com/ngq/gorp"
	}
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return "github.com/ngq/gorp"
}

func writeTemplateFile(base *template.Template, filePath, src string, data any, force bool) error {
	if !force {
		if _, err := os.Stat(filePath); err == nil {
			return fmt.Errorf("file exists: %s (use --force to overwrite)", filePath)
		}
	}

	var f *os.File
	var err error
	if force {
		f, err = os.Create(filePath)
	} else {
		f, err = os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	}
	if err != nil {
		return err
	}
	defer f.Close()

	t, err := base.Clone()
	if err != nil {
		return err
	}
	t = template.Must(t.Parse(src))
	return t.ExecuteTemplate(f, "tpl", data)
}

const apiModelTpl = `{{define "tpl"}}package {{.ModelPkg}}

// Code generated by gorp model api (backend={{.Backend}}); DO NOT EDIT.

type {{.Name}} struct {
{{- range .Cols }}
	{{ toGoName .Name }} {{ goType .Type }} ` + "`" + `gorm:"column:{{.Name}}{{if .PrimaryKey}};primaryKey{{end}}" json:"{{.Name}}"` + "`" + `
{{- end }}
}

func ({{.Name}}) TableName() string { return "{{.Table}}" }
{{end}}`

const serviceTpl = `{{define "tpl"}}package {{.SvcPkg}}

// Code generated by gorp model api (backend={{.Backend}}); DO NOT EDIT.

import (
	"context"

	model "{{.Module}}/app/model/{{.ModelPkg}}"

	"gorm.io/gorm"
)

func Create{{.Name}}(ctx context.Context, db *gorm.DB, fields map[string]any) (*model.{{.Name}}, error) {
	_ = ctx
	item := &model.{{.Name}}{}
	if err := db.Table(item.TableName()).Create(fields).Error; err != nil {
		return nil, err
	}
	return item, nil
}

func Update{{.Name}}(ctx context.Context, db *gorm.DB, id any, fields map[string]any) error {
	_ = ctx
	item := &model.{{.Name}}{}
	return db.Table(item.TableName()).Where("{{.PKName}} = ?", id).Updates(fields).Error
}

func Get{{.Name}}(ctx context.Context, db *gorm.DB, id any) (*model.{{.Name}}, error) {
	_ = ctx
	item := &model.{{.Name}}{}
	if err := db.Table(item.TableName()).Where("{{.PKName}} = ?", id).First(item).Error; err != nil {
		return nil, err
	}
	return item, nil
}

func List{{.Name}}(ctx context.Context, db *gorm.DB, start, size int) ([]model.{{.Name}}, error) {
	_ = ctx
	if size <= 0 {
		size = 20
	}
	items := make([]model.{{.Name}}, 0, size)
	item := &model.{{.Name}}{}
	if err := db.Table(item.TableName()).Limit(size).Offset(start).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func Delete{{.Name}}(ctx context.Context, db *gorm.DB, id any) error {
	_ = ctx
	item := &model.{{.Name}}{}
	return db.Table(item.TableName()).Where("{{.PKName}} = ?", id).Delete(item).Error
}
{{end}}`

const routesTpl = `{{define "tpl"}}package {{.Pkg}}

import (
	"fmt"

	"github.com/ngq/gorp/framework/contract"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type API struct {
	container contract.Container
}

func RegisterRoutes(r *gin.Engine, container contract.Container) {
	api := &API{container: container}
	g := r.Group("{{.HTTPGroup}}")
	{
		g.POST("/create", api.Create)
		g.POST("/update", api.Update)
		g.GET("/get", api.Get)
		g.GET("/list", api.List)
		g.POST("/delete", api.Delete)
	}
}

func (api *API) mustRuntimeGorm() (*gorm.DB, error) {
	v, err := api.container.Make(contract.DBRuntimeKey)
	if err != nil {
		return nil, err
	}
	db, ok := v.(*gorm.DB)
	if !ok {
		return nil, fmt.Errorf("runtime backend is not *gorm.DB")
	}
	return db, nil
}
{{end}}`

const createTpl = `{{define "tpl"}}package {{.Pkg}}

import (
	"net/http"

	svc "{{.Module}}/app/service/{{.SvcPkg}}"

	"github.com/gin-gonic/gin"
)

func (api *API) Create(c *gin.Context) {
	fields := map[string]any{}
	if err := c.ShouldBindJSON(&fields); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db, err := api.mustRuntimeGorm()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	item, err := svc.Create{{.Name}}(c.Request.Context(), db, fields)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}
{{end}}`

const updateTpl = `{{define "tpl"}}package {{.Pkg}}

import (
	"net/http"

	svc "{{.Module}}/app/service/{{.SvcPkg}}"

	"github.com/gin-gonic/gin"
)

type updateParam struct {
	ID     any            ` + "`json:\"{{.PKName}}\" binding:\"required\"`" + `
	Fields map[string]any ` + "`json:\"fields\" binding:\"required\"`" + `
}

func (api *API) Update(c *gin.Context) {
	param := &updateParam{}
	if err := c.ShouldBindJSON(param); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db, err := api.mustRuntimeGorm()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := svc.Update{{.Name}}(c.Request.Context(), db, param.ID, param.Fields); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
{{end}}`

const getTpl = `{{define "tpl"}}package {{.Pkg}}

import (
	"net/http"

	svc "{{.Module}}/app/service/{{.SvcPkg}}"

	"github.com/gin-gonic/gin"
)

func (api *API) Get(c *gin.Context) {
	id := c.Query("{{.PKName}}")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing {{.PKName}}"})
		return
	}
	db, err := api.mustRuntimeGorm()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	item, err := svc.Get{{.Name}}(c.Request.Context(), db, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}
{{end}}`

const listTpl = `{{define "tpl"}}package {{.Pkg}}

import (
	"net/http"
	"strconv"

	svc "{{.Module}}/app/service/{{.SvcPkg}}"

	"github.com/gin-gonic/gin"
)

func (api *API) List(c *gin.Context) {
	start, _ := strconv.Atoi(c.Query("start"))
	size, _ := strconv.Atoi(c.Query("size"))
	db, err := api.mustRuntimeGorm()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	items, err := svc.List{{.Name}}(c.Request.Context(), db, start, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}
{{end}}`

const deleteTpl = `{{define "tpl"}}package {{.Pkg}}

import (
	"net/http"

	svc "{{.Module}}/app/service/{{.SvcPkg}}"

	"github.com/gin-gonic/gin"
)

type deleteParam struct {
	ID any ` + "`json:\"{{.PKName}}\" binding:\"required\"`" + `
}

func (api *API) Delete(c *gin.Context) {
	param := &deleteParam{}
	if err := c.ShouldBindJSON(param); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db, err := api.mustRuntimeGorm()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := svc.Delete{{.Name}}(c.Request.Context(), db, param.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
{{end}}`

const entAPIModelTpl = `{{define "tpl"}}package {{.ModelPkg}}

// Code generated by gorp model api (backend={{.Backend}}); DO NOT EDIT.
//
// 中文说明：
// - 这是 ent backend 的模型定义；
// - 不包含 gorm tag，配合 ent client 使用；
// - 实际 ent 项目应使用 ent generate 生成的类型，此为 API 响应 DTO。

type {{.Name}} struct {
{{- range .Cols }}
	{{ toGoName .Name }} {{ goType .Type }} ` + "`json:\"{{.Name}}\"`" + `
{{- end }}
}
{{end}}`

const entServiceTpl = `{{define "tpl"}}package {{.SvcPkg}}

// Code generated by gorp model api (backend={{.Backend}}); DO NOT EDIT.
//
// 中文说明：
// - ent backend 的 service 实现；
// - 基于 ent client 生成完整的 CRUD 操作；
// - 需要项目先执行 go generate ./ent 生成 ent 代码。

import (
	"context"

	model "{{.Module}}/app/model/{{.ModelPkg}}"
	"{{.Module}}/ent"
	"{{.Module}}/ent/{{.Pkg}}"
)

// MustRuntimeBackend 获取 ent client。
// 中文说明：此函数应由项目提供，返回 *ent.Client。
var MustRuntimeBackend func() *ent.Client

func Create{{.Name}}(ctx context.Context, client *ent.Client, fields map[string]any) (*model.{{.Name}}, error) {
	builder := client.{{.Name}}.Create()
	{{- range .Cols }}
	{{- if not .PrimaryKey }}
	if v, ok := fields["{{.Name}}"]; ok {
		builder = builder.Set{{ toGoName .Name }}(v.({{ goType .Type }}))
	}
	{{- end }}
	{{- end }}
	entity, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}
	return ent{{.Name}}ToModel(entity), nil
}

func Update{{.Name}}(ctx context.Context, client *ent.Client, id {{.PKGoType}}, fields map[string]any) error {
	builder := client.{{.Name}}.UpdateOneID(id)
	{{- range .Cols }}
	{{- if not .PrimaryKey }}
	if v, ok := fields["{{.Name}}"]; ok {
		builder = builder.Set{{ toGoName .Name }}(v.({{ goType .Type }}))
	}
	{{- end }}
	{{- end }}
	return builder.Exec(ctx)
}

func Get{{.Name}}(ctx context.Context, client *ent.Client, id {{.PKGoType}}) (*model.{{.Name}}, error) {
	entity, err := client.{{.Name}}.Query().
		Where({{.Pkg}}.{{.PKGoName}}EQ(id)).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return ent{{.Name}}ToModel(entity), nil
}

func List{{.Name}}(ctx context.Context, client *ent.Client, start, size int) ([]model.{{.Name}}, error) {
	if size <= 0 {
		size = 20
	}
	items, err := client.{{.Name}}.Query().
		Limit(size).
		Offset(start).
		All(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]model.{{.Name}}, len(items))
	for i, item := range items {
		result[i] = *ent{{.Name}}ToModel(item)
	}
	return result, nil
}

func Delete{{.Name}}(ctx context.Context, client *ent.Client, id {{.PKGoType}}) error {
	return client.{{.Name}}.DeleteOneID(id).Exec(ctx)
}

func ent{{.Name}}ToModel(e *ent.{{.Name}}) *model.{{.Name}} {
	return &model.{{.Name}}{
		{{- range .Cols }}
		{{ toGoName .Name }}: e.{{ toGoName .Name }},
		{{- end }}
	}
}
{{end}}`

const entRoutesTpl = `{{define "tpl"}}package {{.Pkg}}

import (
	"fmt"

	"github.com/ngq/gorp/framework/contract"

	"github.com/gin-gonic/gin"
	"{{.Module}}/ent"
)

type API struct {
	container contract.Container
}

func RegisterRoutes(r *gin.Engine, container contract.Container) {
	api := &API{container: container}
	g := r.Group("{{.HTTPGroup}}")
	{
		g.POST("/create", api.Create)
		g.POST("/update", api.Update)
		g.GET("/get", api.Get)
		g.GET("/list", api.List)
		g.POST("/delete", api.Delete)
	}
}

// mustEntClient 获取 ent client。
//
// 中文说明：
// - 从 container 或项目 runtime 获取 *ent.Client；
// - 假设项目已配置 ent backend 并执行 go generate ./ent。
func (api *API) mustEntClient() (*ent.Client, error) {
	// 方式1: 从 container 获取
	v, err := api.container.Make(contract.DBRuntimeKey)
	if err == nil {
		if client, ok := v.(*ent.Client); ok {
			return client, nil
		}
	}
	// 方式2: 从项目 runtime 获取
	// 假设项目中有 MustRuntimeBackend() 函数
	return nil, fmt.Errorf("ent client not available; ensure ent backend is configured")
}
{{end}}`

const entCreateTpl = `{{define "tpl"}}package {{.Pkg}}

import (
	"net/http"

	svc "{{.Module}}/app/service/{{.SvcPkg}}"

	"github.com/gin-gonic/gin"
)

func (api *API) Create(c *gin.Context) {
	fields := map[string]any{}
	if err := c.ShouldBindJSON(&fields); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	client, err := api.mustEntClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	item, err := svc.Create{{.Name}}(c.Request.Context(), client, fields)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}
{{end}}`

const entUpdateTpl = `{{define "tpl"}}package {{.Pkg}}

import (
	"net/http"

	svc "{{.Module}}/app/service/{{.SvcPkg}}"

	"github.com/gin-gonic/gin"
)

type updateParam struct {
	ID     any            ` + "`json:\"{{.PKName}}\" binding:\"required\"`" + `
	Fields map[string]any ` + "`json:\"fields\" binding:\"required\"`" + `
}

func (api *API) Update(c *gin.Context) {
	param := &updateParam{}
	if err := c.ShouldBindJSON(param); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	client, err := api.mustEntClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := svc.Update{{.Name}}(c.Request.Context(), client, param.ID, param.Fields); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
{{end}}`

const entGetTpl = `{{define "tpl"}}package {{.Pkg}}

import (
	"net/http"

	svc "{{.Module}}/app/service/{{.SvcPkg}}"

	"github.com/gin-gonic/gin"
)

func (api *API) Get(c *gin.Context) {
	id := c.Query("{{.PKName}}")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing {{.PKName}}"})
		return
	}
	client, err := api.mustEntClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	item, err := svc.Get{{.Name}}(c.Request.Context(), client, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}
{{end}}`

const entListTpl = `{{define "tpl"}}package {{.Pkg}}

import (
	"net/http"
	"strconv"

	svc "{{.Module}}/app/service/{{.SvcPkg}}"

	"github.com/gin-gonic/gin"
)

func (api *API) List(c *gin.Context) {
	start, _ := strconv.Atoi(c.Query("start"))
	size, _ := strconv.Atoi(c.Query("size"))
	client, err := api.mustEntClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	items, err := svc.List{{.Name}}(c.Request.Context(), client, start, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}
{{end}}`

const entDeleteTpl = `{{define "tpl"}}package {{.Pkg}}

import (
	"net/http"

	svc "{{.Module}}/app/service/{{.SvcPkg}}"

	"github.com/gin-gonic/gin"
)

type deleteParam struct {
	ID any ` + "`json:\"{{.PKName}}\" binding:\"required\"`" + `
}

func (api *API) Delete(c *gin.Context) {
	param := &deleteParam{}
	if err := c.ShouldBindJSON(param); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	client, err := api.mustEntClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := svc.Delete{{.Name}}(c.Request.Context(), client, param.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
{{end}}`