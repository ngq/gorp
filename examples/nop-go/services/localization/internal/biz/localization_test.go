// Package biz_test 本地化服务业务逻辑层单元测试。
//
// 使用 mock 仓储隔离依赖，测试 LocalizationUseCase 的核心业务逻辑。
package biz_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nop-go/services/localization/internal/biz"
)

// ============================================================
// Mock 仓储实现
// ============================================================

// MockLanguageRepository 语言仓储 mock 实现。
type MockLanguageRepository struct {
	Languages map[uint]*biz.Language
	NextID    uint
}

// NewMockLanguageRepository 创建 mock 语言仓储。
func NewMockLanguageRepository() *MockLanguageRepository {
	return &MockLanguageRepository{
		Languages: make(map[uint]*biz.Language),
		NextID:    1,
	}
}

func (m *MockLanguageRepository) Create(ctx context.Context, lang *biz.Language) error {
	lang.ID = m.NextID
	m.NextID++
	m.Languages[lang.ID] = lang
	return nil
}

func (m *MockLanguageRepository) GetByID(ctx context.Context, id uint) (*biz.Language, error) {
	lang, ok := m.Languages[id]
	if !ok {
		return nil, errors.New("language not found")
	}
	return lang, nil
}

func (m *MockLanguageRepository) List(ctx context.Context, page, size int) ([]*biz.Language, int64, error) {
	// 简单实现：返回所有语言
	var result []*biz.Language
	for _, lang := range m.Languages {
		result = append(result, lang)
	}
	return result, int64(len(result)), nil
}

func (m *MockLanguageRepository) Update(ctx context.Context, lang *biz.Language) error {
	m.Languages[lang.ID] = lang
	return nil
}

func (m *MockLanguageRepository) Delete(ctx context.Context, id uint) error {
	delete(m.Languages, id)
	return nil
}

// MockLocaleResourceRepository 本地化资源仓储 mock 实现。
type MockLocaleResourceRepository struct {
	Resources map[uint]*biz.LocaleResource
	NextID    uint
}

// NewMockLocaleResourceRepository 创建 mock 本地化资源仓储。
func NewMockLocaleResourceRepository() *MockLocaleResourceRepository {
	return &MockLocaleResourceRepository{
		Resources: make(map[uint]*biz.LocaleResource),
		NextID:    1,
	}
}

func (m *MockLocaleResourceRepository) Create(ctx context.Context, res *biz.LocaleResource) error {
	res.ID = m.NextID
	m.NextID++
	m.Resources[res.ID] = res
	return nil
}

func (m *MockLocaleResourceRepository) GetByID(ctx context.Context, id uint) (*biz.LocaleResource, error) {
	res, ok := m.Resources[id]
	if !ok {
		return nil, errors.New("resource not found")
	}
	return res, nil
}

func (m *MockLocaleResourceRepository) ListByLanguageID(ctx context.Context, languageID uint, page, size int) ([]*biz.LocaleResource, int64, error) {
	var result []*biz.LocaleResource
	for _, res := range m.Resources {
		if res.LanguageID == languageID {
			result = append(result, res)
		}
	}
	return result, int64(len(result)), nil
}

func (m *MockLocaleResourceRepository) Update(ctx context.Context, res *biz.LocaleResource) error {
	m.Resources[res.ID] = res
	return nil
}

func (m *MockLocaleResourceRepository) Delete(ctx context.Context, id uint) error {
	delete(m.Resources, id)
	return nil
}

func (m *MockLocaleResourceRepository) BatchCreate(ctx context.Context, resources []*biz.LocaleResource) error {
	for _, res := range resources {
		res.ID = m.NextID
		m.NextID++
		m.Resources[res.ID] = res
	}
	return nil
}

func (m *MockLocaleResourceRepository) ListAllByLanguageID(ctx context.Context, languageID uint) ([]*biz.LocaleResource, error) {
	var result []*biz.LocaleResource
	for _, res := range m.Resources {
		if res.LanguageID == languageID {
			result = append(result, res)
		}
	}
	return result, nil
}

