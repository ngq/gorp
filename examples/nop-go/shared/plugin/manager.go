// Package plugin 鎻掍欢绠＄悊鍣?
package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Manager 鎻掍欢绠＄悊鍣?
//
// 涓枃璇存槑:
// - 璐熻矗鎻掍欢鐨勫彂鐜般€佸姞杞姐€佸畨瑁呫€佸嵏杞界瓑鐢熷懡鍛ㄦ湡绠＄悊;
// - 浣跨敤 gorp Container 娉ㄥ唽鎻掍欢鏈嶅姟;
// - 鏀寔鎻掍欢渚濊禆鎺掑簭;
// - Discover 鍙彂鐜颁笉鍔犺浇,Register 鎵嶆敞鍐屽埌 Container銆?
type Manager struct {
	mu         sync.RWMutex
	container  runtimecontract.Container
	plugins    map[string]Plugin      // systemName -> Plugin 瀹炰緥
	metas      map[string]*PluginMeta // systemName -> Meta 鍏冩暟鎹?
	pluginsDir string                 // 鎻掍欢鐩綍璺緞
}

// NewManager 鍒涘缓鎻掍欢绠＄悊鍣?
//
// 涓枃璇存槑:
// - container 鏄?gorp 渚濊禆娉ㄥ叆瀹瑰櫒;
// - pluginsDir 鏄彃浠剁洰褰曡矾寰?閫氬父鏄?"./plugins";
// - 鍒涘缓鍚庨渶瑕佽皟鐢?Discover() 鍙戠幇鍙敤鎻掍欢銆?
func NewManager(container runtimecontract.Container, pluginsDir string) *Manager {
	return &Manager{
		container:  container,
		plugins:    make(map[string]Plugin),
		metas:      make(map[string]*PluginMeta),
		pluginsDir: pluginsDir,
	}
}

// Discover 鎵弿骞跺彂鐜版墍鏈夋彃浠?
//
// 涓枃璇存槑:
// - 鎵弿 plugins/ 鐩綍涓嬬殑鎵€鏈?plugin.json;
// - 瑙ｆ瀽鍏冩暟鎹苟楠岃瘉鐗堟湰鍏煎鎬?
// - 鍙彂鐜颁笉鍔犺浇,涓嶅垱寤烘彃浠跺疄渚?
// - 瀹為檯娉ㄥ唽闇€瑕佽皟鐢?Register() 鏂规硶銆?
func (m *Manager) Discover() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 娓呯┖宸插彂鐜扮殑鍏冩暟鎹?
	m.metas = make(map[string]*PluginMeta)

	// 妫€鏌ユ彃浠剁洰褰曟槸鍚﹀瓨鍦?
	entries, err := os.ReadDir(m.pluginsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 鎻掍欢鐩綍涓嶅瓨鍦?涓嶆槸閿欒
		}
		return fmt.Errorf("read plugins directory: %w", err)
	}

	// 鎵弿姣忎釜瀛愮洰褰?
	for _, entry := range entries {
		if !entry.IsDir() {
			continue // 鍙鐞嗙洰褰?
		}

		pluginDir := filepath.Join(m.pluginsDir, entry.Name())
		metaPath := filepath.Join(pluginDir, "plugin.json")

		// 璇诲彇 plugin.json
		data, err := os.ReadFile(metaPath)
		if err != nil {
			continue // 娌℃湁 plugin.json,璺宠繃
		}

		var meta PluginMeta
		if err := json.Unmarshal(data, &meta); err != nil {
			continue // 瑙ｆ瀽澶辫触,璺宠繃
		}

		// 楠岃瘉鍩烘湰瀛楁
		if meta.SystemName == "" {
			continue // system_name 蹇呭～
		}

		m.metas[meta.SystemName] = &meta
	}

	return nil
}

// Register 娉ㄥ唽涓€涓彃浠跺疄渚?
//
// 涓枃璇存槑:
// - 灏嗘彃浠跺疄渚嬫敞鍐屽埌绠＄悊鍣?
// - 鍚屾椂杞崲涓?ServiceProvider 娉ㄥ唽鍒?gorp Container;
// - 娉ㄥ唽鍚庢彃浠跺彲閫氳繃 Container.Make() 鑾峰彇;
// - 娉ㄥ唽鎴愬姛鍚庝細鑷姩娣诲姞鍒板叏灞€ Registry銆?
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

	// 娉ㄥ唽鍒扮鐞嗗櫒鍐呴儴鏄犲皠
	m.plugins[systemName] = p
	m.metas[systemName] = meta

	// 杞崲涓?ServiceProvider
	sp := p.ToServiceProvider()
	if sp == nil {
		return fmt.Errorf("plugin %s ToServiceProvider returns nil", systemName)
	}

	// 娉ㄥ唽鍒?Container
	if err := m.container.RegisterProvider(sp); err != nil {
		return fmt.Errorf("register plugin %s to container: %w", systemName, err)
	}

	// 娉ㄥ唽鍒板叏灞€ Registry
	GetRegistry().Register(p)

	return nil
}

