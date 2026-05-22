# nop-go 插件系统设计

> **核心结论**：插件机制属于产品层设计，框架层的 ServiceProvider + Container 已经足够。

## 一、设计原则

### 1.1 分层定位

```
┌─────────────────────────────────────────────────────────────────┐
│  产品层（nop-go）                                                │
│  - plugin.json 发现 + 加载                                       │
│  - IPaymentMethod / IShippingPlugin 等业务接口                   │
│  - 插件安装/卸载/迁移                                             │
│  - 基于 framework ServiceProvider 实现                           │
└─────────────────────────────────────────────────────────────────┘
                              │ 使用
┌─────────────────────────────────────────────────────────────────┐
│  框架层（gorp framework）                                        │
│  - ServiceProvider 接口（已有）                                  │
│  - Container（已有）                                             │
│  - 不增加额外插件引擎                                             │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2 技术选型

| 方案 | 阶段 | 说明 |
|------|------|------|
| 编译进主程序 | Phase 1 ✅ | 简单稳定、跨平台，默认方案 |
| Go plugin (.so) | Phase 2 | 可选支持，仅 Linux |
| 进程外 RPC | Phase 3 | 高级场景，完全隔离 |

## 二、插件接口设计

### 2.1 基础插件接口

```go
// internal/plugin/contract.go

package plugin

import (
    "context"
    "github.com/ngq/gorp/framework/contract"
)

// Plugin 产品级插件接口
//
// 中文说明：
// - 所有业务插件都必须实现这个接口；
// - 插件通过 ToServiceProvider() 转换后注册到 gorp Container；
// - 这样可以复用框架已有的 ServiceProvider 生命周期管理。
type Plugin interface {
    // Meta 返回插件元数据（从 plugin.json 加载）
    Meta() *PluginMeta

    // PluginType 返回插件类型
    // 例如: "payment", "shipping", "widget", "discount", "tax", "auth"
    PluginType() string

    // Install 插件安装时执行
    // 用于创建数据库表、初始化配置、写入默认数据
    Install(ctx context.Context, c contract.Container) error

    // Uninstall 插件卸载时执行
    // 用于清理数据、删除表（谨慎操作）
    Uninstall(ctx context.Context, c contract.Container) error

    // Boot 插件启动时执行
    // 用于初始化运行时状态、启动 goroutine 等
    Boot(ctx context.Context, c contract.Container) error

    // ToServiceProvider 转换为 gorp ServiceProvider
    // 这是连接产品插件和框架容器的桥梁
    ToServiceProvider() contract.ServiceProvider
}

// PluginMeta 插件元数据
type PluginMeta struct {
    // Group 插件分组
    // 例如: "Payment", "Shipping", "Misc", "Widgets"
    Group string `json:"group"`

    // FriendlyName 友好名称，显示给用户
    FriendlyName string `json:"friendly_name"`

    // SystemName 系统名称，唯一标识
    // 例如: "Payment.Alipay", "Shipping.FedEx"
    SystemName string `json:"system_name"`

    // Version 插件版本
    Version string `json:"version"`

    // SupportedVersions 支持的 nop-go 版本
    SupportedVersions []string `json:"supported_versions"`

    // Author 作者
    Author string `json:"author"`

    // DisplayOrder 显示顺序
    DisplayOrder int `json:"display_order"`

    // Description 插件描述
    Description string `json:"description"`

    // DependsOn 依赖的其他插件
    DependsOn []string `json:"depends_on"`

    // FileName 编译后的文件名（用于动态加载）
    FileName string `json:"file_name"`

    // Installed 是否已安装
    Installed bool `json:"installed"`

    // InstalledVersion 已安装的版本
    InstalledVersion string `json:"installed_version"`
}
```

### 2.2 业务插件接口（按类型扩展）

```go
// internal/plugin/contract_payment.go

package plugin

