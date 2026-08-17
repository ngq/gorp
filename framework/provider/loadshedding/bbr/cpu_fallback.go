//go:build !linux

package bbr

// sampleSystemCPU 在非 Linux 平台不可用（无 /proc/stat）。
// 返回 ok=false，cpuMonitor 会退化为 goroutine 数代理指标。
// 生产部署建议运行在 Linux 上以获得真实 CPU 采样。
func sampleSystemCPU() (cpuSample, bool) {
	return cpuSample{}, false
}