// Install 瀹夎鎻掍欢
//
// 涓枃璇存槑:
// - 璋冪敤鎻掍欢鐨?Install 鏂规硶;
// - 鐢ㄤ簬鍒涘缓鏁版嵁搴撹〃銆佸垵濮嬪寲閰嶇疆;
// - 瀹夎鍚庢爣璁?meta.Installed = true;
// - 閫氬父鍙渶璋冪敤涓€娆°€?
func (m *Manager) Install(ctx context.Context, systemName string) error {
	m.mu.RLock()
	p, ok := m.plugins[systemName]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("plugin %s not registered", systemName)
	}

	// 璋冪敤鎻掍欢瀹夎鏂规硶
	if err := p.Install(ctx, m.container); err != nil {
		return fmt.Errorf("install plugin %s: %w", systemName, err)
	}

	// 鏇存柊鍏冩暟鎹?
	m.mu.Lock()
	if meta, ok := m.metas[systemName]; ok {
		meta.Installed = true
		meta.InstalledVersion = p.Meta().Version
	}
	m.mu.Unlock()

	return nil
}

// Uninstall 鍗歌浇鎻掍欢
//
// 涓枃璇存槑:
// - 璋冪敤鎻掍欢鐨?Uninstall 鏂规硶;
// - 鐢ㄤ簬娓呯悊鏁版嵁銆佸垹闄よ〃(璋ㄦ厧);
// - 鍗歌浇鍚庝粠绠＄悊鍣ㄧЩ闄?
// - 娉ㄦ剰: 涓嶄粠 Container 绉婚櫎宸茬粦瀹氱殑鏈嶅姟銆?
func (m *Manager) Uninstall(ctx context.Context, systemName string) error {
	m.mu.RLock()
	p, ok := m.plugins[systemName]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("plugin %s not registered", systemName)
	}

	// 璋冪敤鎻掍欢鍗歌浇鏂规硶
	if err := p.Uninstall(ctx, m.container); err != nil {
		return fmt.Errorf("uninstall plugin %s: %w", systemName, err)
	}

	// 浠庣鐞嗗櫒绉婚櫎
	m.mu.Lock()
	delete(m.plugins, systemName)
	if meta, ok := m.metas[systemName]; ok {
		meta.Installed = false
		meta.InstalledVersion = ""
	}
	m.mu.Unlock()

	return nil
}

// Boot 鍚姩鎵€鏈夊凡瀹夎鐨勬彃浠?
//
// 涓枃璇存槑:
// - 鎸変緷璧栭『搴忎緷娆¤皟鐢ㄦ彃浠剁殑 Boot 鏂规硶;
// - 鐢ㄤ簬鍒濆鍖栬繍琛屾椂鐘舵€併€佽鍙栭厤缃?
// - 鍦ㄦ湇鍔″惎鍔ㄦ椂璋冪敤;
// - 濡傛灉鏌愪釜鎻掍欢 Boot 澶辫触,浼氳繑鍥為敊璇€?
func (m *Manager) Boot(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 瑙ｆ瀽渚濊禆椤哄簭
	order, err := m.resolveDependencyOrder()
	if err != nil {
		return fmt.Errorf("resolve dependency order: %w", err)
	}

	// 鎸夐『搴忓惎鍔?
	for _, systemName := range order {
		p, ok := m.plugins[systemName]
		if !ok {
			continue // 鏈敞鍐?璺宠繃
		}

		meta := p.Meta()
		if meta == nil || !meta.Installed {
			continue // 鏈畨瑁?璺宠繃
		}

		if err := p.Boot(ctx, m.container); err != nil {
			return fmt.Errorf("boot plugin %s: %w", systemName, err)
		}
	}

	return nil
}

// GetPlugin 鑾峰彇鎻掍欢瀹炰緥
//
// 涓枃璇存槑:
// - 閫氳繃 systemName 鑾峰彇宸叉敞鍐岀殑鎻掍欢;
// - 鐢ㄤ簬涓氬姟浠ｇ爜璋冪敤鎻掍欢鏂规硶銆?
func (m *Manager) GetPlugin(systemName string) (Plugin, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.plugins[systemName]
	return p, ok
}