// ============================================================
// 测试辅助函数
// ============================================================

// newTestUseCase 创建测试用的 LocalizationUseCase。
func newTestUseCase() (*biz.LocalizationUseCase, *MockLanguageRepository, *MockLocaleResourceRepository) {
	langRepo := NewMockLanguageRepository()
	resRepo := NewMockLocaleResourceRepository()

	uc := biz.NewLocalizationUseCase(langRepo, resRepo)
	return uc, langRepo, resRepo
}

// ============================================================
// 语言 CRUD 测试
// ============================================================

func TestCreateLanguage_Success(t *testing.T) {
	uc, _, _ := newTestUseCase()
	ctx := context.Background()

	lang, err := uc.CreateLanguage(ctx, "简体中文", "zh-CN", "zh", "cn.png", false, true, 1)
	assert.NoError(t, err)
	assert.NotNil(t, lang)
	assert.Equal(t, "简体中文", lang.Name)
	assert.Equal(t, "zh-CN", lang.LanguageCulture)
	assert.Equal(t, "zh", lang.UniqueSeoCode)
	assert.Equal(t, "cn.png", lang.FlagImageFileName)
	assert.False(t, lang.Rtl)
	assert.True(t, lang.IsActive)
	assert.Equal(t, 1, lang.DisplayOrder)
	assert.NotZero(t, lang.ID)
}

func TestGetLanguageByID_Success(t *testing.T) {
	uc, langRepo, _ := newTestUseCase()
	ctx := context.Background()

	// 先创建一个语言
	testLang := &biz.Language{
		Name:            "English",
		LanguageCulture: "en-US",
		UniqueSeoCode:   "en",
		IsActive:        true,
	}
	require.NoError(t, langRepo.Create(ctx, testLang))

	// 获取语言
	lang, err := uc.GetLanguageByID(ctx, testLang.ID)
	assert.NoError(t, err)
	assert.NotNil(t, lang)
	assert.Equal(t, "English", lang.Name)
}

func TestGetLanguageByID_NotFound(t *testing.T) {
	uc, _, _ := newTestUseCase()
	ctx := context.Background()

	lang, err := uc.GetLanguageByID(ctx, 999)
	assert.Error(t, err)
	assert.Nil(t, lang)
	assert.Contains(t, err.Error(), "not found")
}

func TestListLanguages_Success(t *testing.T) {
	uc, langRepo, _ := newTestUseCase()
	ctx := context.Background()

	// 创建多个语言
	lang1 := &biz.Language{Name: "简体中文", LanguageCulture: "zh-CN", IsActive: true}
	lang2 := &biz.Language{Name: "English", LanguageCulture: "en-US", IsActive: true}
	require.NoError(t, langRepo.Create(ctx, lang1))
	require.NoError(t, langRepo.Create(ctx, lang2))

	// 获取列表
	languages, total, err := uc.ListLanguages(ctx, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, languages, 2)
	assert.Equal(t, int64(2), total)
}

func TestListLanguages_Empty(t *testing.T) {
	uc, _, _ := newTestUseCase()
	ctx := context.Background()

	languages, total, err := uc.ListLanguages(ctx, 1, 10)
	assert.NoError(t, err)
	assert.Empty(t, languages)
	assert.Equal(t, int64(0), total)
}

func TestUpdateLanguage_Success(t *testing.T) {
	uc, langRepo, _ := newTestUseCase()
	ctx := context.Background()

	// 先创建一个语言
	testLang := &biz.Language{
		Name:            "English",
		LanguageCulture: "en-US",
		UniqueSeoCode:   "en",
		IsActive:        true,
	}
	require.NoError(t, langRepo.Create(ctx, testLang))

	// 更新语言
	updated, err := uc.UpdateLanguage(ctx, testLang.ID, "English (US)", "en-US", "en", "us.png", false, true, 2)
	assert.NoError(t, err)
	assert.NotNil(t, updated)
	assert.Equal(t, "English (US)", updated.Name)
	assert.Equal(t, "us.png", updated.FlagImageFileName)
	assert.Equal(t, 2, updated.DisplayOrder)
}

