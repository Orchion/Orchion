package capabilities

import (
	"runtime"
	"strconv"

	"github.com/shirou/gopsutil/v3/mem"

	pb "github.com/Orchion/Orchion/node-agent/internal/proto/v1"
)

// Detect returns the system capabilities
func Detect() *pb.Capabilities {
	// Get actual system memory
	var memoryStr string
	if v, err := mem.VirtualMemory(); err == nil {
		totalMemGB := float64(v.Total) / (1024 * 1024 * 1024)
		memoryStr = strconv.FormatFloat(totalMemGB, 'f', 2, 64) + " GB"
	} else {
		// Fallback to Go runtime memory if system call fails
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		totalMemGB := float64(memStats.Sys) / (1024 * 1024 * 1024)
		memoryStr = strconv.FormatFloat(totalMemGB, 'f', 2, 64) + " GB (approximate)"
	}

	return &pb.Capabilities{
		Cpu:    strconv.Itoa(runtime.NumCPU()) + " cores",
		Memory: memoryStr,
		Os:     runtime.GOOS + "/" + runtime.GOARCH,
	}
}