// PaymentMethod 支付方式插件接口
//
// 中文说明：
// - 所有支付插件（支付宝、微信、PayPal 等）都实现此接口；
// - 继承基础 Plugin 接口，增加支付特有能力。
type PaymentMethod interface {
    Plugin

    // ProcessPayment 处理支付
    ProcessPayment(ctx context.Context, req *ProcessPaymentRequest) (*ProcessPaymentResult, error)

    // Refund 退款
    Refund(ctx context.Context, req *RefundRequest) (*RefundResult, error)

    // Capture 捕获预授权
    Capture(ctx context.Context, req *CaptureRequest) (*CaptureResult, error)

    // Void 取消预授权
    Void(ctx context.Context, req *VoidRequest) (*VoidResult, error)

    // GetConfiguration 获取支付配置项
    GetConfiguration() []PaymentConfigItem

    // ValidateConfiguration 验证配置是否正确
    ValidateConfiguration(config map[string]string) error
}

// ProcessPaymentRequest 支付请求
type ProcessPaymentRequest struct {
    OrderID       uint64
    Amount        float64
    Currency      string
    CustomerID    uint64
    ReturnURL     string
    NotifyURL     string
    CustomFields  map[string]string
}

// ProcessPaymentResult 支付结果
type ProcessPaymentResult struct {
    Success       bool
    TransactionID string
    RedirectURL   string   // 需要跳转的支付页面
    ErrorMessage  string
}
```

```go
// internal/plugin/contract_shipping.go

package plugin

// ShippingMethod 配送方式插件接口
type ShippingMethod interface {
    Plugin

    // CalculateShippingRate 计算运费
    CalculateShippingRate(ctx context.Context, req *ShippingRateRequest) (*ShippingRateResult, error)

    // GetTrackingInfo 获取物流追踪信息
    GetTrackingInfo(ctx context.Context, trackingNumber string) (*TrackingInfo, error)

    // CreateShipment 创建运单
    CreateShipment(ctx context.Context, req *CreateShipmentRequest) (*CreateShipmentResult, error)

    // CancelShipment 取消运单
    CancelShipment(ctx context.Context, shipmentID string) error
}
```

```go
// internal/plugin/contract_widget.go

package plugin

// WidgetPlugin 小部件插件接口
//
// 中文说明：
// - 用于向前端页面注入自定义内容块；
// - 例如：广告位、推荐商品、统计代码等。
type WidgetPlugin interface {
    Plugin

    // GetWidgetZones 获取可用的小部件区域
    GetWidgetZones() []string

    // RenderWidget 渲染小部件内容
    RenderWidget(ctx context.Context, zone string, params map[string]interface{}) (string, error)
}
```

## 三、插件元数据规范

### 3.1 plugin.json 示例

```json
{
  "group": "Payment",
  "friendly_name": "支付宝支付",
  "system_name": "Payment.Alipay",
  "version": "1.0.0",
  "supported_versions": ["1.0"],
  "author": "nop-go Team",
  "display_order": 1,
  "description": "支付宝网页支付、APP支付、扫码支付",
  "depends_on": [],
  "file_name": "plugin_payment_alipay.so"
}
```

### 3.2 目录结构

```
examples/nop-go/
├── plugins/                          # 插件目录
│   ├── payment-alipay/               # 支付宝插件
│   │   ├── plugin.json               # 插件元数据
│   │   ├── plugin.go                 # 插件实现
│   │   ├── config.go                 # 配置处理
│   │   ├── migrate/                  # 数据库迁移
│   │   │   └── 001_init.sql
│   │   └── go.mod
│   │
│   ├── payment-wechat/               # 微信支付插件
│   │   ├── plugin.json
│   │   └── plugin.go
│   │
│   ├── shipping-fedex/               # FedEx 配送插件
│   │   ├── plugin.json
│   │   └── plugin.go
│   │
│   └── widget-recommend/             # 商品推荐小部件
│       ├── plugin.json
│       └── plugin.go
│
├── internal/
│   └── plugin/                       # 插件系统核心
│       ├── contract.go               # 基础接口
│       ├── contract_payment.go       # 支付接口
│       ├── contract_shipping.go      # 配送接口
│       ├── contract_widget.go        # 小部件接口
│       ├── manager.go                # 插件管理器
│       ├── registry.go               # 插件注册表
│       └── loader.go                 # 插件加载器
```

## 四、插件管理器实现

### 4.1 插件管理器

```go
// internal/plugin/manager.go

package plugin

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "sort"
    "sync"

    "github.com/ngq/gorp/framework/contract"
)

