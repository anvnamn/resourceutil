package resourceutil

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strconv"
)

// readMemInfo reads and returns the contents of /proc/meminfo.
func readMemInfo() (string, error) {
	b, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		slog.Error("Failed to read memory info", slog.String("path", "/proc/meminfo"), slog.Any("error", err))
		return "", err
	}
	return string(b), nil
}

// extractMemoryValue uses a regex to extract an integer memory value from /proc/meminfo based on the key.
func extractMemoryValue(memStr, key string) (int, error) {
	re := regexp.MustCompile(fmt.Sprintf(`%s:\s*(\d+) kB`, key))
	matches := re.FindStringSubmatch(memStr)

	if len(matches) < 2 {
		return 0, fmt.Errorf("could not find %s information in /proc/meminfo", key)
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("failed to parse %s value: %w", key, err)
	}

	return value, nil
}

// GetAvailableMem retrieves available memory in kilobytes from /proc/meminfo.
func GetAvailableMem() (int, error) {
	memStr, err := readMemInfo()
	if err != nil {
		return 0, err
	}

	memAvailable, err := extractMemoryValue(memStr, "MemAvailable")
	if err != nil {
		return 0, err
	}

	slog.Debug("Retrieved available memory", slog.Int("available_memory_kB", memAvailable))
	return memAvailable, nil
}

// GetTotalMem retrieves total memory in kilobytes from /proc/meminfo.
func GetTotalMem() (int, error) {
	memStr, err := readMemInfo()
	if err != nil {
		return 0, err
	}

	memTotal, err := extractMemoryValue(memStr, "MemTotal")
	if err != nil {
		return 0, err
	}

	slog.Debug("Retrieved total memory", slog.Int("total_memory_kB", memTotal))
	return memTotal, nil
}

// GetMemUsage calculates the current memory usage in percent.
func GetMemUsage() (float64, error) {
	totalMem, err := GetTotalMem()
	if err != nil {
		return 0, err
	}

	availableMem, err := GetAvailableMem()
	if err != nil {
		return 0, err
	}

	if totalMem == 0 {
		return 0, errors.New("divide by zero: total memory is zero")
	}

	memUsage := 100 * float64(totalMem-availableMem) / float64(totalMem)
	slog.Debug("Calculated memory usage", slog.Float64("usage_percent", memUsage))
	return memUsage, nil
}
