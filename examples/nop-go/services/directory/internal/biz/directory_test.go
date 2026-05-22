// Package biz_test 目录服务业务逻辑层单元测试。
//
// 使用 mock 仓储隔离依赖，测试 CountryUseCase、StateUseCase、CurrencyUseCase 的核心业务逻辑。
package biz_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nop-go/services/directory/internal/biz"
)

// ============================================================
// Mock 仓储实现 - 国家
// ============================================================

// MockCountryRepository 国家仓储 mock 实现。
type MockCountryRepository struct {
	Countries map[uint]*biz.Country
	NextID    uint
}

// NewMockCountryRepository 创建 mock 国家仓储。
func NewMockCountryRepository() *MockCountryRepository {
	return &MockCountryRepository{
		Countries: make(map[uint]*biz.Country),
		NextID:    1,
	}
}

func (m *MockCountryRepository) Create(ctx context.Context, country *biz.Country) error {
	country.ID = m.NextID
	m.NextID++
	m.Countries[country.ID] = country
	return nil
}

func (m *MockCountryRepository) GetByID(ctx context.Context, id uint) (*biz.Country, error) {
	country, ok := m.Countries[id]
	if !ok {
		return nil, errors.New("country not found")
	}
	return country, nil
}

func (m *MockCountryRepository) List(ctx context.Context, page, size int) ([]*biz.Country, int64, error) {
	var result []*biz.Country
	for _, country := range m.Countries {
		result = append(result, country)
	}
	return result, int64(len(result)), nil
}

func (m *MockCountryRepository) Update(ctx context.Context, country *biz.Country) error {
	m.Countries[country.ID] = country
	return nil
}

func (m *MockCountryRepository) Delete(ctx context.Context, id uint) error {
	delete(m.Countries, id)
	return nil
}

// ============================================================
// Mock 仓储实现 - 省/州
// ============================================================

// MockStateRepository 省/州仓储 mock 实现。
type MockStateRepository struct {
	States map[uint]*biz.State
	NextID uint
}

// NewMockStateRepository 创建 mock 省/州仓储。
func NewMockStateRepository() *MockStateRepository {
	return &MockStateRepository{
		States: make(map[uint]*biz.State),
		NextID: 1,
	}
}

func (m *MockStateRepository) Create(ctx context.Context, state *biz.State) error {
	state.ID = m.NextID
	m.NextID++
	m.States[state.ID] = state
	return nil
}

func (m *MockStateRepository) GetByID(ctx context.Context, id uint) (*biz.State, error) {
	state, ok := m.States[id]
	if !ok {
		return nil, errors.New("state not found")
	}
	return state, nil
}

func (m *MockStateRepository) ListByCountryID(ctx context.Context, countryID uint, page, size int) ([]*biz.State, int64, error) {
	var result []*biz.State
	for _, state := range m.States {
		if state.CountryID == countryID {
			result = append(result, state)
		}
	}
	return result, int64(len(result)), nil
}

func (m *MockStateRepository) Update(ctx context.Context, state *biz.State) error {
	m.States[state.ID] = state
	return nil
}

func (m *MockStateRepository) Delete(ctx context.Context, id uint) error {
	delete(m.States, id)
	return nil
}

// ============================================================
// Mock 仓储实现 - 货币
// ============================================================

// MockCurrencyRepository 货币仓储 mock 实现。
type MockCurrencyRepository struct {
	Currencies map[uint]*biz.Currency
	NextID     uint
}

// NewMockCurrencyRepository 创建 mock 货币仓储。
func NewMockCurrencyRepository() *MockCurrencyRepository {
	return &MockCurrencyRepository{
		Currencies: make(map[uint]*biz.Currency),
		NextID:     1,
	}
}

func (m *MockCurrencyRepository) Create(ctx context.Context, currency *biz.Currency) error {
	currency.ID = m.NextID
	m.NextID++
	m.Currencies[currency.ID] = currency
	return nil
}

func (m *MockCurrencyRepository) GetByID(ctx context.Context, id uint) (*biz.Currency, error) {
	currency, ok := m.Currencies[id]
	if !ok {
		return nil, errors.New("currency not found")
	}
	return currency, nil
}

func (m *MockCurrencyRepository) List(ctx context.Context, page, size int) ([]*biz.Currency, int64, error) {
	var result []*biz.Currency
	for _, currency := range m.Currencies {
		result = append(result, currency)
	}
	return result, int64(len(result)), nil
}