func TestUpdateLanguage_NotFound(t *testing.T) {
	uc, _, _ := newTestUseCase()
	ctx := context.Background()

	updated, err := uc.UpdateLanguage(ctx, 999, "Test", "test", "t", "t.png", false, true, 1)
	assert.Error(t, err)
	assert.Nil(t, updated)
}

func TestDeleteLanguage_Success(t *testing.T) {
	uc, langRepo, _ := newTestUseCase()
	ctx := context.Background()

	// 先创建一个语言
	testLang := &biz.Language{
		Name:            "Test Language",
		LanguageCulture: "test",
		IsActive:        true,
	}
	require.NoError(t, langRepo.Create(ctx, testLang))

	// 删除语言
	err := uc.DeleteLanguage(ctx, testLang.ID)
	assert.NoError(t, err)

	// 验证已删除
	_, err = langRepo.GetByID(ctx, testLang.ID)
	assert.Error(t, err)
}

func TestDeleteLanguage_NotFound(t *testing.T) {
	uc, _, _ := newTestUseCase()
	ctx := context.Background()

	// 删除不存在的语言（mock 实现返回 nil）
	err := uc.DeleteLanguage(ctx, 999)
	assert.NoError(t, err)
}

// ============================================================
// 本地化资源 CRUD 测试
// ============================================================

func TestAddResource_Success(t *testing.T) {
	uc, _, _ := newTestUseCase()
	ctx := context.Background()

	res, err := uc.AddResource(ctx, 1, "welcome.message", "欢迎访问我们的网站")
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, uint(1), res.LanguageID)
	assert.Equal(t, "welcome.message", res.ResourceName)
	assert.Equal(t, "欢迎访问我们的网站", res.ResourceValue)
	assert.NotZero(t, res.ID)
}

func TestGetResourceByID_Success(t *testing.T) {
	uc, _, resRepo := newTestUseCase()
	ctx := context.Background()

	// 先创建一个资源
	testRes := &biz.LocaleResource{
		LanguageID:    1,
		ResourceName:  "test.key",
		ResourceValue: "test value",
	}
	require.NoError(t, resRepo.Create(ctx, testRes))

	// 获取资源
	res, err := uc.GetResourceByID(ctx, testRes.ID)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "test.key", res.ResourceName)
}

