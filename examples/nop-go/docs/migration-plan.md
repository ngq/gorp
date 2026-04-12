# nopCommerce 剩余功能迁移计划

## 一、迁移概述

### 已完成迁移（12个服务）
| 服务 | 功能 |
|------|------|
| catalog-service | 商品/分类/品牌/属性/规格 |
| customer-service | 客户/地址/角色/认证 |
| order-service | 订单/退货/礼品卡 |
| payment-service | 支付/退款 |
| shipping-service | 配送/物流方式 |
| price-service | 价格/税率/折扣 |
| cms-service | 博客/新闻/论坛/主题内容 |
| notification-service | 通知/模板 |
| inventory-service | 库存/仓库 |
| cart-service | 购物车/愿望清单 |
| admin-service | 后台用户/角色/设置/日志 |
| gateway-service | API网关 |

### 新增已完成服务（5个）✅
| 服务 | 功能 | 完成日期 |
|------|------|---------|
| store-service | 多店铺/供应商管理 | 2026-04-09 |
| media-service | 媒体/图片存储 | 2026-04-09 |
| localization-service | 多语言/本地化 | 2026-04-09 |
| affiliate-service | 联盟营销/推广 | 2026-04-09 |
| seo-service | SEO/URL优化 | 2026-04-09 |

### 待迁移功能（8项）
| 功能 | nopCommerce 模块 | 建议归属 | 优先级 |
|------|-----------------|---------|--------|
| Gdpr | Nop.Services.Gdpr | 合并到 customer-service | P3 |
| Themes | Nop.Services.Themes | 新增 theme-service | P4 |
| Menus | Nop.Services.Menus | 合并到 cms-service | P2 |
| Polls | Nop.Services.Polls | 合并到 cms-service | P2 |
| ExportImport | Nop.Services.ExportImport | 新增 import-service | P3 |
| ArtificialIntelligence | Nop.Services.ArtificialIntelligence | 新增 ai-service | P4 |
| Html | Nop.Services.Html | 合并到 cms-service | P2 |

---

## 二、迁移批次与优先级

### 批次 P0：电商核心扩展（高优先级）

#### P0-1：store-service（多店铺/供应商）
- **目标**：支持多店铺、供应商管理
- **范围**：
  - Store 管理（店铺 CRUD、店铺配置）
  - Vendor 管理（供应商 CRUD、供应商商品关联）
  - 店铺-供应商关系
- **数据模型**：
  ```
  Store: ID, Name, URL, SSL, Hosts, DisplayOrder, DefaultLanguageId
  Vendor: ID, Name, Email, Description, AdminComment, Active, DisplayOrder
  StoreVendor: StoreID, VendorID
  ```
- **API**：
  - GET/POST/PUT/DELETE /api/v1/stores
  - GET/POST/PUT/DELETE /api/v1/vendors
- **依赖**：customer-service（供应商也是用户）
- **Verify**：go build + API 测试

#### P0-2：media-service（媒体/图片存储）
- **目标**：统一图片、文件上传与管理
- **范围**：
  - Picture 管理（图片上传、缩略图生成）
  - 存储后端抽象（本地、云存储）
  - 图片与商品/分类/品牌关联
- **数据模型**：
  ```
  Picture: ID, MimeType, SeoFilename, AltAttribute, TitleAttribute, IsNew
  ProductPicture: ProductID, PictureID, DisplayOrder
  CategoryPicture, ManufacturerPicture, VendorPicture
  ```
- **API**：
  - POST /api/v1/media/upload
  - GET /api/v1/media/:id
  - DELETE /api/v1/media/:id
- **依赖**：catalog-service
- **Verify**：图片上传下载测试

#### P0-3：localization-service（多语言/本地化）
- **目标**：支持多语言、多货币、时区
- **范围**：
  - Language 管理
  - LocaleStringResource（本地化资源）
  - Currency 管理
  - 国家/省份管理
- **数据模型**：
  ```
  Language: ID, Name, LanguageCulture, UniqueSeoCode, FlagImageFileName, Rtl, DefaultCurrencyId
  LocaleStringResource: LanguageId, ResourceName, ResourceValue
  Currency: ID, Name, CurrencyCode, Rate, DisplayOrder
  Country: ID, Name, AllowsBilling, AllowsShipping, TwoLetterIsoCode, ThreeLetterIsoCode
  StateProvince: CountryID, Name, Abbreviation
  ```