func (m *MockCurrencyRepository) Update(ctx context.Context, currency *biz.Currency) error {
	m.Currencies[currency.ID] = currency
	return nil
}

func (m *MockCurrencyRepository) Delete(ctx context.Context, id uint) error {
	delete(m.Currencies, id)
	return nil
}

func (m *MockCurrencyRepository) BatchUpdateRates(ctx context.Context, rates []biz.CurrencyRateItem) error {
	for _, rate := range rates {
		if currency, ok := m.Currencies[rate.CurrencyID]; ok {
			currency.Rate = rate.Rate
		}
	}
	return nil
}

// ============================================================
// 测试辅助函数
// ============================================================

// newTestCountryUseCase 创建测试用的 CountryUseCase。
func newTestCountryUseCase() (*biz.CountryUseCase, *MockCountryRepository) {
	repo := NewMockCountryRepository()
	uc := biz.NewCountryUseCase(repo)
	return uc, repo
}

// newTestStateUseCase 创建测试用的 StateUseCase。
func newTestStateUseCase() (*biz.StateUseCase, *MockStateRepository) {
	repo := NewMockStateRepository()
	uc := biz.NewStateUseCase(repo)
	return uc, repo
}

// newTestCurrencyUseCase 创建测试用的 CurrencyUseCase。
func newTestCurrencyUseCase() (*biz.CurrencyUseCase, *MockCurrencyRepository) {
	repo := NewMockCurrencyRepository()
	uc := biz.NewCurrencyUseCase(repo)
	return uc, repo
}

// ============================================================
// 国家 CRUD 测试
// ============================================================

func TestCreateCountry_Success(t *testing.T) {
	uc, _ := newTestCountryUseCase()
	ctx := context.Background()

	country, err := uc.Create(ctx, "中国", "CN", "CHN", "{name}\n{address}", true)
	assert.NoError(t, err)
	assert.NotNil(t, country)
	assert.Equal(t, "中国", country.Name)
	assert.Equal(t, "CN", country.IsoCode2)
	assert.Equal(t, "CHN", country.IsoCode3)
	assert.Equal(t, "{name}\n{address}", country.AddressFormat)
	assert.True(t, country.PostcodeRequired)
	assert.NotZero(t, country.ID)
}

func TestGetCountryByID_Success(t *testing.T) {
	uc, repo := newTestCountryUseCase()
	ctx := context.Background()

	// 先创建一个国家
	testCountry := &biz.Country{
		Name:     "United States",
		IsoCode2: "US",
		IsoCode3: "USA",
	}
	require.NoError(t, repo.Create(ctx, testCountry))

	// 获取国家
	country, err := uc.GetByID(ctx, testCountry.ID)
	assert.NoError(t, err)
	assert.NotNil(t, country)
	assert.Equal(t, "United States", country.Name)
}

func TestGetCountryByID_NotFound(t *testing.T) {
	uc, _ := newTestCountryUseCase()
	ctx := context.Background()

	country, err := uc.GetByID(ctx, 999)
	assert.Error(t, err)
	assert.Nil(t, country)
	assert.Contains(t, err.Error(), "not found")
}

func TestListCountries_Success(t *testing.T) {
	uc, repo := newTestCountryUseCase()
	ctx := context.Background()

	// 创建多个国家
	country1 := &biz.Country{Name: "中国", IsoCode2: "CN", IsoCode3: "CHN"}
	country2 := &biz.Country{Name: "United States", IsoCode2: "US", IsoCode3: "USA"}
	require.NoError(t, repo.Create(ctx, country1))
	require.NoError(t, repo.Create(ctx, country2))

	// 获取列表
	countries, total, err := uc.List(ctx, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, countries, 2)
	assert.Equal(t, int64(2), total)
}

func TestListCountries_Empty(t *testing.T) {
	uc, _ := newTestCountryUseCase()
	ctx := context.Background()

	countries, total, err := uc.List(ctx, 1, 10)
	assert.NoError(t, err)
	assert.Empty(t, countries)
	assert.Equal(t, int64(0), total)
}

