package feature_test

import (
	"testing"

	featurecontract "github.com/ngq/gorp/framework/contract/feature"
	featureprovider "github.com/ngq/gorp/framework/provider/feature"
	"github.com/stretchr/testify/require"
)

func TestFeatureService_Evaluations(t *testing.T) {
	svc := featureprovider.NewService(map[string]featureprovider.FlagRule{
		"new_checkout": {
			Enabled:       true,
			UserWhitelist: []string{"user100", "user101"},
			Percentage:    50,
		},
		"beta_banner": {
			Enabled:      true,
			DefaultValue: "Welcome Beta User",
		},
		"max_limit": {
			Enabled:      true,
			DefaultValue: 100,
		},
	})

	ctx := t.Context()

	// 1. 白名单用户，必定开启
	eval1 := svc.EvaluateBool(ctx, "new_checkout", false, featurecontract.EvaluationContext{UserID: "user100"})
	require.True(t, eval1)

	// 2. 非白名单用户，走 hash 百分比灰度
	eval2 := svc.EvaluateBool(ctx, "new_checkout", false, featurecontract.EvaluationContext{UserID: "user999"})
	// 结果取决于 Hash
	_ = eval2

	// 3. 字符串 Flag 测试
	strVal := svc.EvaluateString(ctx, "beta_banner", "default", featurecontract.EvaluationContext{})
	require.Equal(t, "Welcome Beta User", strVal)

	// 4. 整数 Flag 测试
	intVal := svc.EvaluateInt(ctx, "max_limit", 10, featurecontract.EvaluationContext{})
	require.Equal(t, 100, intVal)
}