- **API**：
  - GET/POST/PUT/DELETE /api/v1/languages
  - GET/POST/PUT/DELETE /api/v1/currencies
  - GET /api/v1/localization/resources/:languageId
  - GET /api/v1/countries
- **依赖**：无
- **Verify**：多语言切换测试

---

### 批次 P1：营销与SEO（中优先级）

#### P1-1：affiliate-service（联盟营销）
- **目标**：支持联盟推广、佣金计算
- **范围**：
  - Affiliate 管理
  - AffiliateOrder 记录
  - 佣金计算
- **数据模型**：
  ```
  Affiliate: ID, Active, Name, URL, FriendlyUrlName, UseFriendlyUrl
  AffiliateOrder: OrderID, AffiliateID, Amount
  ```
- **API**：
  - GET/POST/PUT/DELETE /api/v1/affiliates
  - GET /api/v1/affiliates/:id/orders
  - GET /api/v1/affiliates/:id/commissions
- **依赖**：order-service, customer-service
- **Verify**：佣金计算测试

#### P1-2：seo-service（SEO优化）
- **目标**：URL友好化、元数据管理
- **范围**：
  - UrlRecord 管理（友好URL）
  - Meta信息管理
  - Sitemap 生成
- **数据模型**：
  ```
  UrlRecord: ID, EntityId, EntityName, Slug, LanguageId, IsActive
  MetaInfo: EntityID, EntityName, MetaTitle, MetaDescription, MetaKeywords
  ```
- **API**：
  - GET/POST/PUT/DELETE /api/v1/seo/urls
  - GET /api/v1/seo/sitemap
- **依赖**：catalog-service, cms-service
- **Verify**：URL重定向测试

---

### 批次 P2：内容与交互（中优先级）

#### P2-1：cms-service 扩展（菜单/投票/HTML）
- **目标**：补充CMS相关功能
- **新增范围**：
  - Menu 管理（导航菜单）
  - MenuItem（菜单项）
  - Poll 管理（投票）
  - PollAnswer（投票选项）
  - HtmlBody（HTML内容块）
- **数据模型**：
  ```
  Menu: ID, Name, SystemName, Active
  MenuItem: MenuID, ParentID, Name, Url, IconClass, DisplayOrder
  Poll: ID, Name, SystemKeyword, ShowOnHomepage, DisplayOrder
  PollAnswer: PollID, Name, NumberOfVotes, DisplayOrder
  PollVotingRecord: PollAnswerID, CustomerID
  HtmlBody: ID, Name, Content
  ```
- **API**：
  - GET/POST/PUT/DELETE /api/v1/cms/menus
  - GET/POST/PUT/DELETE /api/v1/cms/polls
  - POST /api/v1/cms/polls/:id/vote
  - GET/POST/PUT/DELETE /api/v1/cms/html
- **依赖**：customer-service
- **Verify**：菜单渲染、投票测试

---

### 批次 P3：合规与数据（低优先级）

#### P3-1：customer-service 扩展（GDPR）
- **目标**：GDPR 合规
- **新增范围**：
  - GdprConsent（同意项）
  - GdprLog（操作日志）
  - GdprRequest（数据请求）
- **数据模型**：
  ```
  GdprConsent: ID, Message, IsRequired, RequiredMessage, DisplayOrder
  GdprLog: CustomerID, IpAddress, CreatedOnUtc, RequestType, ConsentId
  GdprRequest: CustomerID, RequestType, RequestDetails, Status
  ```
- **API**：
  - GET /api/v1/customers/:id/gdpr/consents
  - POST /api/v1/customers/:id/gdpr/accept
  - POST /api/v1/customers/:id/gdpr/export
  - POST /api/v1/customers/:id/gdpr/delete
- **依赖**：无
- **Verify**：GDPR 流程测试

#### P3-2：import-service（导入导出）
- **目标**：数据批量导入导出
- **范围**：
  - ImportProfile（导入配置）
  - ExportProfile（导出配置）
  - Import/Export 执行
  - 支持格式：CSV、Excel、XML