func TestUpdateCountry_Success(t *testing.T) {
	uc, repo := newTestCountryUseCase()
	ctx := context.Background()

	// 先创建一个国家
	testCountry := &biz.Country{
		Name:     "China",
		IsoCode2: "CN",
		IsoCode3: "CHN",
	}
	require.NoError(t, repo.Create(ctx, testCountry))

	// 更新国家
	updated, err := uc.Update(ctx, testCountry.ID, "中华人民共和国", "CN", "CHN", "{name}", true)
	assert.NoError(t, err)
	assert.NotNil(t, updated)
	assert.Equal(t, "中华人民共和国", updated.Name)
	assert.True(t, updated.PostcodeRequired)
}

func TestUpdateCountry_NotFound(t *testing.T) {
	uc, _ := newTestCountryUseCase()
	ctx := context.Background()

	updated, err := uc.Update(ctx, 999, "Test", "XX", "XXX", "", false)
	assert.Error(t, err)
	assert.Nil(t, updated)
}

func TestDeleteCountry_Success(t *testing.T) {
	uc, repo := newTestCountryUseCase()
	ctx := context.Background()

	// 先创建一个国家
	testCountry := &biz.Country{Name: "Test Country", IsoCode2: "TC", IsoCode3: "TST"}
	require.NoError(t, repo.Create(ctx, testCountry))

	// 删除国家
	err := uc.Delete(ctx, testCountry.ID)
	assert.NoError(t, err)

	// 验证已删除
	_, err = repo.GetByID(ctx, testCountry.ID)
	assert.Error(t, err)
}

// ============================================================
// 省/州 CRUD 测试
// ============================================================

func TestCreateState_Success(t *testing.T) {
	uc, _ := newTestStateUseCase()
	ctx := context.Background()

	state, err := uc.Create(ctx, 1, "北京市", "BJ")
	assert.NoError(t, err)
	assert.NotNil(t, state)
	assert.Equal(t, uint(1), state.CountryID)
	assert.Equal(t, "北京市", state.Name)
	assert.Equal(t, "BJ", state.IsoCode)
	assert.NotZero(t, state.ID)
}

func TestGetStateByID_Success(t *testing.T) {
	uc, repo := newTestStateUseCase()
	ctx := context.Background()

	// 先创建一个省/州
	testState := &biz.State{
		CountryID: 1,
		Name:      "California",
		IsoCode:   "CA",
	}
	require.NoError(t, repo.Create(ctx, testState))

	// 获取省/州
	state, err := uc.GetByID(ctx, testState.ID)
	assert.NoError(t, err)
	assert.NotNil(t, state)
	assert.Equal(t, "California", state.Name)
}

func TestGetStateByID_NotFound(t *testing.T) {
	uc, _ := newTestStateUseCase()
	ctx := context.Background()

	state, err := uc.GetByID(ctx, 999)
	assert.Error(t, err)
	assert.Nil(t, state)
}

func TestListStatesByCountryID_Success(t *testing.T) {
	uc, repo := newTestStateUseCase()
	ctx := context.Background()

	// 创建多个省/州
	state1 := &biz.State{CountryID: 1, Name: "北京市", IsoCode: "BJ"}
	state2 := &biz.State{CountryID: 1, Name: "上海市", IsoCode: "SH"}
	state3 := &biz.State{CountryID: 2, Name: "California", IsoCode: "CA"}
	require.NoError(t, repo.Create(ctx, state1))
	require.NoError(t, repo.Create(ctx, state2))
	require.NoError(t, repo.Create(ctx, state3))

	// 获取国家1的省/州列表
	states, total, err := uc.ListByCountryID(ctx, 1, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, states, 2)
	assert.Equal(t, int64(2), total)
}

func TestListStatesByCountryID_Empty(t *testing.T) {
	uc, _ := newTestStateUseCase()
	ctx := context.Background()

	states, total, err := uc.ListByCountryID(ctx, 1, 1, 10)
	assert.NoError(t, err)
	assert.Empty(t, states)
	assert.Equal(t, int64(0), total)
}

func TestUpdateState_Success(t *testing.T) {
	uc, repo := newTestStateUseCase()
	ctx := context.Background()

	// 先创建一个省/州
	testState := &biz.State{
		CountryID: 1,
		Name:      "Beijing",
		IsoCode:   "BJ",
	}
	require.NoError(t, repo.Create(ctx, testState))

	// 更新省/州
	updated, err := uc.Update(ctx, testState.ID, "北京市", "BJ")
	assert.NoError(t, err)
	assert.NotNil(t, updated)
	assert.Equal(t, "北京市", updated.Name)
}

