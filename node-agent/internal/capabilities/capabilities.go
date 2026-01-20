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
	gpuType, gpuVramTotal, gpuVramAvailable, gpuVramUsed, gpuTemperature, gpuPowerUsage := detectGPU()

	// Detect system power usage (deprecated, but kept for backward compatibility)
	powerUsage := detectPowerUsage()

	return &pb.Capabilities{
		Cpu:              strconv.Itoa(runtime.NumCPU()) + " cores",
		Memory:           memoryStr,
		Os:               runtime.GOOS + "/" + runtime.GOARCH,
		GpuType:          gpuType,
		GpuVramTotal:     gpuVramTotal,
		GpuVramAvailable: gpuVramAvailable,
		GpuVramUsed:      gpuVramUsed,
		GpuTemperature:   gpuTemperature,
		GpuPowerUsage:    gpuPowerUsage,
		PowerUsage:       powerUsage,
	}
}

// detectGPU attempts to detect GPU information using system commands
func detectGPU() (gpuType, vramTotal, vramAvailable, vramUsed, temperature, powerUsage string) {
	// Try NVIDIA GPUs first
	if gpuType, vramTotal, vramAvailable, vramUsed, temperature, powerUsage := detectNVIDIAGPU(); gpuType != "" {
		return gpuType, vramTotal, vramAvailable, vramUsed, temperature, powerUsage
	}

	// Try AMD GPUs
	if gpuType, vramTotal, vramAvailable, vramUsed, temperature, powerUsage := detectAMDGPU(); gpuType != "" {
		return gpuType, vramTotal, vramAvailable, vramUsed, temperature, powerUsage
	}

	// Try Intel GPUs
	if gpuType, vramTotal, vramAvailable, vramUsed, temperature, powerUsage := detectIntelGPU(); gpuType != "" {
		return gpuType, vramTotal, vramAvailable, vramUsed, temperature, powerUsage
	}

	// Fallback: try to detect any GPU
	if gpuType := detectGenericGPU(); gpuType != "" {
		return gpuType, "Unknown", "Unknown", "Unknown", "Unknown", "Unknown"
	}

	return "No GPU detected", "N/A", "N/A", "N/A", "N/A", "N/A"
}