- **数据模型**：
  ```
  ImportProfile: ID, Name, EntityTypeId, FilePath, Separator, SkipAttributeValidation
  ExportProfile: ID, Name, EntityTypeId, FilePath, ExportWithIds
  ImportHistory: ProfileID, StartTime, EndTime, Status, ErrorMessage
  ```
- **API**：
  - GET/POST/PUT/DELETE /api/v1/import/profiles
  - POST /api/v1/import/:id/execute
  - GET/POST/PUT/DELETE /api/v1/export/profiles
  - POST /api/v1/export/:id/execute
- **依赖**：catalog-service, customer-service
- **Verify**：导入导出测试

---

### 批次 P4：高级功能（最低优先级）

#### P4-1：theme-service（主题系统）
- **目标**：支持前端主题切换
- **范围**：
  - Theme 管理
  - ThemeVariable（主题变量）
  - 主题配置
- **数据模型**：
  ```
  Theme: ID, Name, Title, PreviewImageUrl, SupportRtl
  ThemeVariable: ThemeID, Name, Value, Type
  ```
- **API**：
  - GET/POST/PUT/DELETE /api/v1/themes
  - GET /api/v1/themes/:id/variables
- **依赖**：无
- **Verify**：主题切换测试

#### P4-2：ai-service（人工智能）
- **目标**：AI 辅助功能
- **范围**：
  - AI 对话接口
  - 商品推荐
  - 智能客服
  - 内容生成
- **数据模型**：
  ```
  AIConversation: ID, CustomerID, CreatedOnUtc
  AIMessage: ConversationID, Role, Content, CreatedOnUtc
  AIRecommendation: CustomerID, ProductID, Score
  ```
- **API**：
  - POST /api/v1/ai/chat
  - GET /api/v1/ai/recommendations/:customerId
  - POST /api/v1/ai/generate
- **依赖**：customer-service, catalog-service
- **Verify**：AI 接口测试

---

## 三、服务架构总览

迁移完成后，nop-go 将包含：

```
services/
├── store-service          # 多店铺/供应商（新增）
├── media-service          # 媒体/图片存储（新增）
├── localization-service   # 多语言/本地化（新增）
├── affiliate-service      # 联盟营销（新增）
├── seo-service            # SEO优化（新增）
├── import-service         # 导入导出（新增）
├── theme-service          # 主题系统（新增）
├── ai-service             # 人工智能（新增）
├── catalog-service        # 商品目录（扩展：关联media）
├── cms-service            # 内容管理（扩展：菜单/投票/HTML）
├── customer-service       # 客户（扩展：GDPR）
├── order-service          # 订单
├── payment-service        # 支付
├── shipping-service       # 物流
├── price-service          # 价格/税
├── notification-service   # 通知
├── inventory-service      # 库存
├── cart-service           # 购物车
├── admin-service          # 后台管理
└── gateway-service        # API网关
```

**总计：20个服务**（现有12个 + 新增8个）

---

## 四、执行计划

### 阶段一：P0 核心扩展（预计 3-5 天）
1. store-service（多店铺/供应商）
2. media-service（媒体存储）
3. localization-service（多语言）

### 阶段二：P1 营销SEO（预计 2-3 天）
1. affiliate-service（联盟营销）
2. seo-service（SEO优化）

### 阶段三：P2 内容交互（预计 2 天）
1. cms-service 扩展（菜单/投票/HTML）

### 阶段四：P3 合规数据（预计 2 天）
1. customer-service 扩展（GDPR）
2. import-service（导入导出）

### 阶段五：P4 高级功能（预计 2-3 天）
1. theme-service（主题系统）
2. ai-service（人工智能）

---

## 五、验收标准

每个服务完成后需验证：
1. `go build ./services/xxx-service/...` 通过
2. 数据模型迁移正常
3. API 可正常调用
4. 集成测试通过

---

## 六、备注

- 迁移过程中遵循 framework 现有能力（auth.jwt、bootstrap、orm.runtime）
- 新服务统一使用 `shared/bootstrap` 启动
- 认证统一使用 framework `auth.jwt`
- 配置统一使用 `auth.jwt.*` 格式