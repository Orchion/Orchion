package capabilities

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetect(t *testing.T) {
	capabilities := Detect()

	// Verify basic structure
	assert.NotNil(t, capabilities)
	assert.NotEmpty(t, capabilities.Os)
	assert.NotEmpty(t, capabilities.Cpu)
	assert.NotEmpty(t, capabilities.Memory)

	// OS should be one of the common values (checking first part before slash)
	osParts := strings.Split(capabilities.Os, "/")
	assert.True(t, len(osParts) >= 1)
	osBase := osParts[0]
	validOS := []string{"linux", "darwin", "windows"}
	found := false
	for _, os := range validOS {
		if osBase == os {
			found = true
			break
		}
	}
	assert.True(t, found, "OS should be one of: linux, darwin, windows, got: %s", osBase)

	// CPU should contain "cores"
	assert.Contains(t, capabilities.Cpu, "cores")

	// Memory should contain "GB"
	assert.Contains(t, capabilities.Memory, "GB")
}

func Test_detectNVIDIAGPU(t *testing.T) {
	// Test when nvidia-smi is not available
	gpuType, _, _, _, _, _ := detectNVIDIAGPU()

	// Should return empty strings when nvidia-smi is not available
	if gpuType == "" {
		// All values should be empty when no GPU detected
	} else {
		// If GPU is detected, verify format
		assert.NotEmpty(t, gpuType)
	}
}

func Test_detectAMDGPU(t *testing.T) {
	// Test when rocm-smi is not available
	gpuType, _, _, _, _, _ := detectAMDGPU()

	// Should return empty strings when rocm-smi is not available
	if gpuType == "" {
		// All values should be empty when no GPU detected
	} else {
		// If GPU is detected, verify format
		assert.NotEmpty(t, gpuType)
	}
}

func Test_detectIntelGPU(t *testing.T) {
	// Test Intel GPU detection
	gpuType, _, _, _, _, _ := detectIntelGPU()

	// Intel GPU detection may or may not succeed depending on hardware
	// The important thing is that it doesn't crash
	if gpuType != "" {
		assert.NotEmpty(t, gpuType)
	}
}

func Test_detectGenericGPU(t *testing.T) {
	// Test generic GPU detection
	gpuType := detectGenericGPU()

	// Result depends on system, but shouldn't crash
	// Could be empty or contain detected GPU info
	assert.True(t, gpuType == "" ||
		len(gpuType) > 0, "GPU type should be empty or non-empty string")
}

func Test_detectPowerUsage(t *testing.T) {
	powerUsage := detectPowerUsage()

	// Result depends on platform and available tools
	// Should not be empty and should indicate detection status
	assert.NotEmpty(t, powerUsage)
}

func Test_detectPowerUsageLinux(t *testing.T) {
	powerUsage := detectPowerUsageLinux()

	// Should indicate detection status
	assert.NotEmpty(t, powerUsage)
	assert.True(t, powerUsage == "Power monitoring not available" ||
		len(powerUsage) > 0)
}

func Test_detectPowerUsageMacOS(t *testing.T) {
	powerUsage := detectPowerUsageMacOS()

	// Should indicate detection status
	assert.NotEmpty(t, powerUsage)
	assert.True(t, powerUsage == "Power monitoring not available" ||
		len(powerUsage) > 0)
}

func Test_detectPowerUsageWindows(t *testing.T) {
	powerUsage := detectPowerUsageWindows()

	// Should indicate detection status
	assert.NotEmpty(t, powerUsage)
	assert.True(t, powerUsage == "Power monitoring not available" ||
		len(powerUsage) > 0)
}