func TestUpdateState_NotFound(t *testing.T) {
	uc, _ := newTestStateUseCase()
	ctx := context.Background()

	updated, err := uc.Update(ctx, 999, "Test", "TS")
	assert.Error(t, err)
	assert.Nil(t, updated)
}

func TestDeleteState_Success(t *testing.T) {
	uc, repo := newTestStateUseCase()
	ctx := context.Background()

	// 先创建一个省/州
	testState := &biz.State{CountryID: 1, Name: "Test State", IsoCode: "TS"}
	require.NoError(t, repo.Create(ctx, testState))

	// 删除省/州
	err := uc.Delete(ctx, testState.ID)
	assert.NoError(t, err)

	// 验证已删除
	_, err = repo.GetByID(ctx, testState.ID)
	assert.Error(t, err)
}

// ============================================================
// 货币 CRUD 测试
// ============================================================

func TestCreateCurrency_Success(t *testing.T) {
	uc, _ := newTestCurrencyUseCase()
	ctx := context.Background()

	currency, err := uc.Create(ctx, "人民币", "CNY", "¥", 1.0, true)
	assert.NoError(t, err)
	assert.NotNil(t, currency)
	assert.Equal(t, "人民币", currency.Name)
	assert.Equal(t, "CNY", currency.Code)
	assert.Equal(t, "¥", currency.Symbol)
	assert.Equal(t, 1.0, currency.Rate)
	assert.True(t, currency.IsActive)
	assert.NotZero(t, currency.ID)
}

func TestGetCurrencyByID_Success(t *testing.T) {
	uc, repo := newTestCurrencyUseCase()
	ctx := context.Background()

	// 先创建一个货币
	testCurrency := &biz.Currency{
		Name:     "US Dollar",
		Code:     "USD",
		Symbol:   "$",
		Rate:     0.14,
		IsActive: true,
	}
	require.NoError(t, repo.Create(ctx, testCurrency))

	// 获取货币
	currency, err := uc.GetByID(ctx, testCurrency.ID)
	assert.NoError(t, err)
	assert.NotNil(t, currency)
	assert.Equal(t, "US Dollar", currency.Name)
}

func TestGetCurrencyByID_NotFound(t *testing.T) {
	uc, _ := newTestCurrencyUseCase()
	ctx := context.Background()

	currency, err := uc.GetByID(ctx, 999)
	assert.Error(t, err)
	assert.Nil(t, currency)
}

func TestListCurrencies_Success(t *testing.T) {
	uc, repo := newTestCurrencyUseCase()
	ctx := context.Background()

	// 创建多个货币
	currency1 := &biz.Currency{Name: "人民币", Code: "CNY", Symbol: "¥", Rate: 1.0, IsActive: true}
	currency2 := &biz.Currency{Name: "US Dollar", Code: "USD", Symbol: "$", Rate: 0.14, IsActive: true}
	require.NoError(t, repo.Create(ctx, currency1))
	require.NoError(t, repo.Create(ctx, currency2))

	// 获取列表
	currencies, total, err := uc.List(ctx, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, currencies, 2)
	assert.Equal(t, int64(2), total)
}

func TestListCurrencies_Empty(t *testing.T) {
	uc, _ := newTestCurrencyUseCase()
	ctx := context.Background()

	currencies, total, err := uc.List(ctx, 1, 10)
	assert.NoError(t, err)
	assert.Empty(t, currencies)
	assert.Equal(t, int64(0), total)
}

func TestUpdateCurrency_Success(t *testing.T) {
	uc, repo := newTestCurrencyUseCase()
	ctx := context.Background()

	// 先创建一个货币
	testCurrency := &biz.Currency{
		Name:     "RMB",
		Code:     "CNY",
		Symbol:   "Y",
		Rate:     1.0,
		IsActive: true,
	}
	require.NoError(t, repo.Create(ctx, testCurrency))

	// 更新货币
	updated, err := uc.Update(ctx, testCurrency.ID, "人民币", "CNY", "¥", 1.0, true)
	assert.NoError(t, err)
	assert.NotNil(t, updated)
	assert.Equal(t, "人民币", updated.Name)
	assert.Equal(t, "¥", updated.Symbol)
}