func TestGetResourceByID_NotFound(t *testing.T) {
	uc, _, _ := newTestUseCase()
	ctx := context.Background()

	res, err := uc.GetResourceByID(ctx, 999)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestListResources_Success(t *testing.T) {
	uc, _, resRepo := newTestUseCase()
	ctx := context.Background()

	// 创建多个资源
	res1 := &biz.LocaleResource{LanguageID: 1, ResourceName: "key1", ResourceValue: "value1"}
	res2 := &biz.LocaleResource{LanguageID: 1, ResourceName: "key2", ResourceValue: "value2"}
	res3 := &biz.LocaleResource{LanguageID: 2, ResourceName: "key3", ResourceValue: "value3"}
	require.NoError(t, resRepo.Create(ctx, res1))
	require.NoError(t, resRepo.Create(ctx, res2))
	require.NoError(t, resRepo.Create(ctx, res3))

	// 获取语言1的资源
	resources, total, err := uc.ListResources(ctx, 1, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, resources, 2)
	assert.Equal(t, int64(2), total)
}

func TestListResources_Empty(t *testing.T) {
	uc, _, _ := newTestUseCase()
	ctx := context.Background()

	resources, total, err := uc.ListResources(ctx, 1, 1, 10)
	assert.NoError(t, err)
	assert.Empty(t, resources)
	assert.Equal(t, int64(0), total)
}

func TestUpdateResource_Success(t *testing.T) {
	uc, _, resRepo := newTestUseCase()
	ctx := context.Background()

	// 先创建一个资源
	testRes := &biz.LocaleResource{
		LanguageID:    1,
		ResourceName:  "old.key",
		ResourceValue: "old value",
	}
	require.NoError(t, resRepo.Create(ctx, testRes))

	// 更新资源
	updated, err := uc.UpdateResource(ctx, testRes.ID, "new.key", "new value")
	assert.NoError(t, err)
	assert.NotNil(t, updated)
	assert.Equal(t, "new.key", updated.ResourceName)
	assert.Equal(t, "new value", updated.ResourceValue)
}

func TestUpdateResource_NotFound(t *testing.T) {
	uc, _, _ := newTestUseCase()
	ctx := context.Background()

	updated, err := uc.UpdateResource(ctx, 999, "key", "value")
	assert.Error(t, err)
	assert.Nil(t, updated)
}

func TestDeleteResource_Success(t *testing.T) {
	uc, _, resRepo := newTestUseCase()
	ctx := context.Background()

	// 先创建一个资源
	testRes := &biz.LocaleResource{
		LanguageID:    1,
		ResourceName:  "test.key",
		ResourceValue: "test value",
	}
	require.NoError(t, resRepo.Create(ctx, testRes))

	// 删除资源
	err := uc.DeleteResource(ctx, testRes.ID)
	assert.NoError(t, err)

	// 验证已删除
	_, err = resRepo.GetByID(ctx, testRes.ID)
	assert.Error(t, err)
}

// ============================================================
// 导入导出测试
// ============================================================

func TestExportResources_Success(t *testing.T) {
	uc, _, resRepo := newTestUseCase()
	ctx := context.Background()

	// 创建多个资源
	res1 := &biz.LocaleResource{LanguageID: 1, ResourceName: "key1", ResourceValue: "value1"}
	res2 := &biz.LocaleResource{LanguageID: 1, ResourceName: "key2", ResourceValue: "value2"}
	res3 := &biz.LocaleResource{LanguageID: 2, ResourceName: "key3", ResourceValue: "value3"}
	require.NoError(t, resRepo.Create(ctx, res1))
	require.NoError(t, resRepo.Create(ctx, res2))
	require.NoError(t, resRepo.Create(ctx, res3))

	// 导出语言1的资源
	resources, err := uc.ExportResources(ctx, 1)
	assert.NoError(t, err)
	assert.Len(t, resources, 2)
}

func TestExportResources_Empty(t *testing.T) {
	uc, _, _ := newTestUseCase()
	ctx := context.Background()

	resources, err := uc.ExportResources(ctx, 1)
	assert.NoError(t, err)
	assert.Empty(t, resources)
}

func TestImportResources_Success(t *testing.T) {
	uc, _, resRepo := newTestUseCase()
	ctx := context.Background()

	// 导入资源
	importData := []struct {
		ResourceName  string
		ResourceValue string
	}{
		{ResourceName: "import.key1", ResourceValue: "导入值1"},
		{ResourceName: "import.key2", ResourceValue: "导入值2"},
		{ResourceName: "import.key3", ResourceValue: "导入值3"},
	}

	err := uc.ImportResources(ctx, 1, importData)
	assert.NoError(t, err)

	// 验证导入结果
	resources, total, err := resRepo.ListByLanguageID(ctx, 1, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, resources, 3)
	assert.Equal(t, int64(3), total)
}

func TestImportResources_Empty(t *testing.T) {
	uc, _, _ := newTestUseCase()
	ctx := context.Background()

	// 导入空资源列表
	err := uc.ImportResources(ctx, 1, []struct {
		ResourceName  string
		ResourceValue string
	}{})
	assert.NoError(t, err)
}

// ============================================================
// 时间戳测试
// ============================================================

func TestLanguage_Timestamps(t *testing.T) {
	uc, _, _ := newTestUseCase()
	ctx := context.Background()

	beforeCreate := time.Now()

	// 创建语言
	lang, err := uc.CreateLanguage(ctx, "测试语言", "test", "t", "t.png", false, true, 1)
	assert.NoError(t, err)

	// 验证时间戳已设置
	assert.False(t, lang.CreatedAt.IsZero())
	assert.False(t, lang.UpdatedAt.IsZero())
	assert.True(t, lang.CreatedAt.After(beforeCreate) || lang.CreatedAt.Equal(beforeCreate))

	// 更新语言
	time.Sleep(10 * time.Millisecond)
	beforeUpdate := time.Now()
	updated, err := uc.UpdateLanguage(ctx, lang.ID, "更新语言", "test2", "t2", "t2.png", true, false, 2)
	assert.NoError(t, err)

	// 验证更新时间已更新
	assert.True(t, updated.UpdatedAt.After(beforeUpdate) || updated.UpdatedAt.Equal(beforeUpdate))
}

func TestResource_Timestamps(t *testing.T) {
	uc, _, _ := newTestUseCase()
	ctx := context.Background()

	beforeCreate := time.Now()

	// 创建资源
	res, err := uc.AddResource(ctx, 1, "test.key", "test value")
	assert.NoError(t, err)

	// 验证时间戳已设置
	assert.False(t, res.CreatedAt.IsZero())
	assert.False(t, res.UpdatedAt.IsZero())
	assert.True(t, res.CreatedAt.After(beforeCreate) || res.CreatedAt.Equal(beforeCreate))

	// 更新资源
	time.Sleep(10 * time.Millisecond)
	beforeUpdate := time.Now()
	updated, err := uc.UpdateResource(ctx, res.ID, "new.key", "new value")
	assert.NoError(t, err)

	// 验证更新时间已更新
	assert.True(t, updated.UpdatedAt.After(beforeUpdate) || updated.UpdatedAt.Equal(beforeUpdate))
}

// ============================================================
// 边界条件测试
// ============================================================

func TestCreateLanguage_RTL(t *testing.T) {
	uc, _, _ := newTestUseCase()
	ctx := context.Background()

	// 测试 RTL 语言（如阿拉伯语）
	lang, err := uc.CreateLanguage(ctx, "العربية", "ar-SA", "ar", "sa.png", true, true, 1)
	assert.NoError(t, err)
	assert.True(t, lang.Rtl)
}

func TestCreateLanguage_EmptyFields(t *testing.T) {
	uc, _, _ := newTestUseCase()
	ctx := context.Background()

	// 测试空字段（业务层不做校验，由仓储层或校验层处理）
	lang, err := uc.CreateLanguage(ctx, "", "", "", "", false, false, 0)
	assert.NoError(t, err)
	assert.NotNil(t, lang)
}

func TestAddResource_EmptyValues(t *testing.T) {
	uc, _, _ := newTestUseCase()
	ctx := context.Background()

	// 测试空值
	res, err := uc.AddResource(ctx, 0, "", "")
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, uint(0), res.LanguageID)
}

func TestImportResources_LargeBatch(t *testing.T) {
	uc, _, resRepo := newTestUseCase()
	ctx := context.Background()

	// 测试大量导入
	var importData []struct {
		ResourceName  string
		ResourceValue string
	}
	for i := 0; i < 100; i++ {
		importData = append(importData, struct {
			ResourceName  string
			ResourceValue string
		}{
			ResourceName:  "key" + string(rune(i)),
			ResourceValue: "value" + string(rune(i)),
		})
	}

	err := uc.ImportResources(ctx, 1, importData)
	assert.NoError(t, err)

	// 验证导入结果
	resources, _, err := resRepo.ListByLanguageID(ctx, 1, 1, 1000)
	assert.NoError(t, err)
	assert.Len(t, resources, 100)
}
