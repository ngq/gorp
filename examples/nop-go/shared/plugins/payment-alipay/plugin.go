// Package alipay 鏀粯瀹濇敮浠樻彃浠?
//
// 涓枃璇存槑:
// - 瀹炵幇鏀粯瀹濈綉椤垫敮浠樸€丄PP鏀粯銆佹壂鐮佹敮浠?
// - 鏀寔 PaymentMethod 鎺ュ彛;
// - 閫氳繃 ToServiceProvider 娉ㄥ唽鍒?gorp Container;
// - 閰嶇疆椤瑰寘鎷? app_id銆乸rivate_key銆乸ublic_key銆乻andbox 妯″紡銆?
package alipay

import (
	"context"
	"fmt"
	"time"

	"nop-go/shared/plugin"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// AlipayPlugin 鏀粯瀹濇敮浠樻彃浠跺疄鐜?
//
// 涓枃璇存槑:
// - 瀹炵幇 plugin.Plugin 鍜?plugin.PaymentMethod 鎺ュ彛;
// - 閫氳繃 config 瀛楁瀛樺偍浠庨厤缃湇鍔¤鍙栫殑鍙傛暟;
// - sandbox 妯″紡鐢ㄤ簬寮€鍙戞祴璇曘€?
type AlipayPlugin struct {
	meta   *plugin.PluginMeta
	config map[string]string
}

// New 鍒涘缓鏀粯瀹濇彃浠跺疄渚?
//
// 涓枃璇存槑:
// - 鍒涘缓鏃跺垵濮嬪寲榛樿鍏冩暟鎹?
// - 閰嶇疆浠?gorp Config 鏈嶅姟璇诲彇;
// - 閫氬父鍦ㄦ湇鍔″惎鍔ㄦ椂鍒涘缓骞舵敞鍐屻€?
func New() *AlipayPlugin {
	return &AlipayPlugin{
		meta: &plugin.PluginMeta{
			Group:             "Payment",
			FriendlyName:      "鏀粯瀹濇敮浠?,
			SystemName:        "Payment.Alipay",
			Version:           "1.0.0",
			SupportedVersions: []string{"1.0"},
			Author:            "nop-go Team",
			DisplayOrder:      1,
			Description:       "鏀粯瀹濈綉椤垫敮浠樸€丄PP鏀粯銆佹壂鐮佹敮浠?,
		},
		config: make(map[string]string),
	}
}

// Meta 杩斿洖鎻掍欢鍏冩暟鎹?
func (p *AlipayPlugin) Meta() *plugin.PluginMeta {
	return p.meta
}

// PluginType 杩斿洖鎻掍欢绫诲瀷
func (p *AlipayPlugin) PluginType() string {
	return "payment"
}

// Install 鎻掍欢瀹夎鏃舵墽琛?
//
// 涓枃璇存槑:
// - 鍒涘缓鏀粯瀹濇敮浠樻柟寮忚褰曞埌鏁版嵁搴?
// - 瀹為檯椤圭洰涓簲璇ョ敤 GORM 鍒涘缓 payment_methods 琛ㄨ褰?
// - 杩欓噷绠€鍖栦负鎵撳嵃鏃ュ織銆?
func (p *AlipayPlugin) Install(ctx context.Context, c runtimecontract.Container) error {
	// 瀹為檯椤圭洰涓簲璇?
	// 1. 鍒涘缓 payment_methods 璁板綍
	// 2. 鍒涘缓榛樿閰嶇疆
	// 3. 閫氱煡鐢ㄦ埛鍘婚厤缃敮浠樺弬鏁?

	// 鑾峰彇鏃ュ織鏈嶅姟(鍙€?
	logger, err := c.Make(observabilitycontract.LogKey)
	if err == nil {
		if log, ok := logger.(observabilitycontract.Logger); ok {
			log.Info("Installing Alipay payment plugin...")
		}
	}

	fmt.Println("[Alipay] Plugin installed")
	return nil
}

// Uninstall 鎻掍欢鍗歌浇鏃舵墽琛?
//
// 涓枃璇存槑:
// - 鏍囪 payment_methods 璁板綍涓哄凡鍗歌浇;
// - 閫氬父涓嶅缓璁垹闄ゆ暟鎹€?
func (p *AlipayPlugin) Uninstall(ctx context.Context, c runtimecontract.Container) error {
	fmt.Println("[Alipay] Plugin uninstalled")
	return nil
}

// Boot 鎻掍欢鍚姩鏃舵墽琛?
//
// 涓枃璇存槑:
// - 浠庨厤缃湇鍔¤鍙栨敮浠樺疂鍙傛暟;
// - 鍙傛暟鍖呮嫭: app_id銆乸rivate_key銆乸ublic_key銆乻andbox;
// - 鍒濆鍖栨敮浠樺疂 SDK 瀹㈡埛绔€?
func (p *AlipayPlugin) Boot(ctx context.Context, c runtimecontract.Container) error {
	// 浠庨厤缃湇鍔¤鍙栧弬鏁?
	cfg, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		// 閰嶇疆鏈嶅姟涓嶅彲鐢ㄦ椂浣跨敤绌洪厤缃?
		// 瀹為檯椤圭洰涓簲璇ヨ繑鍥為敊璇垨浣跨敤榛樿鍊?
		fmt.Println("[Alipay] Config service not available, using defaults")
		return nil
	}

	config, ok := cfg.(datacontract.Config)
	if !ok {
		return nil
	}

	// 璇诲彇鏀粯瀹濋厤缃?
	p.config["app_id"] = config.GetString("plugins.alipay.app_id")
	p.config["private_key"] = config.GetString("plugins.alipay.private_key")
	p.config["public_key"] = config.GetString("plugins.alipay.public_key")
	p.config["sandbox"] = config.GetString("plugins.alipay.sandbox")

	fmt.Printf("[Alipay] Plugin booted, app_id: %s, sandbox: %s\n",
		p.config["app_id"], p.config["sandbox"])

	return nil
}

