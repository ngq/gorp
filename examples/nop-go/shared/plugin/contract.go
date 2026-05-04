// Package plugin nop-go 鎻掍欢绯荤粺鏍稿績鎺ュ彛
//
// 涓枃璇存槑:
// - 鎻掍欢鏈哄埗灞炰簬浜у搧灞傝璁?妗嗘灦灞?ServiceProvider + Container 宸茶冻澶?
// - 鎵€鏈変笟鍔℃彃浠堕兘瀹炵幇 Plugin 鎺ュ彛;
// - 閫氳繃 ToServiceProvider() 杞崲鍚庢敞鍐屽埌 gorp Container;
// - 澶嶇敤妗嗘灦宸叉湁鐨?ServiceProvider 鐢熷懡鍛ㄦ湡绠＄悊銆?
package plugin

import (
	"context"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Plugin 浜у搧绾ф彃浠跺熀纭€鎺ュ彛
//
// 涓枃璇存槑:
// - 鎵€鏈変笟鍔℃彃浠?payment/shipping/widget绛?閮藉繀椤诲疄鐜拌繖涓帴鍙?
// - 鎻掍欢閫氳繃 ToServiceProvider() 杞崲涓?gorp ServiceProvider;
// - 杩欐牱鍙互澶嶇敤妗嗘灦宸叉湁鐨?ServiceProvider 鐢熷懡鍛ㄦ湡绠＄悊(Register/Boot);
// - Install/Uninstall 鐢ㄤ簬鏁版嵁搴撹縼绉诲拰鍒濆鍖栭厤缃€?
type Plugin interface {
	// Meta 杩斿洖鎻掍欢鍏冩暟鎹?
	//
	// 涓枃璇存槑:
	// - 鍏冩暟鎹粠 plugin.json 鍔犺浇,鍖呭惈鎻掍欢鍩烘湰淇℃伅;
	// - 鍖呭惈: Group/FriendlyName/SystemName/Version 绛夈€?
	Meta() *PluginMeta

	// PluginType 杩斿洖鎻掍欢绫诲瀷
	//
	// 涓枃璇存槑:
	// - 绫诲瀷鐢ㄤ簬鍒嗙被绠＄悊,渚嬪: "payment", "shipping", "widget";
	// - Registry 鍙寜绫诲瀷鏌ユ壘鎻掍欢銆?
	PluginType() string

	// Install 鎻掍欢瀹夎鏃舵墽琛?
	//
	// 涓枃璇存槑:
	// - 鐢ㄤ簬鍒涘缓鏁版嵁搴撹〃銆佸垵濮嬪寲閰嶇疆銆佸啓鍏ラ粯璁ゆ暟鎹?
	// - 閫氬父鍦ㄩ娆″惎鐢ㄦ彃浠舵椂璋冪敤涓€娆?
	// - 搴斿寘鍚暟鎹簱杩佺Щ閫昏緫銆?
	Install(ctx context.Context, c runtimecontract.Container) error

	// Uninstall 鎻掍欢鍗歌浇鏃舵墽琛?
	//
	// 涓枃璇存槑:
	// - 鐢ㄤ簬娓呯悊鏁版嵁銆佸垹闄よ〃(璋ㄦ厧鎿嶄綔);
	// - 鐢ㄦ埛鏄庣‘鍗歌浇鏃惰皟鐢?
	// - 閫氬父寤鸿淇濈暀鏁版嵁,鍙爣璁颁负宸插嵏杞姐€?
	Uninstall(ctx context.Context, c runtimecontract.Container) error

	// Boot 鎻掍欢鍚姩鏃舵墽琛?
	//
	// 涓枃璇存槑:
	// - 姣忔鏈嶅姟鍚姩鏃惰皟鐢?
	// - 鐢ㄤ簬鍒濆鍖栬繍琛屾椂鐘舵€併€佽鍙栭厤缃€佸惎鍔?goroutine;
	// - 鍦?ServiceProvider.Boot 涓璋冪敤銆?
	Boot(ctx context.Context, c runtimecontract.Container) error

	// ToServiceProvider 杞崲涓?gorp ServiceProvider
	//
	// 涓枃璇存槑:
	// - 杩欐槸杩炴帴浜у搧鎻掍欢鍜屾鏋跺鍣ㄧ殑妗ユ;
	// - 杩斿洖鐨?ServiceProvider 浼氳娉ㄥ唽鍒?Container;
	// - 瀹炵幇涓簲杩斿洖涓€涓寘瑁呬簡鎻掍欢鏈韩鐨?ServiceProvider銆?
	ToServiceProvider() runtimecontract.ServiceProvider
}

// PluginMeta 鎻掍欢鍏冩暟鎹?
//
// 涓枃璇存槑:
// - 瀵瑰簲 plugin.json 鏂囦欢鍐呭;
// - 鐢ㄤ簬鎻掍欢鐨勫彂鐜般€佸睍绀哄拰鐗堟湰绠＄悊;
// - SystemName 鏄敮涓€鏍囪瘑,鍛藉悕鏍煎紡: {Type}.{Provider}銆?
type PluginMeta struct {
	// Group 鎻掍欢鍒嗙粍
	//
	// 涓枃璇存槑:
	// - 鐢ㄤ簬鍦ㄧ鐞嗙晫闈㈠垎缁勫睍绀?
	// - 渚嬪: "Payment", "Shipping", "Misc", "Widgets"銆?
	Group string `json:"group"`

	// FriendlyName 鍙嬪ソ鍚嶇О
	//
	// 涓枃璇存槑:
	// - 鏄剧ず缁欑敤鎴风殑鍚嶇О;
	// - 渚嬪: "鏀粯瀹濇敮浠?, "椤轰赴閫熻繍"銆?
	FriendlyName string `json:"friendly_name"`

	// SystemName 绯荤粺鍚嶇О
	//
	// 涓枃璇存槑:
	// - 鍞竴鏍囪瘑,鐢ㄤ簬鏌ユ壘鍜岄厤缃?
	// - 鍛藉悕鏍煎紡: {Type}.{Provider};
	// - 渚嬪: "Payment.Alipay", "Shipping.FedEx"銆?
	SystemName string `json:"system_name"`

	// Version 鎻掍欢鐗堟湰
	//
	// 涓枃璇存槑:
	// - 閬靛惊璇箟鍖栫増鏈鑼?
	// - 渚嬪: "1.0.0", "2.1.0"銆?
	Version string `json:"version"`

	// SupportedVersions 鏀寔鐨?nop-go 鐗堟湰
	//
	// 涓枃璇存槑:
	// - 鎸囨槑鍏煎鐨勪富鐗堟湰;
	// - 渚嬪: ["1.0", "1.1"]銆?
	SupportedVersions []string `json:"supported_versions"`

	// Author 浣滆€?
	Author string `json:"author"`

	// DisplayOrder 鏄剧ず椤哄簭
	//
	// 涓枃璇存槑:
	// - 鐢ㄤ簬鍒楄〃鎺掑簭;
	// - 鏁板瓧瓒婂皬鎺掑湪瓒婂墠闈€?
	DisplayOrder int `json:"display_order"`

	// Description 鎻掍欢鎻忚堪
	Description string `json:"description"`

	// DependsOn 渚濊禆鐨勫叾浠栨彃浠?
	//
	// 涓枃璇存槑:
	// - 鍛芥槑鍓嶇疆渚濊禆鐨?SystemName 鍒楄〃;
	// - Manager 浼氭寜渚濊禆椤哄簭鍔犺浇銆?
	DependsOn []string `json:"depends_on"`

	// FileName 缂栬瘧鍚庣殑鏂囦欢鍚?
	//
	// 涓枃璇存槑:
	// - 鐢ㄤ簬 Phase 2 鍔ㄦ€佸姞杞?.so);
	// - Phase 1 缂栬瘧杩涗富绋嬪簭鏃跺彲涓虹┖銆?
	FileName string `json:"file_name"`

	// Installed 鏄惁宸插畨瑁?
	//
	// 涓枃璇存槑:
	// - 鐢?Manager 缁存姢;
	// - true 琛ㄧず Install 宸叉墽琛岃繃銆?
	Installed bool `json:"installed"`

	// InstalledVersion 宸插畨瑁呯殑鐗堟湰
	//
	// 涓枃璇存槑:
	// - 鐢ㄤ簬妫€娴嬬増鏈洿鏂?
	// - 涓?Version 涓嶅悓鏃跺彲鑳介渶瑕佸崌绾с€?
	InstalledVersion string `json:"installed_version"`
}
