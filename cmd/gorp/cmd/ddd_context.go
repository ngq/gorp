package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

var (
	dddContextName   string // bounded context 名称，如 article, order
	dddContextModule string // go module 路径
	dddContextOut    string // 输出根目录
	dddContextForce  bool   // 是否覆盖已存在文件
)

// dddContextCmd 生成完整的 DDD bounded context。
//
// 中文说明：
// - 这是 DDD 生成器的主要入口；
// - 一次生成完整的四层结构：
//   1. domain/<context>/ - 实体和仓储接口
//   2. application/<context>/ - 用例（create/get/list/update/delete）
//   3. infrastructure/persistence/<context>_repository.go - 仓储实现
//   4. interfaces/http/handler/<context>.go - HTTP 处理器
// - 生成后需要在 routes.go 中手动注册路由。
var dddContextCmd = &cobra.Command{
	Use:   "context",
	Short: "Generate a DDD bounded context inside an existing DDD starter project",
	Long: `Generate a DDD bounded context inside an existing DDD starter project.

This command creates:
  - domain/<context>/entity.go - Domain entity with validation
  - domain/<context>/repository.go - Repository port (interface)
  - application/<context>/create.go - Create use case
  - application/<context>/get.go - Get by ID use case
  - application/<context>/list.go - List use case
  - application/<context>/update.go - Update use case
  - application/<context>/delete.go - Delete use case
  - infrastructure/persistence/<context>_repository.go - Repository implementation
  - interfaces/http/handler/<context>.go - HTTP handlers
  - interfaces/http/request/<context>.go - Request DTOs

Usage expectation:
  - First create a DDD starter project with 'gorp new offline --template ddd'
  - Then run 'gorp ddd context --name <context>' inside that project

After generation, register the routes in your routes.go:
  handler.Register<Context>Routes(api.Group("/<context>"))`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if dddContextName == "" {
			return fmt.Errorf("--name is required (e.g., article, order, task)")
		}
		if err := validateDDDStarterProject(dddContextOut); err != nil {
			return err
		}

		// 规范化 context 名称
		name := strings.ToLower(strings.TrimSpace(dddContextName))
		goName := toDDDGoName(name)

		// 确定输出目录
		outDir := dddContextOut
		if outDir == "" {
			outDir = "app"
		}

		// 确定 module 路径（用于 import）
		module := dddContextModule
		if module == "" {
			// 尝试从 go.mod 读取
			module = detectModule()
		}
		if module == "" {
			module = "your-module"
		}

		data := dddContextData{
			Name:      goName,      // Article
			NameLower: name,        // article
			Module:    module,      // github.com/xxx/yyy
			Force:     dddContextForce,
		}

		tpl := template.Must(template.New("root").Funcs(template.FuncMap{
			"toGoName": toDDDGoName,
		}).Parse(""))

		// 1. 生成 domain 层
		if err := generateDDDDomain(tpl, outDir, data); err != nil {
			return err
		}

		// 2. 生成 application 层
		if err := generateDDDApplication(tpl, outDir, data); err != nil {
			return err
		}

		// 3. 生成 infrastructure 层
		if err := generateDDDInfrastructure(tpl, outDir, data); err != nil {
			return err
		}

		// 4. 生成 interfaces 层
		if err := generateDDDInterfaces(tpl, outDir, data); err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "\nGenerated bounded context: %s\n", goName)
		fmt.Fprintln(cmd.OutOrStdout(), "\nNext steps:")
		fmt.Fprintf(cmd.OutOrStdout(), "  1. Review domain/%s/entity.go and add business rules\n", name)
		fmt.Fprintf(cmd.OutOrStdout(), "  2. Implement repository methods in infrastructure/persistence/%s_repository.go\n", name)
		fmt.Fprintf(cmd.OutOrStdout(), "  3. Register routes in interfaces/http/routes.go:\n")
		fmt.Fprintf(cmd.OutOrStdout(), "     handler.Register%sRoutes(api.Group(\"/%s\"))\n", goName, name)

		return nil
	},
}

func init() {
	dddCmd.AddCommand(dddContextCmd)
	dddContextCmd.Flags().StringVar(&dddContextName, "name", "", "bounded context name (e.g., article, order, task)")
	dddContextCmd.Flags().StringVar(&dddContextModule, "module", "", "go module path (auto-detected from go.mod if empty)")
	dddContextCmd.Flags().StringVar(&dddContextOut, "out", "", "output root directory (default: app)")
	dddContextCmd.Flags().BoolVar(&dddContextForce, "force", false, "overwrite existing files")
}

