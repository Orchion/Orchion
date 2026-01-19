package capabilities

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

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

	// Detect GPU information
	gpuType, gpuVramTotal, gpuVramAvailable := detectGPU()

	// Detect power usage
	powerUsage := detectPowerUsage()

	return &pb.Capabilities{
		Cpu:              strconv.Itoa(runtime.NumCPU()) + " cores",
		Memory:           memoryStr,
		Os:               runtime.GOOS + "/" + runtime.GOARCH,
		GpuType:          gpuType,
		GpuVramTotal:     gpuVramTotal,
		GpuVramAvailable: gpuVramAvailable,
		PowerUsage:       powerUsage,
	}
}

// detectGPU attempts to detect GPU information using system commands
func detectGPU() (gpuType, vramTotal, vramAvailable string) {
	// Try NVIDIA GPUs first
	if gpuType, vramTotal, vramAvailable := detectNVIDIAGPU(); gpuType != "" {
		return gpuType, vramTotal, vramAvailable
	}

	// Try AMD GPUs
	if gpuType, vramTotal, vramAvailable := detectAMDGPU(); gpuType != "" {
		return gpuType, vramTotal, vramAvailable
	}

	// Try Intel GPUs
	if gpuType, vramTotal, vramAvailable := detectIntelGPU(); gpuType != "" {
		return gpuType, vramTotal, vramAvailable
	}

	// Fallback: try to detect any GPU
	if gpuType := detectGenericGPU(); gpuType != "" {
		return gpuType, "Unknown", "Unknown"
	}

	return "No GPU detected", "N/A", "N/A"
}

// detectNVIDIAGPU detects NVIDIA GPUs using nvidia-smi
func detectNVIDIAGPU() (gpuType, vramTotal, vramAvailable string) {
	// Check if nvidia-smi is available
	if _, err := exec.LookPath("nvidia-smi"); err != nil {
		return "", "", ""
	}

	// Get GPU name
	if output, err := exec.Command("nvidia-smi", "--query-gpu=name", "--format=csv,noheader,nounits").Output(); err == nil {
		gpuType = strings.TrimSpace(string(output))
		// Remove trailing newline if present
		gpuType = strings.TrimSuffix(gpuType, "\n")
		if strings.Contains(gpuType, ",") {
			// Multiple GPUs, take the first one
			gpuType = strings.Split(gpuType, ",")[0]
		}
	} else {
		return "", "", ""
	}

	// Get total VRAM
	if output, err := exec.Command("nvidia-smi", "--query-gpu=memory.total", "--format=csv,noheader,nounits").Output(); err == nil {
		vramTotalStr := strings.TrimSpace(string(output))
		vramTotalStr = strings.TrimSuffix(vramTotalStr, "\n")
		if strings.Contains(vramTotalStr, ",") {
			vramTotalStr = strings.Split(vramTotalStr, ",")[0]
		}
		if vramMB, err := strconv.ParseFloat(strings.TrimSpace(vramTotalStr), 64); err == nil {
			vramTotal = fmt.Sprintf("%.1f GB", vramMB/1024)
		}
	}

	// Get available VRAM
	if output, err := exec.Command("nvidia-smi", "--query-gpu=memory.free", "--format=csv,noheader,nounits").Output(); err == nil {
		vramFreeStr := strings.TrimSpace(string(output))
		vramFreeStr = strings.TrimSuffix(vramFreeStr, "\n")
		if strings.Contains(vramFreeStr, ",") {
			vramFreeStr = strings.Split(vramFreeStr, ",")[0]
		}
		if vramMB, err := strconv.ParseFloat(strings.TrimSpace(vramFreeStr), 64); err == nil {
			vramAvailable = fmt.Sprintf("%.1f GB", vramMB/1024)
		}
	}

	return gpuType, vramTotal, vramAvailable
}

// detectAMDGPU detects AMD GPUs (placeholder - would need rocm-smi or similar)
func detectAMDGPU() (gpuType, vramTotal, vramAvailable string) {
	// TODO: Implement AMD GPU detection using rocm-smi or other tools
	return "", "", ""
}