// Manager 插件管理器
//
// 中文说明：
// - 负责插件的发现、加载、安装、卸载等生命周期管理；
// - 使用 gorp Container 注册插件服务；
// - 支持插件依赖排序。
type Manager struct {
    mu       sync.RWMutex
    container contract.Container
    plugins  map[string]Plugin     // systemName -> Plugin
    metas    map[string]*PluginMeta // systemName -> Meta
    pluginsDir string               // 插件目录
}

// NewManager 创建插件管理器
func NewManager(container contract.Container, pluginsDir string) *Manager {
    return &Manager{
        container:   container,
        plugins:     make(map[string]Plugin),
        metas:       make(map[string]*PluginMeta),
        pluginsDir:  pluginsDir,
    }
}

// Discover 扫描并发现所有插件
//
// 中文说明：
// - 扫描 plugins/ 目录下的所有 plugin.json；
// - 解析元数据并验证版本兼容性；
// - 不加载插件，只发现。
func (m *Manager) Discover() error {
    m.mu.Lock()
    defer m.mu.Unlock()

    // 清空已发现的元数据
    m.metas = make(map[string]*PluginMeta)

    // 扫描插件目录
    entries, err := os.ReadDir(m.pluginsDir)
    if err != nil {
        if os.IsNotExist(err) {
            return nil // 插件目录不存在，不是错误
        }
        return fmt.Errorf("read plugins directory: %w", err)
    }

    for _, entry := range entries {
        if !entry.IsDir() {
            continue
        }

        pluginDir := filepath.Join(m.pluginsDir, entry.Name())
        metaPath := filepath.Join(pluginDir, "plugin.json")

        // 读取 plugin.json
        data, err := os.ReadFile(metaPath)
        if err != nil {
            continue // 没有 plugin.json，跳过
        }

        var meta PluginMeta
        if err := json.Unmarshal(data, &meta); err != nil {
            continue // 解析失败，跳过
        }

        m.metas[meta.SystemName] = &meta
    }

    return nil
}

// Register 注册一个插件
//
// 中文说明：
// - 将插件实例注册到管理器；
// - 同时注册到 gorp Container。
func (m *Manager) Register(p Plugin) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    meta := p.Meta()
    if meta == nil {
        return fmt.Errorf("plugin meta is nil")
    }

    systemName := meta.SystemName
    if systemName == "" {
        return fmt.Errorf("plugin system_name is empty")
    }

    // 注册到管理器
    m.plugins[systemName] = p
    m.metas[systemName] = meta

    // 转换为 ServiceProvider 并注册到 Container
    sp := p.ToServiceProvider()
    if sp == nil {
        return fmt.Errorf("plugin %s ToServiceProvider returns nil", systemName)
    }

    if err := m.container.RegisterProvider(sp); err != nil {
        return fmt.Errorf("register plugin %s: %w", systemName, err)
    }

    return nil
}

// Install 安装插件
func (m *Manager) Install(ctx context.Context, systemName string) error {
    m.mu.RLock()
    p, ok := m.plugins[systemName]
    m.mu.RUnlock()

    if !ok {
        return fmt.Errorf("plugin %s not found", systemName)
    }

    if err := p.Install(ctx, m.container); err != nil {
        return fmt.Errorf("install plugin %s: %w", systemName, err)
    }

    // 更新元数据
    m.mu.Lock()
    if meta, ok := m.metas[systemName]; ok {
        meta.Installed = true
        meta.InstalledVersion = p.Meta().Version
    }
    m.mu.Unlock()

    return nil
}

// Uninstall 卸载插件
func (m *Manager) Uninstall(ctx context.Context, systemName string) error {
    m.mu.RLock()
    p, ok := m.plugins[systemName]
    m.mu.RUnlock()

    if !ok {
        return fmt.Errorf("plugin %s not found", systemName)
    }

    if err := p.Uninstall(ctx, m.container); err != nil {
        return fmt.Errorf("uninstall plugin %s: %w", systemName, err)
    }

    m.mu.Lock()
    delete(m.plugins, systemName)
    if meta, ok := m.metas[systemName]; ok {
        meta.Installed = false
        meta.InstalledVersion = ""
    }
    m.mu.Unlock()

    return nil
}