type dddContextData struct {
	Name      string // Go 风格名称：Article
	NameLower string // 小写名称：article
	Module    string // go module 路径
	Force     bool
}

// toDDDGoName 将下划线/短横线命名转为 Go 风格驼峰。
func toDDDGoName(s string) string {
	s = strings.ToLower(s)
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-'
	})
	var result string
	for _, p := range parts {
		if len(p) > 0 {
			result += strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return result
}

// detectModule 尝试从 go.mod 读取 module 路径。
func detectModule() string {
	content, err := os.ReadFile("go.mod")
	if err != nil {
		return ""
	}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return ""
}

// validateDDDStarterProject 检查当前目录（或指定 out 根）是否具备 DDD starter 的最小宿主结构。
//
// 中文说明：
// - `gorp ddd context` 不是“在任意空模块里直接起一个 DDD 项目”的命令；
// - 它的职责是在已有 DDD starter 项目中追加 bounded context；
// - 因此在生成前要先检查关键目录/文件是否存在，避免先生成一堆代码再编译失败。
func validateDDDStarterProject(outRoot string) error {
	root := outRoot
	if strings.TrimSpace(root) == "" {
		root = "app"
	}
	checks := []string{
		filepath.Join(root, "interfaces", "http", "response"),
		filepath.Join(root, "interfaces", "http", "routes.go"),
		filepath.Join(root, "domain"),
		filepath.Join(root, "application"),
		filepath.Join(root, "infrastructure"),
	}
	for _, p := range checks {
		if _, err := os.Stat(p); err != nil {
			return fmt.Errorf("ddd context must be run inside an existing DDD starter project; missing required path: %s", p)
		}
	}
	return nil
}

// writeDDDFile 统一封装 DDD 生成文件的落盘逻辑。
func writeDDDFile(base *template.Template, filePath, src string, data dddContextData) error {
	if !data.Force {
		if _, err := os.Stat(filePath); err == nil {
			return fmt.Errorf("file exists: %s (use --force to overwrite)", filePath)
		}
	}
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return err
	}
	t, err := base.Clone()
	if err != nil {
		return err
	}
	t = template.Must(t.Parse(src))
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	return t.ExecuteTemplate(f, "tpl", data)
}

// generateDDDDomain 生成 domain 层文件。
func generateDDDDomain(tpl *template.Template, outDir string, data dddContextData) error {
	domainDir := filepath.Join(outDir, "domain", data.NameLower)

	// entity.go
	entityFile := filepath.Join(domainDir, "entity.go")
	if err := writeDDDFile(tpl, entityFile, dddEntityTpl, data); err != nil {
		return err
	}
	fmt.Printf("generated: %s\n", entityFile)

	// repository.go
	repoFile := filepath.Join(domainDir, "repository.go")
	if err := writeDDDFile(tpl, repoFile, dddRepositoryTpl, data); err != nil {
		return err
	}
	fmt.Printf("generated: %s\n", repoFile)

	return nil
}

// generateDDDApplication 生成 application 层文件。
func generateDDDApplication(tpl *template.Template, outDir string, data dddContextData) error {
	appDir := filepath.Join(outDir, "application", data.NameLower)

	usecases := map[string]string{
		"create.go": dddCreateTpl,
		"get.go":    dddGetTpl,
		"list.go":   dddListTpl,
		"update.go": dddUpdateTpl,
		"delete.go": dddDeleteTpl,
	}

	for fname, src := range usecases {
		fpath := filepath.Join(appDir, fname)
		if err := writeDDDFile(tpl, fpath, src, data); err != nil {
			return err
		}
		fmt.Printf("generated: %s\n", fpath)
	}

	return nil
}

// generateDDDInfrastructure 生成 infrastructure 层文件。
func generateDDDInfrastructure(tpl *template.Template, outDir string, data dddContextData) error {
	infraDir := filepath.Join(outDir, "infrastructure", "persistence")

	repoFile := filepath.Join(infraDir, data.NameLower+"_repository.go")
	if err := writeDDDFile(tpl, repoFile, dddRepoImplTpl, data); err != nil {
		return err
	}
	fmt.Printf("generated: %s\n", repoFile)

	return nil
}