// ToServiceProvider 杞崲涓?gorp ServiceProvider
//
// 涓枃璇存槑:
// - 杩欐槸杩炴帴浜у搧鎻掍欢鍜屾鏋跺鍣ㄧ殑妗ユ;
// - 杩斿洖鐨?ServiceProvider 浼氳娉ㄥ唽鍒?Container;
// - Boot 鏃跺皢鎻掍欢娉ㄥ唽鍒板叏灞€ Registry銆?
func (p *AlipayPlugin) ToServiceProvider() runtimecontract.ServiceProvider {
	return &AlipayServiceProvider{plugin: p}
}

// ProcessPayment 澶勭悊鏀粯
//
// 涓枃璇存槑:
// - 鍒涘缓鏀粯瀹濇敮浠樿鍗?
// - 杩斿洖鏀粯閾炬帴渚涘墠绔烦杞?
// - 瀹為檯椤圭洰涓簲璋冪敤鏀粯瀹?SDK銆?
func (p *AlipayPlugin) ProcessPayment(ctx context.Context, req *plugin.ProcessPaymentRequest) (*plugin.ProcessPaymentResult, error) {
	// 妫€鏌ラ厤缃?
	if p.config["app_id"] == "" {
		return nil, fmt.Errorf("alipay app_id not configured")
	}

	// 瀹為檯椤圭洰涓簲璇?
	// 1. 璋冪敤鏀粯瀹?SDK 鍒涘缓璁㈠崟
	// 2. 鐢熸垚鏀粯閾炬帴鎴栦簩缁寸爜
	// 3. 璁板綍璇锋眰鏃ュ織

	// 绠€鍖栧疄鐜?鐢熸垚妯℃嫙鏀粯閾炬帴
	txnID := fmt.Sprintf("ALI%s%d", time.Now().Format("20060102150405"), req.OrderID)

	// 鍒ゆ柇娌欑妯″紡
	baseURL := "https://openapi.alipay.com/gateway.do"
	if p.config["sandbox"] == "true" {
		baseURL = "https://openapi.alipaydev.com/gateway.do"
	}

	redirectURL := fmt.Sprintf("%s?app_id=%s&order_id=%d&amount=%.2f",
		baseURL, p.config["app_id"], req.OrderID, req.Amount)

	return &plugin.ProcessPaymentResult{
		Success:       true,
		TransactionID: txnID,
		RedirectURL:   redirectURL,
		RawData: map[string]interface{}{
			"app_id":     p.config["app_id"],
			"order_id":   req.OrderID,
			"amount":     req.Amount,
			"currency":   req.Currency,
			"created_at": time.Now().Format("2006-01-02 15:04:05"),
		},
	}, nil
}