// Boot 启动所有已安装的插件
func (m *Manager) Boot(ctx context.Context) error {
    m.mu.RLock()
    defer m.mu.RUnlock()

    // 按依赖顺序排序
    order, err := m.resolveDependencyOrder()
    if err != nil {
        return err
    }

    for _, systemName := range order {
        p, ok := m.plugins[systemName]
        if !ok {
            continue
        }

        meta := p.Meta()
        if meta == nil || !meta.Installed {
            continue
        }

        if err := p.Boot(ctx, m.container); err != nil {
            return fmt.Errorf("boot plugin %s: %w", systemName, err)
        }
    }

    return nil
}

// GetPlugin 获取插件实例
func (m *Manager) GetPlugin(systemName string) (Plugin, bool) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    p, ok := m.plugins[systemName]
    return p, ok
}

// ListPlugins 列出所有插件
func (m *Manager) ListPlugins() []*PluginMeta {
    m.mu.RLock()
    defer m.mu.RUnlock()

    result := make([]*PluginMeta, 0, len(m.metas))
    for _, meta := range m.metas {
        result = append(result, meta)
    }

    // 按 DisplayOrder 排序
    sort.Slice(result, func(i, j int) bool {
        return result[i].DisplayOrder < result[j].DisplayOrder
    })

    return result
}

// resolveDependencyOrder 解析依赖顺序（拓扑排序）
func (m *Manager) resolveDependencyOrder() ([]string, error) {
    // 构建依赖图
    graph := make(map[string][]string)
    for systemName, meta := range m.metas {
        graph[systemName] = meta.DependsOn
    }

    // 拓扑排序
    var result []string
    visited := make(map[string]bool)
    visiting := make(map[string]bool)

    var visit func(string) error
    visit = func(node string) error {
        if visited[node] {
            return nil
        }
        if visiting[node] {
            return fmt.Errorf("circular dependency detected: %s", node)
        }

        visiting[node] = true
        for _, dep := range graph[node] {
            if err := visit(dep); err != nil {
                return err
            }
        }
        visiting[node] = false
        visited[node] = true
        result = append(result, node)
        return nil
    }

    for node := range graph {
        if err := visit(node); err != nil {
            return nil, err
        }
    }

    return result, nil
}
```

### 4.2 插件注册表

```go
// internal/plugin/registry.go

package plugin

import "sync"

// Registry 插件注册表（全局单例）
//
// 中文说明：
// - 用于按类型查找已注册的插件；
// - 例如：获取所有支付插件、所有配送插件。
type Registry struct {
    mu          sync.RWMutex
    byType      map[string][]Plugin  // pluginType -> []Plugin
    bySystemName map[string]Plugin    // systemName -> Plugin
}

var globalRegistry = &Registry{
    byType:       make(map[string][]Plugin),
    bySystemName: make(map[string]Plugin),
}

// GetRegistry 获取全局注册表
func GetRegistry() *Registry {
    return globalRegistry
}

// Register 注册插件到注册表
func (r *Registry) Register(p Plugin) {
    r.mu.Lock()
    defer r.mu.Unlock()

    systemName := p.Meta().SystemName
    pluginType := p.PluginType()

    r.bySystemName[systemName] = p
    r.byType[pluginType] = append(r.byType[pluginType], p)
}

// GetByType 按类型获取插件
func (r *Registry) GetByType(pluginType string) []Plugin {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return r.byType[pluginType]
}

// GetBySystemName 按系统名获取插件
func (r *Registry) GetBySystemName(systemName string) (Plugin, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    p, ok := r.bySystemName[systemName]
    return p, ok
}
```

## 五、示例插件实现

### 5.1 支付宝支付插件

```go
// plugins/payment-alipay/plugin.go

package main

import (
    "context"
    "encoding/json"
    "fmt"
    "os"

    "nop-go/internal/plugin"
    "github.com/ngq/gorp/framework/contract"
)

// AlipayPlugin 支付宝支付插件
type AlipayPlugin struct {
    meta   *plugin.PluginMeta
    config map[string]string
}