// detectNVIDIAGPU detects NVIDIA GPUs using nvidia-smi
func detectNVIDIAGPU() (gpuType, vramTotal, vramAvailable, vramUsed, temperature, powerUsage string) {
	// Check if nvidia-smi is available
	if _, err := exec.LookPath("nvidia-smi"); err != nil {
		return "", "", "", "", "", ""
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
		return "", "", "", "", "", ""
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

	// Get used VRAM (total - free)
	if vramTotal != "" && vramAvailable != "" {
		if totalMB, err := strconv.ParseFloat(strings.TrimSuffix(strings.TrimSpace(vramTotal), " GB"), 64); err == nil {
			if availMB, err := strconv.ParseFloat(strings.TrimSuffix(strings.TrimSpace(vramAvailable), " GB"), 64); err == nil {
				usedMB := totalMB - availMB
				vramUsed = fmt.Sprintf("%.1f GB", usedMB)
			}
		}
	}

	// Get GPU temperature
	if output, err := exec.Command("nvidia-smi", "--query-gpu=temperature.gpu", "--format=csv,noheader,nounits").Output(); err == nil {
		tempStr := strings.TrimSpace(string(output))
		tempStr = strings.TrimSuffix(tempStr, "\n")
		if strings.Contains(tempStr, ",") {
			tempStr = strings.Split(tempStr, ",")[0]
		}
		if temp, err := strconv.ParseFloat(strings.TrimSpace(tempStr), 64); err == nil {
			temperature = fmt.Sprintf("%.0f°C", temp)
		}
	}

	// Get GPU power usage
	if output, err := exec.Command("nvidia-smi", "--query-gpu=power.draw", "--format=csv,noheader,nounits").Output(); err == nil {
		powerStr := strings.TrimSpace(string(output))
		powerStr = strings.TrimSuffix(powerStr, "\n")
		if strings.Contains(powerStr, ",") {
			powerStr = strings.Split(powerStr, ",")[0]
		}
		if power, err := strconv.ParseFloat(strings.TrimSpace(powerStr), 64); err == nil {
			powerUsage = fmt.Sprintf("%.1f W", power)
		}
	}

	return gpuType, vramTotal, vramAvailable, vramUsed, temperature, powerUsage
}

// detectAMDGPU detects AMD GPUs using rocm-smi
func detectAMDGPU() (gpuType, vramTotal, vramAvailable, vramUsed, temperature, powerUsage string) {
	// Try rocm-smi for AMD GPUs
	if _, err := exec.LookPath("rocm-smi"); err != nil {
		return "", "", "", "", "", ""
	}

	// Get GPU name
	if output, err := exec.Command("rocm-smi", "--showproductname").Output(); err == nil {
		outputStr := string(output)
		// Parse output to extract GPU name
		lines := strings.Split(outputStr, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "GPU") && strings.Contains(line, "Radeon") {
				// Extract GPU name from line like "GPU[0] : Radeon RX 7900 XT"
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					gpuType = strings.TrimSpace(parts[1])
					break
				}
			}
		}
	}

	if gpuType == "" {
		return "", "", "", "", "", ""
	}

	// Get VRAM info using rocm-smi
	if output, err := exec.Command("rocm-smi", "--showmeminfo", "vram").Output(); err == nil {
		outputStr := string(output)
		lines := strings.Split(outputStr, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "VRAM Total Memory") {
				// Parse "VRAM Total Memory (GB): 16.0"
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					if totalGB, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64); err == nil {
						vramTotal = fmt.Sprintf("%.1f GB", totalGB)
					}
				}
			} else if strings.Contains(line, "VRAM Total Used Memory") {
				// Parse "VRAM Total Used Memory (GB): 2.1"
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					if usedGB, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64); err == nil {
						vramUsed = fmt.Sprintf("%.1f GB", usedGB)
						// Calculate available if we have total
						if vramTotal != "" {
							if totalGB, err := strconv.ParseFloat(strings.TrimSuffix(vramTotal, " GB"), 64); err == nil {
								availGB := totalGB - usedGB
								vramAvailable = fmt.Sprintf("%.1f GB", availGB)
							}
						}
					}
				}
			}
		}
	}

	// Get temperature
	if output, err := exec.Command("rocm-smi", "--showtemp").Output(); err == nil {
		outputStr := string(output)
		lines := strings.Split(outputStr, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "Temperature") {
				// Parse temperature line
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					tempStr := strings.TrimSpace(parts[1])
					// Remove unit if present
					tempStr = strings.TrimSuffix(tempStr, "c")
					tempStr = strings.TrimSuffix(tempStr, "C")
					if temp, err := strconv.ParseFloat(tempStr, 64); err == nil {
						temperature = fmt.Sprintf("%.0f°C", temp)
						break
					}
				}
			}
		}
	}

	// Get power usage
	if output, err := exec.Command("rocm-smi", "--showpower").Output(); err == nil {
		outputStr := string(output)
		lines := strings.Split(outputStr, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "Average Graphics Package Power") {
				// Parse "Average Graphics Package Power (W): 45.0"
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					if power, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64); err == nil {
						powerUsage = fmt.Sprintf("%.1f W", power)
						break
					}
				}
			}
		}
	}

	return gpuType, vramTotal, vramAvailable, vramUsed, temperature, powerUsage
}

// detectIntelGPU detects Intel GPUs using intel-gpu-top or other tools
func detectIntelGPU() (gpuType, vramTotal, vramAvailable, vramUsed, temperature, powerUsage string) {
	// Try intel-gpu-top for Intel GPUs
	if _, err := exec.LookPath("intel-gpu-top"); err != nil {
		return "", "", "", "", "", ""
	}

	// For Intel GPUs, VRAM detection is more complex as they often use shared system memory
	// This is a basic implementation - Intel GPU monitoring is less standardized

	// Get GPU name from lspci or similar
	if runtime.GOOS == "linux" {
		if output, err := exec.Command("lspci", "-v").Output(); err == nil {
			outputStr := string(output)
			if strings.Contains(strings.ToLower(outputStr), "intel") {
				lines := strings.Split(outputStr, "\n")
				for _, line := range lines {
					if strings.Contains(strings.ToLower(line), "intel") &&
					   (strings.Contains(strings.ToLower(line), "graphics") ||
					    strings.Contains(strings.ToLower(line), "display")) {
						// Extract GPU name
						parts := strings.Split(line, ": ")
						if len(parts) >= 2 {
							gpuType = strings.TrimSpace(parts[1])
							break
						}
					}
				}
			}
		}
	}

	if gpuType == "" {
		return "", "", "", "", "", ""
	}

	// Intel GPUs typically use shared system memory, so VRAM info is harder to get
	// For now, we'll mark as shared memory
	vramTotal = "Shared"
	vramAvailable = "Shared"
	vramUsed = "Shared"

	// Get temperature if available (limited support for Intel GPUs)
	if _, err := exec.Command("intel-gpu-top", "-l", "1").Output(); err == nil {
		// intel-gpu-top output parsing would be complex
		// For now, just indicate temperature monitoring is available
		temperature = "Available (intel-gpu-top)"
	}

	// Power usage for Intel GPUs is also limited
	powerUsage = "Limited support"

	return gpuType, vramTotal, vramAvailable, vramUsed, temperature, powerUsage
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