// Refund 閫€娆?
//
// 涓枃璇存槑:
// - 鏀粯瀹濇敮鎸佸叏棰濆拰閮ㄥ垎閫€娆?
// - 闇€瑕佸師浜ゆ槗娴佹按鍙?
// - 杩斿洖閫€娆炬祦姘村彿銆?
func (p *AlipayPlugin) Refund(ctx context.Context, req *plugin.RefundRequest) (*plugin.RefundResult, error) {
	if p.config["app_id"] == "" {
		return nil, fmt.Errorf("alipay app_id not configured")
	}

	// 瀹為檯椤圭洰涓簲璇ヨ皟鐢ㄦ敮浠樺疂閫€娆炬帴鍙?
	refundTxnID := fmt.Sprintf("REF%s%s", time.Now().Format("20060102150405"), req.TransactionID)

	return &plugin.RefundResult{
		Success:             true,
		RefundTransactionID: refundTxnID,
	}, nil
}

// Capture 鎹曡幏棰勬巿鏉?
//
// 涓枃璇存槑:
// - 鏀粯瀹濅笉鏀寔棰勬巿鏉冩ā寮?
// - 杩斿洖閿欒鎻愮ず銆?
func (p *AlipayPlugin) Capture(ctx context.Context, req *plugin.CaptureRequest) (*plugin.CaptureResult, error) {
	return nil, fmt.Errorf("alipay does not support capture (pre-authorization)")
}

// Void 鍙栨秷棰勬巿鏉?
//
// 涓枃璇存槑:
// - 鏀粯瀹濅笉鏀寔棰勬巿鏉冩ā寮?
// - 杩斿洖閿欒鎻愮ず銆?
func (p *AlipayPlugin) Void(ctx context.Context, req *plugin.VoidRequest) (*plugin.VoidResult, error) {
	return nil, fmt.Errorf("alipay does not support void (cancel pre-authorization)")
}

// GetConfiguration 鑾峰彇鏀粯閰嶇疆椤?
//
// 涓枃璇存槑:
// - 杩斿洖闇€瑕佸湪绠＄悊鍚庡彴閰嶇疆鐨勫瓧娈靛垪琛?
// - 鐢ㄤ簬鐢熸垚鍔ㄦ€侀厤缃〃鍗?
// - Type 鍐冲畾杈撳叆鎺т欢绫诲瀷銆?
func (p *AlipayPlugin) GetConfiguration() []plugin.PaymentConfigItem {
	return []plugin.PaymentConfigItem{
		{
			Name:     "app_id",
			Label:    "搴旂敤ID (AppID)",
			Type:     "text",
			Required: true,
			HelpText: "鍦ㄦ敮浠樺疂寮€鏀惧钩鍙板垱寤哄簲鐢ㄥ悗鑾峰彇",
		},
		{
			Name:     "private_key",
			Label:    "搴旂敤绉侀挜",
			Type:     "textarea",
			Required: true,
			HelpText: "浣跨敤 RSA2 绠楁硶鐢熸垚鐨勫簲鐢ㄧ閽?,
		},
		{
			Name:     "public_key",
			Label:    "鏀粯瀹濆叕閽?,
			Type:     "textarea",
			Required: true,
			HelpText: "鏀粯瀹濆簲鐢ㄧ殑鍏挜,鐢ㄤ簬楠岀",
		},
		{
			Name:     "sandbox",
			Label:    "娌欑妯″紡",
			Type:     "boolean",
			Required: false,
			Default:  "false",
			HelpText: "寮€鍚悗浣跨敤鏀粯瀹濇矙绠辩幆澧冭繘琛屾祴璇?,
		},
		{
			Name:     "notify_url",
			Label:    "寮傛閫氱煡URL",
			Type:     "text",
			Required: false,
			HelpText: "鏀粯缁撴灉寮傛閫氱煡鍦板潃,闇€澶栫綉鍙闂?,
		},
		{
			Name:     "return_url",
			Label:    "鍚屾杩斿洖URL",
			Type:     "text",
			Required: false,
			HelpText: "鏀粯瀹屾垚鍚庢祻瑙堝櫒璺宠浆鍦板潃",
		},
	}
}

