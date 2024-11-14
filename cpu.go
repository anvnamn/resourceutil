package resourceutil

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

// GetCPULoad calculates the current CPU load percentage by reading data from the /proc/stat file.
func GetCPULoad() (float64, error) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		slog.Error("Failed to read process info", slog.String("path", "/proc/stat"), slog.Any("error", err))
		return 0, err
	}
	defer file.Close()

	// Create a new scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	var loadPercent float64

	// Iterate over each line in /proc/stat
	for scanner.Scan() {
		line := scanner.Text()

		// Look for aggregate cpu statistics, which start with "cpu "
		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)
			slog.Debug("Found CPU line", slog.String("cpu_statistics", line))

			// The CPU statistics line in /proc/stat has these fields:
			// 0: Prefix ("cpu" for aggregate statistics or "cpuN" for individual cores)

			// Core CPU statistics fields
			// 1: User       - Time spent in user mode
			// 2: Nice       - Time spent in user mode with low priority (nice)
			// 3: System     - Time spent in system mode
			// 4: Idle       - Time spent in the idle task

			// Added in Linux 2.5.41:
			// 5: IOWait     - Time spent waiting for I/O to complete
			// 6: IRQ        - Time spent servicing hardware interrupts
			// 7: SoftIRQ    - Time spent servicing software interrupts

			// Added in Linux 2.6.11:
			// 8: Steal      - Time spent in other operating systems when running in a virtualized environment

			// Added in Linux 2.6.24:
			// 9: Guest      - Time spent running a virtual CPU for guest operating systems
			// 10: GuestNice - Time spent running a low-priority virtual CPU for guest operating systems

			if len(fields) > 11 || len(fields) < 5 {
				return 0, fmt.Errorf("unexpected number of cpu fields in /proc/stat, cpu line: %s", line)
			}

			totalTime := 0.0
			idleTime := 0.0

			for i := 1; i < len(fields); i++ {
				value, err := strconv.ParseFloat(fields[i], 64)
				if err != nil {
					return 0, fmt.Errorf("failed to parse CPU field %d: %w", i, err)
				}
				totalTime += value
				if i == 4 || i == 5 {
					idleTime += value
				}
			}

			loadPercent = 100 * (totalTime - idleTime) / totalTime
			slog.Debug("Calculated CPU load percent", slog.Float64("cpu_load_percent", loadPercent))
			break
		}
	}
	// Handle any scanning error
	if err := scanner.Err(); err != nil {
		slog.Error("Failed to read /proc/stat", slog.Any("error", err))
		return 0, err
	}

	return loadPercent, nil
}