// generateDDDInterfaces 生成 interfaces 层文件。
func generateDDDInterfaces(tpl *template.Template, outDir string, data dddContextData) error {
	// handler
	handlerDir := filepath.Join(outDir, "interfaces", "http", "handler")
	handlerFile := filepath.Join(handlerDir, data.NameLower+".go")
	if err := writeDDDFile(tpl, handlerFile, dddHandlerTpl, data); err != nil {
		return err
	}
	fmt.Printf("generated: %s\n", handlerFile)

	// request
	reqDir := filepath.Join(outDir, "interfaces", "http", "request")
	reqFile := filepath.Join(reqDir, data.NameLower+".go")
	if err := writeDDDFile(tpl, reqFile, dddRequestTpl, data); err != nil {
		return err
	}
	fmt.Printf("generated: %s\n", reqFile)

	return nil
}

// ============ 模板定义 ============

const dddEntityTpl = `{{define "tpl"}}package {{.NameLower}}

// {{.Name}} 是 {{.NameLower}} 领域实体。
//
// 中文说明：
// - 这里放置领域核心业务逻辑和不变性约束；
// - 实体通过 ID 标识，而不是通过属性值；
// - 业务规则应放在实体方法中，而不是散落在 service 层。
type {{.Name}} struct {
	ID        uint   ` + "`json:\"id\"`" + `
	// TODO: 添加领域属性
}

// New{{.Name}} 创建新的 {{.Name}} 实体。
//
// 中文说明：
// - 工厂方法确保实体创建时的合法性；
// - 不要直接使用 &{{.Name}}{}，而是通过工厂方法创建。
func New{{.Name}}() *{{.Name}} {
	return &{{.Name}}{}
}

// Validate 验证实体的业务规则。
func (e *{{.Name}}) Validate() error {
	// TODO: 添加业务规则验证
	return nil
}
{{end}}`

const dddRepositoryTpl = `{{define "tpl"}}package {{.NameLower}}

import "context"

// {{.Name}}Repository 是 {{.Name}} 的仓储接口（端口）。
//
// 中文说明：
// - 仓储接口定义在 domain 层，属于领域的一部分；
// - 实现放在 infrastructure 层；
// - 这样保证 domain 层不依赖具体持久化技术。
type {{.Name}}Repository interface {
	// Create 创建新实体
	Create(ctx context.Context, entity *{{.Name}}) error
	// Get 根据 ID 获取实体
	Get(ctx context.Context, id uint) (*{{.Name}}, error)
	// List 获取实体列表
	List(ctx context.Context, offset, limit int) ([]*{{.Name}}, error)
	// Update 更新实体
	Update(ctx context.Context, entity *{{.Name}}) error
	// Delete 删除实体
	Delete(ctx context.Context, id uint) error
}
{{end}}`

const dddCreateTpl = `{{define "tpl"}}package {{.NameLower}}

import (
	"context"

	"{{.Module}}/app/domain/{{.NameLower}}"
)

// Create{{.Name}}UseCase 创建 {{.Name}} 的用例。
//
// 中文说明：
// - 用例编排业务流程，协调领域对象；
// - 用例不包含业务规则，业务规则在领域实体中；
// - 用例负责事务边界和持久化触发。
type Create{{.Name}}UseCase struct {
	repo {{.NameLower}}.{{.Name}}Repository
}

func NewCreate{{.Name}}UseCase(repo {{.NameLower}}.{{.Name}}Repository) *Create{{.Name}}UseCase {
	return &Create{{.Name}}UseCase{repo: repo}
}

type Create{{.Name}}Input struct {
	// TODO: 添加创建所需的输入字段
}

type Create{{.Name}}Output struct {
	Entity *{{.NameLower}}.{{.Name}}
}

func (uc *Create{{.Name}}UseCase) Execute(ctx context.Context, input Create{{.Name}}Input) (*Create{{.Name}}Output, error) {
	// 1. 创建实体
	entity := {{.NameLower}}.New{{.Name}}()

	// 2. 验证业务规则
	if err := entity.Validate(); err != nil {
		return nil, err
	}

	// 3. 持久化
	if err := uc.repo.Create(ctx, entity); err != nil {
		return nil, err
	}

	return &Create{{.Name}}Output{Entity: entity}, nil
}
{{end}}`

const dddGetTpl = `{{define "tpl"}}package {{.NameLower}}

import (
	"context"

	"{{.Module}}/app/domain/{{.NameLower}}"
)

// Get{{.Name}}UseCase 获取单个 {{.Name}} 的用例。
type Get{{.Name}}UseCase struct {
	repo {{.NameLower}}.{{.Name}}Repository
}

func NewGet{{.Name}}UseCase(repo {{.NameLower}}.{{.Name}}Repository) *Get{{.Name}}UseCase {
	return &Get{{.Name}}UseCase{repo: repo}
}

type Get{{.Name}}Input struct {
	ID uint
}

type Get{{.Name}}Output struct {
	Entity *{{.NameLower}}.{{.Name}}
}

func (uc *Get{{.Name}}UseCase) Execute(ctx context.Context, input Get{{.Name}}Input) (*Get{{.Name}}Output, error) {
	entity, err := uc.repo.Get(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	return &Get{{.Name}}Output{Entity: entity}, nil
}
{{end}}`

