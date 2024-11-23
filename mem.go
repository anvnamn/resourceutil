package resourceutil

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strconv"
)

// MemUsage represents memory usage metrics.
// Fields:
//   - TotalGB (float64): The total memory size in gigabytes.
//   - FreeGB (float64): The available memory in gigabytes.
//   - UsedGB (float64): The amount of used memory in gigabytes.
//   - UsedPercent (float64): The percentage of memory in use.
type MemUsage struct {
	TotalGB     float64
	AvailableGB float64
	UsedGB      float64
	UsedPercent float64
}

func GetMemUsage() (MemUsage, error) {
	memStr, err := readMemInfo()
	if err != nil {
		return MemUsage{}, err
	}

	memAvailable, err := extractMemValue(memStr, "MemAvailable")
	if err != nil {
		return MemUsage{}, err
	}
	slog.Debug("Retrieved available memory", slog.Float64("available_memory_GB", memAvailable))

	memTotal, err := extractMemValue(memStr, "MemTotal")
	if err != nil {
		return MemUsage{}, err
	}
	slog.Debug("Retrieved total memory", slog.Float64("total_memory_GB", memTotal))

	if memTotal == 0 {
		return MemUsage{}, errors.New("divide by zero: total memory is zero")
	}

	usagePercent := 100 * float64(memTotal-memAvailable) / float64(memTotal)
	slog.Debug("Calculated memory usage", slog.Float64("usage_percent", usagePercent))

	memUsage := MemUsage{
		TotalGB:     memTotal,
		AvailableGB: memAvailable,
		UsedGB:      memTotal - memAvailable,
		UsedPercent: usagePercent,
	}

	return memUsage, nil
}

// readMemInfo reads and returns the contents of /proc/meminfo.
func readMemInfo() (string, error) {
	b, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		slog.Error("Failed to read memory info", slog.String("path", "/proc/meminfo"), slog.Any("error", err))
		return "", err
	}
	slog.Debug("Read memory info", slog.String("mem_info", string(b)))
	return string(b), nil
}

// extractMemoryValue uses a regex to extract an integer value in kB from /proc/meminfo based on the key.
// Returns the memory value in GB.
func extractMemValue(memStr, key string) (float64, error) {
	re := regexp.MustCompile(fmt.Sprintf(`%s:\s*(\d+) kB`, key))
	matches := re.FindStringSubmatch(memStr)

	if len(matches) < 2 {
		return 0, fmt.Errorf("could not find %s information in /proc/meminfo", key)
	}

	valuekB, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("failed to parse %s value: %w", key, err)
	}

	valueGB := float64(valuekB) / (1024 * 1024)
	return valueGB, nil
}
