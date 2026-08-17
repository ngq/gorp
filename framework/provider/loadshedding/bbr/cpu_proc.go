//go:build linux

package bbr

import (
	"os"
	"strconv"
	"strings"
)

// sampleSystemCPU 读取 /proc/stat 的聚合 cpu 行，返回真实的系统 CPU 计数器。
// 任何读取/解析失败都返回 ok=false，调用方回退到代理指标。
func sampleSystemCPU() (cpuSample, bool) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return cpuSample{}, false
	}
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "cpu ") {
			continue
		}
		// 格式：cpu user nice system idle iowait irq softirq steal guest guest_nice
		fields := strings.Fields(line)[1:]
		if len(fields) < 5 {
			return cpuSample{}, false
		}
		var vals [8]uint64
		for i := 0; i < len(fields) && i < len(vals); i++ {
			v, err := strconv.ParseUint(fields[i], 10, 64)
			if err != nil {
				return cpuSample{}, false
			}
			vals[i] = v
		}
		var total uint64
		for i := 0; i < len(vals); i++ {
			total += vals[i]
		}
		idle := vals[3] + vals[4] // idle + iowait 计为非活跃
		return cpuSample{user: vals[0], system: vals[2], idle: idle, total: total}, true
	}
	return cpuSample{}, false
}
