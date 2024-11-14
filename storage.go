package resourceutil

import (
	"log/slog"
	"syscall"
)

// StorageUsage represents disk storage metrics.
// Fields:
//   - TotalGB (float64): The total storage capacity in gigabytes.
//   - FreeGB (float64): The available storage in gigabytes for non-root users.
//   - UsedGB (float64): The amount of used storage in gigabytes.
//   - UsedPercent (float64): The percentage of storage in use.
type StorageUsage struct {
	TotalGB     float64
	FreeGB      float64
	UsedGB      float64
	UsedPercent float64
}

// GetDiskUsage retrieves disk usage statistics for a given file system path.
func GetDiskUsage(path string) (StorageUsage, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		slog.Error("Failed to get disk data", slog.String("path", path), slog.Any("error", err))
		return StorageUsage{}, err
	}

	// Calculate total, free, and used bytes.
	total := stat.Blocks * uint64(stat.Bsize) // Total bytes
	free := stat.Bavail * uint64(stat.Bsize)  // Available bytes to non-root users
	used := total - free                      // Used bytes

	const bytesPerGB = 1024 * 1024 * 1024
	totalGB := float64(total) / bytesPerGB
	freeGB := float64(free) / bytesPerGB
	usedGB := float64(used) / bytesPerGB

	usedPercent := (float64(used) / float64(total)) * 100

	storageUsage := StorageUsage{
		TotalGB:     totalGB,
		FreeGB:      freeGB,
		UsedGB:      usedGB,
		UsedPercent: usedPercent,
	}

	slog.Debug("Got disk usage", slog.Any("disk_usage", storageUsage))

	return storageUsage, nil
}