const dddListTpl = `{{define "tpl"}}package {{.NameLower}}

import (
	"context"

	"{{.Module}}/app/domain/{{.NameLower}}"
)

// List{{.Name}}UseCase 获取 {{.Name}} 列表的用例。
type List{{.Name}}UseCase struct {
	repo {{.NameLower}}.{{.Name}}Repository
}

func NewList{{.Name}}UseCase(repo {{.NameLower}}.{{.Name}}Repository) *List{{.Name}}UseCase {
	return &List{{.Name}}UseCase{repo: repo}
}

type List{{.Name}}Input struct {
	Offset int
	Limit  int
}

type List{{.Name}}Output struct {
	Entities []*{{.NameLower}}.{{.Name}}
}

func (uc *List{{.Name}}UseCase) Execute(ctx context.Context, input List{{.Name}}Input) (*List{{.Name}}Output, error) {
	if input.Limit <= 0 {
		input.Limit = 20
	}
	entities, err := uc.repo.List(ctx, input.Offset, input.Limit)
	if err != nil {
		return nil, err
	}
	return &List{{.Name}}Output{Entities: entities}, nil
}
{{end}}`

const dddUpdateTpl = `{{define "tpl"}}package {{.NameLower}}

import (
	"context"

	"{{.Module}}/app/domain/{{.NameLower}}"
)

// Update{{.Name}}UseCase 更新 {{.Name}} 的用例。
type Update{{.Name}}UseCase struct {
	repo {{.NameLower}}.{{.Name}}Repository
}

func NewUpdate{{.Name}}UseCase(repo {{.NameLower}}.{{.Name}}Repository) *Update{{.Name}}UseCase {
	return &Update{{.Name}}UseCase{repo: repo}
}

type Update{{.Name}}Input struct {
	ID uint
	// TODO: 添加更新字段
}

type Update{{.Name}}Output struct {
	Entity *{{.NameLower}}.{{.Name}}
}

func (uc *Update{{.Name}}UseCase) Execute(ctx context.Context, input Update{{.Name}}Input) (*Update{{.Name}}Output, error) {
	// 1. 获取现有实体
	entity, err := uc.repo.Get(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	// 2. 更新属性（通过实体方法）
	// entity.UpdateXXX(...)

	// 3. 验证
	if err := entity.Validate(); err != nil {
		return nil, err
	}

	// 4. 持久化
	if err := uc.repo.Update(ctx, entity); err != nil {
		return nil, err
	}

	return &Update{{.Name}}Output{Entity: entity}, nil
}
{{end}}`

const dddDeleteTpl = `{{define "tpl"}}package {{.NameLower}}

import (
	"context"

	"{{.Module}}/app/domain/{{.NameLower}}"
)

// Delete{{.Name}}UseCase 删除 {{.Name}} 的用例。
type Delete{{.Name}}UseCase struct {
	repo {{.NameLower}}.{{.Name}}Repository
}

func NewDelete{{.Name}}UseCase(repo {{.NameLower}}.{{.Name}}Repository) *Delete{{.Name}}UseCase {
	return &Delete{{.Name}}UseCase{repo: repo}
}

type Delete{{.Name}}Input struct {
	ID uint
}

func (uc *Delete{{.Name}}UseCase) Execute(ctx context.Context, input Delete{{.Name}}Input) error {
	return uc.repo.Delete(ctx, input.ID)
}
{{end}}`

const dddRepoImplTpl = `{{define "tpl"}}package persistence

import (
	"context"

	"{{.Module}}/app/domain/{{.NameLower}}"

	"gorm.io/gorm"
)

// {{.Name}}RepositoryImpl 是 {{.Name}}Repository 的 GORM 实现。
//
// 中文说明：
// - 仓储实现在 infrastructure 层；
// - 依赖具体持久化技术（这里是 GORM）；
// - 可以根据需要替换为其他实现（如 MongoDB、Redis 等）。
type {{.Name}}RepositoryImpl struct {
	db *gorm.DB
}

func New{{.Name}}Repository(db *gorm.DB) {{.NameLower}}.{{.Name}}Repository {
	return &{{.Name}}RepositoryImpl{db: db}
}

func (r *{{.Name}}RepositoryImpl) Create(ctx context.Context, entity *{{.NameLower}}.{{.Name}}) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

func (r *{{.Name}}RepositoryImpl) Get(ctx context.Context, id uint) (*{{.NameLower}}.{{.Name}}, error) {
	var entity {{.NameLower}}.{{.Name}}
	if err := r.db.WithContext(ctx).First(&entity, id).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *{{.Name}}RepositoryImpl) List(ctx context.Context, offset, limit int) ([]*{{.NameLower}}.{{.Name}}, error) {
	var entities []*{{.NameLower}}.{{.Name}}
	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *{{.Name}}RepositoryImpl) Update(ctx context.Context, entity *{{.NameLower}}.{{.Name}}) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

func (r *{{.Name}}RepositoryImpl) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&{{.NameLower}}.{{.Name}}{}, id).Error
}
{{end}}`