func TestUpdateCurrency_NotFound(t *testing.T) {
	uc, _ := newTestCurrencyUseCase()
	ctx := context.Background()

	updated, err := uc.Update(ctx, 999, "Test", "TST", "T", 1.0, true)
	assert.Error(t, err)
	assert.Nil(t, updated)
}

func TestDeleteCurrency_Success(t *testing.T) {
	uc, repo := newTestCurrencyUseCase()
	ctx := context.Background()

	// 先创建一个货币
	testCurrency := &biz.Currency{Name: "Test Currency", Code: "TST", Symbol: "T", Rate: 1.0, IsActive: false}
	require.NoError(t, repo.Create(ctx, testCurrency))

	// 删除货币
	err := uc.Delete(ctx, testCurrency.ID)
	assert.NoError(t, err)

	// 验证已删除
	_, err = repo.GetByID(ctx, testCurrency.ID)
	assert.Error(t, err)
}

// ============================================================
// 汇率批量更新测试
// ============================================================

func TestApplyRates_Success(t *testing.T) {
	uc, repo := newTestCurrencyUseCase()
	ctx := context.Background()

	// 创建多个货币
	currency1 := &biz.Currency{Name: "人民币", Code: "CNY", Symbol: "¥", Rate: 1.0, IsActive: true}
	currency2 := &biz.Currency{Name: "US Dollar", Code: "USD", Symbol: "$", Rate: 0.14, IsActive: true}
	currency3 := &biz.Currency{Name: "Euro", Code: "EUR", Symbol: "€", Rate: 0.13, IsActive: true}
	require.NoError(t, repo.Create(ctx, currency1))
	require.NoError(t, repo.Create(ctx, currency2))
	require.NoError(t, repo.Create(ctx, currency3))

	// 批量更新汇率
	rates := []biz.CurrencyRateItem{
		{CurrencyID: currency1.ID, Rate: 1.0},
		{CurrencyID: currency2.ID, Rate: 0.15},
		{CurrencyID: currency3.ID, Rate: 0.14},
	}
	err := uc.ApplyRates(ctx, rates)
	assert.NoError(t, err)

	// 验证汇率已更新
	usd, _ := repo.GetByID(ctx, currency2.ID)
	assert.Equal(t, 0.15, usd.Rate)

	eur, _ := repo.GetByID(ctx, currency3.ID)
	assert.Equal(t, 0.14, eur.Rate)
}

func TestApplyRates_Empty(t *testing.T) {
	uc, _ := newTestCurrencyUseCase()
	ctx := context.Background()

	// 空汇率列表
	err := uc.ApplyRates(ctx, []biz.CurrencyRateItem{})
	assert.NoError(t, err)
}

func TestApplyRates_NonExistentCurrency(t *testing.T) {
	uc, _ := newTestCurrencyUseCase()
	ctx := context.Background()

	// 更新不存在的货币汇率（mock 实现忽略不存在的货币）
	rates := []biz.CurrencyRateItem{
		{CurrencyID: 999, Rate: 1.0},
	}
	err := uc.ApplyRates(ctx, rates)
	assert.NoError(t, err)
}

// ============================================================
// 时间戳测试
// ============================================================

func TestCountry_Timestamps(t *testing.T) {
	uc, _ := newTestCountryUseCase()
	ctx := context.Background()

	beforeCreate := time.Now()

	// 创建国家
	country, err := uc.Create(ctx, "中国", "CN", "CHN", "", true)
	assert.NoError(t, err)

	// 验证时间戳已设置
	assert.False(t, country.CreatedAt.IsZero())
	assert.False(t, country.UpdatedAt.IsZero())
	assert.True(t, country.CreatedAt.After(beforeCreate) || country.CreatedAt.Equal(beforeCreate))

	// 更新国家
	time.Sleep(10 * time.Millisecond)
	beforeUpdate := time.Now()
	updated, err := uc.Update(ctx, country.ID, "中华人民共和国", "CN", "CHN", "", true)
	assert.NoError(t, err)

	// 验证更新时间已更新
	assert.True(t, updated.UpdatedAt.After(beforeUpdate) || updated.UpdatedAt.Equal(beforeUpdate))
}

