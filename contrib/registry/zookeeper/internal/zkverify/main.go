// Package main 是 ZooKeeper 注册中心集成验证，验证 REG-02 修复：
// 会话过期后 ephemeral 临时节点自动重建（否则服务从发现列表消失且永不恢复）。
//
// 编排方式：
//   go run ./internal/zkverify                    # 常驻：注册 + Watch + 等待恢复
//   （另开终端）docker network disconnect bridge zk
//   sleep 15
//   docker network connect bridge zk
//
// 用法：在 contrib/registry/zookeeper 模块目录下执行
//   go run ./internal/zkverify -servers localhost:2181 -wait 60
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/go-zookeeper/zk"

	"github.com/ngq/gorp/contrib/registry/zookeeper"
)

// zkConnect 建立到 ZK 的原始连接。
func zkConnect(servers []string, timeout time.Duration) (*zk.Conn, <-chan zk.Event, error) {
	conn, events, err := zk.Connect(servers, timeout)
	if err != nil {
		return nil, nil, err
	}
	// 等待连接就绪（最多 3 次事件循环）
	for i := 0; i < 10; i++ {
		select {
		case ev := <-events:
			if ev.State == zk.StateConnected || ev.State == zk.StateHasSession {
				return conn, events, nil
			}
		case <-time.After(300 * time.Millisecond):
			if conn.State() == zk.StateHasSession {
				return conn, events, nil
			}
		}
	}
	return conn, events, nil
}

func main() {
	servers := flag.String("servers", "localhost:2181", "zookeeper servers")
	wait := flag.Int("wait", 70, "total seconds to run (window for external network cut)")
	flag.Parse()

	cfg := &zookeeper.ZookeeperConfig{
		Servers:        []string{*servers},
		SessionTimeout: 5 * time.Second,
		BasePath:       "/gorp-verify",
		ServiceName:    "verify-svc",
		ServiceMeta:    map[string]string{"zone": "a"},
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*wait)*time.Second)
	defer cancel()

	reg, err := zookeeper.NewRegistry(cfg)
	if err != nil {
		fmt.Printf("FAIL NewRegistry: %v\n", err)
		os.Exit(1)
	}
	defer reg.Close()

	// 1. 注册 ephemeral 临时节点
	if err := reg.Register(ctx, "verify-svc", "localhost:8080", nil); err != nil {
		fmt.Printf("FAIL Register: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("OK registered ephemeral node")

	// 2. 确认 Discover 可见
	instances, err := reg.Discover(ctx, "verify-svc")
	if err != nil || len(instances) == 0 {
		fmt.Printf("FAIL discover after register: %v / %+v\n", err, instances)
		os.Exit(1)
	}
	fmt.Printf("OK discovered %d instance\n", len(instances))

	// 3. 启动 Watch：会话过期时，watch loop 会检测到并重建临时节点。
	watchCh, err := reg.Watch(ctx, "verify-svc")
	if err != nil {
		fmt.Printf("FAIL Watch: %v\n", err)
		os.Exit(1)
	}

	// 4. 等待外部把网络切断（docker network disconnect bridge zk），
	//    会话过期后 go-zookeeper 重连（新 session），ephemeral 节点被服务端删除。
	//    本进程的 Watch 循环应检测到 ErrSessionExpired 并重建节点。
	fmt.Printf(">>> now cut the network for > %s (docker network disconnect bridge zk), I will re-check\n", cfg.SessionTimeout)

	// 直接连 ZK 检查节点存在性（绕过 registry 的 endpointCache，
	// 缓存会掩盖节点消失）。用独立的 zk.Conn 保证检查的是服务端真实状态。
	rawConn, _, err := zkConnect([]string{*servers}, 5*time.Second)
	if err != nil {
		fmt.Printf("FAIL direct zk connect: %v\n", err)
		os.Exit(1)
	}
	defer rawConn.Close()
	nodePath := "/gorp-verify/verify-svc/localhost:8080"

	fmt.Printf(">>> now cut the network for > %s (docker network disconnect bridge zk), I will re-check\n", cfg.SessionTimeout)

	// 记录节点当前的 ephemeralOwner（创建者 session id）。
	_, oldStat, err := rawConn.Exists(nodePath)
	if err != nil || oldStat == nil {
		fmt.Printf("FAIL cannot stat node: %v\n", err)
		os.Exit(1)
	}
	oldOwner := oldStat.EphemeralOwner
	fmt.Printf("initial ephemeralOwner=%d\n", oldOwner)

	// 会话过期后，ZK 会删除旧 session 的 ephemeral 节点；若节点仍存在且
	// ephemeralOwner 变为新 session id，则证明是 REG-02 的重建逻辑新建的
	// （ephemeral 节点不可能跨 session 存活）。
	sawNewOwner := false
	deadline := time.Now().Add(time.Duration(*wait) * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(2 * time.Second)
		exists, stat, eErr := rawConn.Exists(nodePath)
		if eErr != nil {
			continue // 网络切断期间连接断，忽略
		}
		if exists && stat != nil && stat.EphemeralOwner != 0 && stat.EphemeralOwner != oldOwner {
			sawNewOwner = true
			fmt.Printf("OK ephemeral node recreated: ephemeralOwner changed %d -> %d (new session owns it)\n", oldOwner, stat.EphemeralOwner)
			break
		}
	}
	// 消费 watch 事件避免阻塞（不强制）
	select {
	case <-watchCh:
	default:
	}

	if sawNewOwner {
		fmt.Println("ALL PASS (ephemeral node recreated under new session after session expiry)")
	} else {
		fmt.Println("FAIL ephemeral node NOT recreated under a new session (or session never expired)")
		os.Exit(1)
	}
}