const dddHandlerTpl = `{{define "tpl"}}package handler

import (
	"net/http"
	"strconv"

	"{{.Module}}/app/application/{{.NameLower}}"
	"{{.Module}}/app/interfaces/http/response"

	"github.com/gin-gonic/gin"
)

// {{.Name}}Handler 处理 {{.Name}} 相关的 HTTP 请求。
type {{.Name}}Handler struct {
	createUseCase *{{.NameLower}}.Create{{.Name}}UseCase
	getUseCase    *{{.NameLower}}.Get{{.Name}}UseCase
	listUseCase   *{{.NameLower}}.List{{.Name}}UseCase
	updateUseCase *{{.NameLower}}.Update{{.Name}}UseCase
	deleteUseCase *{{.NameLower}}.Delete{{.Name}}UseCase
}

func New{{.Name}}Handler(
	create *{{.NameLower}}.Create{{.Name}}UseCase,
	get *{{.NameLower}}.Get{{.Name}}UseCase,
	list *{{.NameLower}}.List{{.Name}}UseCase,
	update *{{.NameLower}}.Update{{.Name}}UseCase,
	delete *{{.NameLower}}.Delete{{.Name}}UseCase,
) *{{.Name}}Handler {
	return &{{.Name}}Handler{
		createUseCase: create,
		getUseCase:    get,
		listUseCase:   list,
		updateUseCase: update,
		deleteUseCase: delete,
	}
}

func Register{{.Name}}Routes(g *gin.RouterGroup, h *{{.Name}}Handler) {
	g.POST("", h.Create)
	g.GET("/:id", h.Get)
	g.GET("", h.List)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
}

func (h *{{.Name}}Handler) Create(c *gin.Context) {
	// TODO: 绑定请求体
	input := {{.NameLower}}.Create{{.Name}}Input{}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.Failure("BAD_REQUEST", err.Error()))
		return
	}

	output, err := h.createUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Failure("INTERNAL_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(output.Entity))
}

func (h *{{.Name}}Handler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Failure("BAD_REQUEST", "invalid id"))
		return
	}

	output, err := h.getUseCase.Execute(c.Request.Context(), {{.NameLower}}.Get{{.Name}}Input{ID: uint(id)})
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Failure("INTERNAL_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(output.Entity))
}

func (h *{{.Name}}Handler) List(c *gin.Context) {
	offset, _ := strconv.Atoi(c.Query("offset"))
	limit, _ := strconv.Atoi(c.Query("limit"))

	output, err := h.listUseCase.Execute(c.Request.Context(), {{.NameLower}}.List{{.Name}}Input{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Failure("INTERNAL_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(output.Entities))
}

func (h *{{.Name}}Handler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Failure("BAD_REQUEST", "invalid id"))
		return
	}

	input := {{.NameLower}}.Update{{.Name}}Input{ID: uint(id)}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.Failure("BAD_REQUEST", err.Error()))
		return
	}

	output, err := h.updateUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Failure("INTERNAL_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(output.Entity))
}

func (h *{{.Name}}Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Failure("BAD_REQUEST", "invalid id"))
		return
	}

	if err := h.deleteUseCase.Execute(c.Request.Context(), {{.NameLower}}.Delete{{.Name}}Input{ID: uint(id)}); err != nil {
		c.JSON(http.StatusInternalServerError, response.Failure("INTERNAL_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"deleted": id}))
}
{{end}}`

const dddRequestTpl = `{{define "tpl"}}package request

// {{.Name}}CreateRequest 创建 {{.Name}} 的请求 DTO。
type {{.Name}}CreateRequest struct {
	// TODO: 添加创建请求字段
}

// {{.Name}}UpdateRequest 更新 {{.Name}} 的请求 DTO。
type {{.Name}}UpdateRequest struct {
	// TODO: 添加更新请求字段
}
{{end}}`