// New 创建插件实例
func New() *AlipayPlugin {
    return &AlipayPlugin{
        meta: &plugin.PluginMeta{
            Group:             "Payment",
            FriendlyName:      "支付宝支付",
            SystemName:        "Payment.Alipay",
            Version:           "1.0.0",
            SupportedVersions: []string{"1.0"},
            Author:            "nop-go Team",
            DisplayOrder:      1,
            Description:       "支付宝网页支付、APP支付、扫码支付",
        },
        config: make(map[string]string),
    }
}

func (p *AlipayPlugin) Meta() *plugin.PluginMeta { return p.meta }
func (p *AlipayPlugin) PluginType() string       { return "payment" }

func (p *AlipayPlugin) Install(ctx context.Context, c contract.Container) error {
    // 创建支付宝配置表
    // 实际项目中应该用 GORM 迁移
    fmt.Println("Installing Alipay plugin...")
    return nil
}

func (p *AlipayPlugin) Uninstall(ctx context.Context, c contract.Container) error {
    fmt.Println("Uninstalling Alipay plugin...")
    return nil
}

func (p *AlipayPlugin) Boot(ctx context.Context, c contract.Container) error {
    // 从配置服务读取支付宝配置
    cfg, err := c.Make(contract.ConfigKey)
    if err == nil {
        if config, ok := cfg.(contract.Config); ok {
            p.config["app_id"] = config.GetString("plugins.alipay.app_id")
            p.config["private_key"] = config.GetString("plugins.alipay.private_key")
        }
    }
    fmt.Println("Alipay plugin booted")
    return nil
}

func (p *AlipayPlugin) ToServiceProvider() contract.ServiceProvider {
    return &AlipayServiceProvider{plugin: p}
}

// 实现 PaymentMethod 接口
func (p *AlipayPlugin) ProcessPayment(ctx context.Context, req *plugin.ProcessPaymentRequest) (*plugin.ProcessPaymentResult, error) {
    // 调用支付宝 SDK 创建支付订单
    // 这里是简化示例
    return &plugin.ProcessPaymentResult{
        Success:       true,
        TransactionID: fmt.Sprintf("ALI%d", req.OrderID),
        RedirectURL:   fmt.Sprintf("https://openapi.alipay.com/gateway.do?order=%d", req.OrderID),
    }, nil
}

func (p *AlipayPlugin) Refund(ctx context.Context, req *plugin.RefundRequest) (*plugin.RefundResult, error) {
    return &plugin.RefundResult{Success: true}, nil
}

func (p *AlipayPlugin) Capture(ctx context.Context, req *plugin.CaptureRequest) (*plugin.CaptureResult, error) {
    return nil, fmt.Errorf("alipay does not support capture")
}

func (p *AlipayPlugin) Void(ctx context.Context, req *plugin.VoidRequest) (*plugin.VoidResult, error) {
    return nil, fmt.Errorf("alipay does not support void")
}

func (p *AlipayPlugin) GetConfiguration() []plugin.PaymentConfigItem {
    return []plugin.PaymentConfigItem{
        {Name: "app_id", Label: "应用ID", Type: "text", Required: true},
        {Name: "private_key", Label: "应用私钥", Type: "textarea", Required: true},
        {Name: "public_key", Label: "支付宝公钥", Type: "textarea", Required: true},
        {Name: "sandbox", Label: "沙箱模式", Type: "boolean", Required: false, Default: "false"},
    }
}

func (p *AlipayPlugin) ValidateConfiguration(config map[string]string) error {
    if config["app_id"] == "" {
        return fmt.Errorf("app_id is required")
    }
    if config["private_key"] == "" {
        return fmt.Errorf("private_key is required")
    }
    return nil
}

// AlipayServiceProvider gorp ServiceProvider 实现
type AlipayServiceProvider struct {
    plugin *AlipayPlugin
}

func (sp *AlipayServiceProvider) Name() string { return "plugin.payment.alipay" }
func (sp *AlipayServiceProvider) IsDefer() bool { return false }
func (sp *AlipayServiceProvider) Provides() []string {
    return []string{"plugin.payment.alipay"}
}

func (sp *AlipayServiceProvider) Register(c contract.Container) error {
    c.Bind("plugin.payment.alipay", func(c contract.Container) (interface{}, error) {
        return sp.plugin, nil
    }, true)
    return nil
}