// detectIntelGPU detects Intel GPUs (placeholder)
func detectIntelGPU() (gpuType, vramTotal, vramAvailable string) {
	// TODO: Implement Intel GPU detection
	return "", "", ""
}

// detectGenericGPU tries to detect any GPU using system tools
func detectGenericGPU() string {
	// Try lspci on Linux
	if runtime.GOOS == "linux" {
		if output, err := exec.Command("lspci", "-v").Output(); err == nil {
			outputStr := string(output)
			if strings.Contains(strings.ToLower(outputStr), "nvidia") {
				return "NVIDIA GPU (lspci)"
			}
			if strings.Contains(strings.ToLower(outputStr), "amd") || strings.Contains(strings.ToLower(outputStr), "radeon") {
				return "AMD GPU (lspci)"
			}
			if strings.Contains(strings.ToLower(outputStr), "intel") {
				return "Intel GPU (lspci)"
			}
		}
	}

	// Try system_profiler on macOS
	if runtime.GOOS == "darwin" {
		if output, err := exec.Command("system_profiler", "SPDisplaysDataType").Output(); err == nil {
			outputStr := string(output)
			if strings.Contains(strings.ToLower(outputStr), "nvidia") {
				return "NVIDIA GPU (macOS)"
			}
			if strings.Contains(strings.ToLower(outputStr), "amd") || strings.Contains(strings.ToLower(outputStr), "radeon") {
				return "AMD GPU (macOS)"
			}
			if strings.Contains(strings.ToLower(outputStr), "intel") {
				return "Intel GPU (macOS)"
			}
		}
	}

	return ""
}

// detectPowerUsage attempts to detect system power usage
func detectPowerUsage() string {
	// Try different power monitoring tools based on platform
	switch runtime.GOOS {
	case "linux":
		return detectPowerUsageLinux()
	case "darwin":
		return detectPowerUsageMacOS()
	case "windows":
		return detectPowerUsageWindows()
	default:
		return "Power monitoring not supported on " + runtime.GOOS
	}
}

// detectPowerUsageLinux tries various Linux power monitoring tools
func detectPowerUsageLinux() string {
	// Try powertop (if available)
	if _, err := exec.LookPath("powertop"); err == nil {
		// powertop can take time to gather data, so we'll skip for now
		// In a real implementation, you'd run it briefly and parse output
		return "Power monitoring available (powertop)"
	}

	// Try reading from /sys/class/power_supply/ (battery info)
	if output, err := exec.Command("cat", "/sys/class/power_supply/BAT*/power_now", "2>/dev/null", "||", "echo", "No battery").Output(); err == nil {
		powerStr := strings.TrimSpace(string(output))
		if powerStr != "No battery" && powerStr != "" {
			if powerW, err := strconv.ParseFloat(strings.TrimSpace(powerStr), 64); err == nil {
				return fmt.Sprintf("%.2f W (battery)", powerW/1000000) // Convert from microwatts
			}
		}
	}

	// Try upower
	if _, err := exec.LookPath("upower"); err == nil {
		if _, err := exec.Command("upower", "-e").Output(); err == nil {
			// upower is available but parsing would be complex
			return "Power monitoring available (upower)"
		}
	}

	return "Power monitoring not available"
}

// detectPowerUsageMacOS tries macOS power monitoring
func detectPowerUsageMacOS() string {
	// Try pmset
	if _, err := exec.LookPath("pmset"); err == nil {
		if output, err := exec.Command("pmset", "-g", "batt").Output(); err == nil {
			outputStr := string(output)
			if strings.Contains(outputStr, "discharging") || strings.Contains(outputStr, "charging") {
				return "Battery power monitoring available (pmset)"
			}
		}
		return "Power monitoring available (pmset)"
	}

	return "Power monitoring not available"
}

// detectPowerUsageWindows tries Windows power monitoring
func detectPowerUsageWindows() string {
	// Try powercfg
	if _, err := exec.LookPath("powercfg"); err == nil {
		return "Power monitoring available (powercfg)"
	}

	// Try WMIC
	if _, err := exec.LookPath("wmic"); err == nil {
		return "Power monitoring available (wmic)"
	}

	return "Power monitoring not available"
}
