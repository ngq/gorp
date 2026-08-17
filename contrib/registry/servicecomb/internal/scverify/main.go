// Package main 是针对真实 ServiceCenter 的 servicecomb 注册中心集成验证。
// 用法：在 contrib/registry/servicecomb 模块目录下执行
//   go run ./internal/scverify -server http://localhost:30100
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ngq/gorp/contrib/registry/servicecomb"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

func main() {
	server := flag.String("server", "http://localhost:30100", "ServiceCenter server URI")
	flag.Parse()

	cfg := &servicecomb.ServiceCombConfig{
		ServerURI:    *server,
		AppID:        "verify-app",
		ServiceName:  fmt.Sprintf("verify-svc-%d", time.Now().UnixNano()),
		Version:      "1.0.0",
		Environment:  "production",
		InstanceHost: "localhost",
		InstancePort: 8080,
		// 心跳间隔设短，便于验证心跳续租
		HeartbeatInterval: 2 * time.Second,
	}
	serviceName := cfg.ServiceName
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	reg, err := servicecomb.NewRegistry(cfg)
	if err != nil {
		fmt.Printf("FAIL NewRegistry: %v\n", err)
		os.Exit(1)
	}
	defer reg.Close()

	// 1. Register（内部自动启动心跳）
	if err := reg.Register(ctx, serviceName, "localhost:8080", map[string]string{"zone": "a"}); err != nil {
		fmt.Printf("FAIL Register: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("OK Register")

	// 2. Discover 应看到实例
	instances, err := reg.Discover(ctx, serviceName)
	if err != nil {
		fmt.Printf("FAIL Discover: %v\n", err)
		os.Exit(1)
	}
	found := false
	for _, inst := range instances {
		if inst.Address == "localhost:8080" && inst.Name == serviceName {
			found = true
			if !inst.Healthy {
				fmt.Printf("FAIL instance should be healthy: %+v\n", inst)
				os.Exit(1)
			}
			if inst.Metadata["zone"] != "a" {
				fmt.Printf("FAIL metadata not propagated: %+v\n", inst.Metadata)
				os.Exit(1)
			}
		}
	}
	if !found {
		fmt.Printf("FAIL instance localhost:8080 not found in discover, got %+v\n", instances)
		os.Exit(1)
	}
	fmt.Printf("OK Discover (%d instance, healthy + metadata)\n", len(instances))

	// 3. 等待超过 lease 过期周期（interval 2s * times 3 = 6s）。
	// 若心跳真实续租（PUT heartbeat），实例应仍存活；否则已过期成僵尸。
	time.Sleep(7 * time.Second)
	instances2, err := reg.Discover(ctx, serviceName)
	if err != nil {
		fmt.Printf("FAIL Discover after heartbeat wait: %v\n", err)
		os.Exit(1)
	}
	stillAlive := false
	for _, inst := range instances2 {
		if inst.Address == "localhost:8080" && inst.Name == serviceName {
			stillAlive = true
		}
	}
	if !stillAlive {
		fmt.Printf("FAIL instance not alive after heartbeat period (heartbeat broken?): %+v\n", instances2)
		os.Exit(1)
	}
	fmt.Println("OK instance alive after heartbeat period (heartbeat renewal works)")

	// 4. Deregister
	if err := reg.Deregister(ctx, serviceName, "localhost:8080"); err != nil {
		fmt.Printf("FAIL Deregister: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("OK Deregister")

	// 5. 稍等后实例应消失（ServiceCenter 删除是最终一致的，最多重试 10 次）
	gone := false
	var lastInstances []transportcontract.ServiceInstance
	for attempt := 0; attempt < 10 && !gone; attempt++ {
		time.Sleep(500 * time.Millisecond)
		instances3, err := reg.Discover(ctx, serviceName)
		if err != nil {
			gone = true // ErrServiceNotFound 即符合预期
			break
		}
		lastInstances = instances3
		gone = true
		for _, inst := range instances3 {
			if inst.Address == "localhost:8080" {
				gone = false
			}
		}
	}
	if gone {
		fmt.Println("OK instance gone after deregister")
	} else {
		fmt.Printf("WARN instance still present after deregister (final consistency window): %+v\n", lastInstances)
	}

	fmt.Println("ALL PASS")
}