// GetPaymentMethod 鑾峰彇鏀粯鎻掍欢
//
// 涓枃璇存槑:
// - 绫诲瀷瀹夊叏鐨勮幏鍙栨敮浠樻彃浠?
// - 杩斿洖 PaymentMethod 鎺ュ彛;
// - 濡傛灉鎻掍欢涓嶆槸鏀粯绫诲瀷,杩斿洖 false銆?
func (m *Manager) GetPaymentMethod(systemName string) (PaymentMethod, bool) {
	p, ok := m.GetPlugin(systemName)
	if !ok {
		return nil, false
	}

	pm, ok := p.(PaymentMethod)
	return pm, ok
}

// GetShippingMethod 鑾峰彇閰嶉€佹彃浠?
func (m *Manager) GetShippingMethod(systemName string) (ShippingMethod, bool) {
	p, ok := m.GetPlugin(systemName)
	if !ok {
		return nil, false
	}

	sm, ok := p.(ShippingMethod)
	return sm, ok
}

// GetWidgetPlugin 鑾峰彇灏忛儴浠舵彃浠?
func (m *Manager) GetWidgetPlugin(systemName string) (WidgetPlugin, bool) {
	p, ok := m.GetPlugin(systemName)
	if !ok {
		return nil, false
	}

	wp, ok := p.(WidgetPlugin)
	return wp, ok
}

// ListPlugins 鍒楀嚭鎵€鏈夊凡鍙戠幇鐨勬彃浠跺厓鏁版嵁
//
// 涓枃璇存槑:
// - 鍖呭惈宸叉敞鍐屽拰鏈敞鍐岀殑;
// - 鎸?DisplayOrder 鎺掑簭;
// - 鐢ㄤ簬绠＄悊鍚庡彴灞曠ず鎻掍欢鍒楄〃銆?
func (m *Manager) ListPlugins() []*PluginMeta {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*PluginMeta, 0, len(m.metas))
	for _, meta := range m.metas {
		result = append(result, meta)
	}

	// 鎸?DisplayOrder 鎺掑簭
	sort.Slice(result, func(i, j int) bool {
		return result[i].DisplayOrder < result[j].DisplayOrder
	})

	return result
}

// ListInstalledPlugins 鍒楀嚭宸插畨瑁呯殑鎻掍欢
func (m *Manager) ListInstalledPlugins() []*PluginMeta {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*PluginMeta, 0)
	for _, meta := range m.metas {
		if meta.Installed {
			result = append(result, meta)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].DisplayOrder < result[j].DisplayOrder
	})

	return result
}

// ListPluginsByType 鎸夌被鍨嬪垪鍑烘彃浠?
//
// 涓枃璇存槑:
// - pluginType 渚嬪: "payment", "shipping";
// - 鐢ㄤ簬绠＄悊鍚庡彴鎸夊垎绫诲睍绀恒€?
func (m *Manager) ListPluginsByType(pluginType string) []*PluginMeta {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*PluginMeta, 0)
	for systemName, meta := range m.metas {
		p, ok := m.plugins[systemName]
		if ok && p.PluginType() == pluginType {
			result = append(result, meta)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].DisplayOrder < result[j].DisplayOrder
	})

	return result
}

// resolveDependencyOrder 瑙ｆ瀽渚濊禆椤哄簭(鎷撴墤鎺掑簭)
//
// 涓枃璇存槑:
// - 鏍规嵁 meta.DependsOn 鏋勫缓渚濊禆鍥?
// - 鎷撴墤鎺掑簭纭繚渚濊禆鍏堝姞杞?
// - 妫€娴嬪惊鐜緷璧栧苟杩斿洖閿欒銆?
func (m *Manager) resolveDependencyOrder() ([]string, error) {
	// 鏋勫缓渚濊禆鍥?
	graph := make(map[string][]string)
	for systemName, meta := range m.metas {
		graph[systemName] = meta.DependsOn
	}

	// 鎷撴墤鎺掑簭
	var result []string
	visited := make(map[string]bool)
	visiting := make(map[string]bool) // 鐢ㄤ簬妫€娴嬪惊鐜?

	var visit func(string) error
	visit = func(node string) error {
		if visited[node] {
			return nil // 宸茶闂?
		}
		if visiting[node] {
			return fmt.Errorf("circular dependency detected at %s", node)
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

	// 閬嶅巻鎵€鏈夎妭鐐?
	for node := range graph {
		if err := visit(node); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// SetInstalled 鏇存柊鎻掍欢瀹夎鐘舵€?
//
// 涓枃璇存槑:
// - 鐢ㄤ簬浠庢暟鎹簱鎭㈠瀹夎鐘舵€?
// - 鍚姩鏃惰皟鐢?鎭㈠涓婃淇濆瓨鐨勭姸鎬併€?
func (m *Manager) SetInstalled(systemName string, installed bool, version string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if meta, ok := m.metas[systemName]; ok {
		meta.Installed = installed
		meta.InstalledVersion = version
	}
}