func TestState_Timestamps(t *testing.T) {
	uc, _ := newTestStateUseCase()
	ctx := context.Background()

	beforeCreate := time.Now()

	// 创建省/州
	state, err := uc.Create(ctx, 1, "北京市", "BJ")
	assert.NoError(t, err)

	// 验证时间戳已设置
	assert.False(t, state.CreatedAt.IsZero())
	assert.False(t, state.UpdatedAt.IsZero())
	assert.True(t, state.CreatedAt.After(beforeCreate) || state.CreatedAt.Equal(beforeCreate))

	// 更新省/州
	time.Sleep(10 * time.Millisecond)
	beforeUpdate := time.Now()
	updated, err := uc.Update(ctx, state.ID, "北京", "BJ")
	assert.NoError(t, err)

	// 验证更新时间已更新
	assert.True(t, updated.UpdatedAt.After(beforeUpdate) || updated.UpdatedAt.Equal(beforeUpdate))
}

func TestCurrency_Timestamps(t *testing.T) {
	uc, _ := newTestCurrencyUseCase()
	ctx := context.Background()

	beforeCreate := time.Now()

	// 创建货币
	currency, err := uc.Create(ctx, "人民币", "CNY", "¥", 1.0, true)
	assert.NoError(t, err)

	// 验证时间戳已设置
	assert.False(t, currency.CreatedAt.IsZero())
	assert.False(t, currency.UpdatedAt.IsZero())
	assert.True(t, currency.CreatedAt.After(beforeCreate) || currency.CreatedAt.Equal(beforeCreate))

	// 更新货币
	time.Sleep(10 * time.Millisecond)
	beforeUpdate := time.Now()
	updated, err := uc.Update(ctx, currency.ID, "中国人民币", "CNY", "¥", 1.0, true)
	assert.NoError(t, err)

	// 验证更新时间已更新
	assert.True(t, updated.UpdatedAt.After(beforeUpdate) || updated.UpdatedAt.Equal(beforeUpdate))
}

// ============================================================
// 边界条件测试
// ============================================================

func TestCreateCountry_EmptyFields(t *testing.T) {
	uc, _ := newTestCountryUseCase()
	ctx := context.Background()

	// 测试空字段
	country, err := uc.Create(ctx, "", "", "", "", false)
	assert.NoError(t, err)
	assert.NotNil(t, country)
}

func TestCreateState_EmptyFields(t *testing.T) {
	uc, _ := newTestStateUseCase()
	ctx := context.Background()

	// 测试空字段
	state, err := uc.Create(ctx, 0, "", "")
	assert.NoError(t, err)
	assert.NotNil(t, state)
}

func TestCreateCurrency_ZeroRate(t *testing.T) {
	uc, _ := newTestCurrencyUseCase()
	ctx := context.Background()

	// 测试零汇率
	currency, err := uc.Create(ctx, "Test", "TST", "T", 0, false)
	assert.NoError(t, err)
	assert.NotNil(t, currency)
	assert.Equal(t, 0.0, currency.Rate)
}

func TestCreateCurrency_NegativeRate(t *testing.T) {
	uc, _ := newTestCurrencyUseCase()
	ctx := context.Background()

	// 测试负汇率（业务层不做校验）
	currency, err := uc.Create(ctx, "Test", "TST", "T", -1.0, false)
	assert.NoError(t, err)
	assert.NotNil(t, currency)
	assert.Equal(t, -1.0, currency.Rate)
}

func TestApplyRates_LargeBatch(t *testing.T) {
	uc, repo := newTestCurrencyUseCase()
	ctx := context.Background()

	// 创建大量货币
	var currencyIDs []uint
	for i := 0; i < 100; i++ {
		currency := &biz.Currency{
			Name:     "Currency" + string(rune(i)),
			Code:     "C" + string(rune(i)),
			Symbol:   "$",
			Rate:     1.0,
			IsActive: true,
		}
		require.NoError(t, repo.Create(ctx, currency))
		currencyIDs = append(currencyIDs, currency.ID)
	}

	// 批量更新汇率
	var rates []biz.CurrencyRateItem
	for i, id := range currencyIDs {
		rates = append(rates, biz.CurrencyRateItem{
			CurrencyID: id,
			Rate:       float64(i) * 0.01,
		})
	}
	err := uc.ApplyRates(ctx, rates)
	assert.NoError(t, err)
}
