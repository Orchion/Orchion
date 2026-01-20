package main

import (
	"fmt"
	"log"

	"github.com/Orchion/Orchion/node-agent/internal/capabilities"
)

func main() {
	log.Println("Testing capability detection...")
	caps := capabilities.Detect()

	fmt.Printf("Capabilities detected:\n")
	fmt.Printf("  CPU: %s\n", caps.Cpu)
	fmt.Printf("  Memory: %s\n", caps.Memory)
	fmt.Printf("  OS: %s\n", caps.Os)
	fmt.Printf("  GPU Type: %s\n", caps.GpuType)
	fmt.Printf("  GPU VRAM Total: %s\n", caps.GpuVramTotal)
	fmt.Printf("  GPU VRAM Available: %s\n", caps.GpuVramAvailable)
	fmt.Printf("  GPU VRAM Used: %s\n", caps.GpuVramUsed)
	fmt.Printf("  GPU Temperature: %s\n", caps.GpuTemperature)
	fmt.Printf("  GPU Power Usage: %s\n", caps.GpuPowerUsage)
	fmt.Printf("  System Power Usage: %s\n", caps.PowerUsage)
}