func (sp *AlipayServiceProvider) Boot(c contract.Container) error {
    // 注册到全局注册表
    plugin.GetRegistry().Register(sp.plugin)
    return nil
}
```

## 六、集成到服务

### 6.1 在 payment-service 中使用

```go
// services/payment-service/cmd/main.go

func main() {
    // ... 初始化 gorp Container ...

    // 初始化插件管理器
    pluginManager := plugin.NewManager(c, "./plugins")
    
    // 发现插件
    if err := pluginManager.Discover(); err != nil {
        logger.Error(fmt.Sprintf("discover plugins: %v", err))
    }

    // 注册内置插件
    alipayPlugin := alipay.New()
    pluginManager.Register(alipayPlugin)

    // 安装插件
    if err := pluginManager.Install(context.Background(), "Payment.Alipay"); err != nil {
        logger.Error(fmt.Sprintf("install alipay: %v", err))
    }

    // 启动插件
    if err := pluginManager.Boot(context.Background()); err != nil {
        logger.Error(fmt.Sprintf("boot plugins: %v", err))
    }

    // ... 启动服务 ...
}
```

### 6.2 在业务代码中使用

```go
// services/payment-service/internal/biz/payment_usecase.go

func (uc *PaymentUseCase) ProcessPayment(ctx context.Context, method string, req *ProcessPaymentRequest) error {
    // 从注册表获取支付插件
    registry := plugin.GetRegistry()
    p, ok := registry.GetBySystemName(method)
    if !ok {
        return fmt.Errorf("payment method %s not found", method)
    }

    paymentPlugin, ok := p.(plugin.PaymentMethod)
    if !ok {
        return fmt.Errorf("plugin %s is not a payment method", method)
    }

    // 调用插件处理支付
    result, err := paymentPlugin.ProcessPayment(ctx, &plugin.ProcessPaymentRequest{
        OrderID:  req.OrderID,
        Amount:   req.Amount,
        Currency: req.Currency,
        // ...
    })

    if err != nil {
        return err
    }

    // 保存支付记录...
    return nil
}
```

## 七、插件开发规范

### 7.1 命名规范

| 类型 | 命名格式 | 示例 |
|------|---------|------|
| 支付插件 | `Payment.{Provider}` | `Payment.Alipay`, `Payment.Wechat` |
| 配送插件 | `Shipping.{Provider}` | `Shipping.FedEx`, `Shipping.DHL` |
| 小部件插件 | `Widgets.{Name}` | `Widgets.Recommend`, `Widgets.AdBanner` |
| 折扣插件 | `Discount.{Type}` | `Discount.BuyXGetY`, `Discount.Quantity` |
| 杂项插件 | `Misc.{Name}` | `Misc.GoogleAnalytics`, `Misc.Brevo` |

### 7.2 插件开发步骤

1. 创建插件目录 `plugins/{plugin-name}/`
2. 编写 `plugin.json` 元数据
3. 实现 `Plugin` 接口和业务接口（如 `PaymentMethod`）
4. 实现 `ToServiceProvider()` 返回 gorp ServiceProvider
5. 编写数据库迁移脚本（如需）
6. 编写单元测试

### 7.3 插件配置

插件配置统一放在应用配置文件中：

```yaml
# config/config.yaml
plugins:
  alipay:
    app_id: "2021001234567890"
    private_key: "MIIEvQIBADANBg..."
    public_key: "MIIBIjANBgkq..."
    sandbox: false
  
  wechat:
    app_id: "wx1234567890"
    mch_id: "1234567890"
    api_key: "abcdefg..."
```

## 八、后续规划

### Phase 1（当前）
- [x] 插件接口定义
- [x] 插件管理器
- [x] 插件注册表
- [ ] 支付宝支付插件
- [ ] 微信支付插件
- [ ] 示例配送插件

### Phase 2
- [ ] CLI 命令：`gorp plugin list/install/uninstall`
- [ ] 插件数据库迁移机制
- [ ] 插件配置管理 UI

### Phase 3
- [ ] Linux 下支持 Go plugin (.so) 动态加载
- [ ] 插件市场
- [ ] 第三方插件审核机制