// ValidateConfiguration 楠岃瘉閰嶇疆鏄惁姝ｇ‘
//
// 涓枃璇存槑:
// - 淇濆瓨閰嶇疆鍓嶉獙璇?
// - 妫€鏌ュ繀濉」銆佹牸寮忕瓑;
// - 杩斿洖鍏蜂綋鐨勯敊璇俊鎭究浜庝慨姝ｃ€?
func (p *AlipayPlugin) ValidateConfiguration(config map[string]string) error {
	if config["app_id"] == "" {
		return fmt.Errorf("app_id 鏄繀濉」")
	}
	if len(config["app_id"]) < 16 {
		return fmt.Errorf("app_id 鏍煎紡涓嶆纭?搴斾负16浣嶆暟瀛?)
	}
	if config["private_key"] == "" {
		return fmt.Errorf("private_key 鏄繀濉」")
	}
	if config["public_key"] == "" {
		return fmt.Errorf("public_key 鏄繀濉」")
	}

	// 瀹為檯椤圭洰涓繕鍙互:
	// 1. 楠岃瘉绉侀挜鏍煎紡
	// 2. 璋冪敤鏀粯瀹濇帴鍙ｉ獙璇侀厤缃湁鏁堟€?
	// 3. 妫€鏌?notify_url 鏄惁澶栫綉鍙闂?

	return nil
}

// AlipayServiceProvider gorp ServiceProvider 瀹炵幇
//
// 涓枃璇存槑:
// - 鍖呰 AlipayPlugin 瀹炵幇 ServiceProvider 鎺ュ彛;
// - Register 鏃剁粦瀹氭彃浠跺疄渚嬪埌 Container;
// - Boot 鏃舵敞鍐屽埌鍏ㄥ眬 Registry銆?
type AlipayServiceProvider struct {
	plugin *AlipayPlugin
}

// Name 杩斿洖 ServiceProvider 鍚嶇О
func (sp *AlipayServiceProvider) Name() string {
	return "plugin.payment.alipay"
}

// IsDefer 鏄惁寤惰繜鍔犺浇
//
// 涓枃璇存槑:
// - 鏀粯鎻掍欢閫氬父闇€瑕佺珛鍗冲姞杞?
// - 杩斿洖 false 琛ㄧず娉ㄥ唽鏃剁珛鍗虫墽琛?Register/Boot銆?
func (sp *AlipayServiceProvider) IsDefer() bool {
	return false
}

// Provides 杩斿洖鎻愪緵鐨勬湇鍔?Key 鍒楄〃
//
// 涓枃璇存槑:
// - 澹版槑姝?Provider 鎻愪緵鍝簺鏈嶅姟;
// - Container.Make("plugin.payment.alipay") 鍙幏鍙栨彃浠跺疄渚嬨€?
func (sp *AlipayServiceProvider) Provides() []string {
	return []string{"plugin.payment.alipay"}
}

// Register 娉ㄥ唽闃舵
//
// 涓枃璇存槑:
// - 灏嗘彃浠跺疄渚嬬粦瀹氬埌 Container;
// - singleton=true 琛ㄧず鍗曚緥妯″紡;
// - 涓氬姟浠ｇ爜鍙€氳繃 Make 鑾峰彇鎻掍欢銆?
func (sp *AlipayServiceProvider) Register(c runtimecontract.Container) error {
	c.Bind("plugin.payment.alipay", func(c runtimecontract.Container) (interface{}, error) {
		return sp.plugin, nil
	}, true)
	return nil
}

// Boot 鍚姩闃舵
//
// 涓枃璇存槑:
// - 娉ㄥ唽鍒板叏灞€ Registry;
// - 杩欐牱鍏朵粬鏈嶅姟鍙€氳繃 Registry 鏌ユ壘鎻掍欢;
// - 渚嬪: GetRegistry().GetPaymentMethod("Payment.Alipay")銆?
func (sp *AlipayServiceProvider) Boot(c runtimecontract.Container) error {
	// 娉ㄥ唽鍒板叏灞€ Registry
	plugin.GetRegistry().Register(sp.plugin)

	// 鎵ц鎻掍欢鐨?Boot 鏂规硶
	ctx := context.Background()
	return sp.plugin.Boot(ctx, c)
}