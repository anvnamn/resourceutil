package resourceutil

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	loadMeasurements      [10]float64
	loadMeasurementsMutex sync.Mutex
	isMeasuring           bool
	measurementMutex      sync.Mutex
)

// Starts the goroutine that measures the CPU load
func StartCPUMeasuring() {
	measurementMutex.Lock()
	if isMeasuring {
		slog.Warn("Unable to start CPU load measurement as it is already started")
		return
	}
	isMeasuring = true
	measurementMutex.Unlock()

	go func() {
		for {
			cpuLoad, err := doCPUMeasure()
			if err != nil {
				slog.Error("Failed to measure CPU load", slog.Any("error", err))
				continue
			}

			// Update the averaged CPU load safely
			loadMeasurementsMutex.Lock()
			for i := len(loadMeasurements) - 1; i > 0; i-- {
				loadMeasurements[i] = loadMeasurements[i-1]
			}
			loadMeasurements[0] = cpuLoad
			loadMeasurementsMutex.Unlock()

			slog.Debug("Added new measurement", slog.Float64("new_measurement", cpuLoad), slog.Any("measurement_array", loadMeasurements))
		}
	}()
}

// GetCPULoad retrieves the CPU load averaged over 1 second
// Throws an error if the measurement loop has not started.
func GetCPULoad() (float64, error) {
	if !isMeasuring {
		return 0.0, fmt.Errorf("CPU measurement loop has not started, start measurement before trying to read load")
	}

	avgLoad := 0.0

	measurementMutex.Lock()
	defer measurementMutex.Unlock()

	for i := 0; i < len(loadMeasurements); i++ {
		avgLoad += loadMeasurements[i]
	}

	avgLoad = avgLoad / float64(len(loadMeasurements))

	return avgLoad, nil
}

// Does one blocking measurement of CPU load over a period of 100 ms
func doCPUMeasure() (float64, error) {
	// Helper function to read CPU stats from /proc/stat
	readCPUStats := func() (totalTime, idleTime float64, err error) {
		file, err := os.Open("/proc/stat")
		if err != nil {
			slog.Error("Failed to read process info", slog.String("path", "/proc/stat"), slog.Any("error", err))
			return 0, 0, err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			line := scanner.Text()

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

				// Validate the number of fields
				if len(fields) > 11 || len(fields) < 5 {
					return 0, 0, fmt.Errorf("unexpected number of CPU fields in /proc/stat, cpu line: %s", line)
				}

				for i := 1; i < len(fields); i++ {
					value, err := strconv.ParseFloat(fields[i], 64)
					if err != nil {
						return 0, 0, fmt.Errorf("failed to parse CPU field %d: %w", i, err)
					}
					totalTime += value
					if i == 4 || i == 5 { // Idle and IOWait
						idleTime += value
					}
				}
				break
			}
		}

		// Handle scanning errors
		if err := scanner.Err(); err != nil {
			slog.Error("Failed to scan /proc/stat", slog.Any("error", err))
			return 0, 0, err
		}

		return totalTime, idleTime, nil
	}

	// Read the first snapshot
	totalTime1, idleTime1, err := readCPUStats()
	if err != nil {
		return 0, err
	}

	// 100 ms between two measurements
	time.Sleep(time.Millisecond * 100)

	// Read the second snapshot
	totalTime2, idleTime2, err := readCPUStats()
	if err != nil {
		return 0, err
	}

	// Calculate the differences
	totalDiff := totalTime2 - totalTime1
	idleDiff := idleTime2 - idleTime1

	// Calculate CPU load percentage
	if totalDiff == 0 {
		return 0, fmt.Errorf("no CPU activity detected during the interval")
	}

	cpuLoad := 100 * (totalDiff - idleDiff) / totalDiff
	slog.Debug("Calculated CPU load over duration", slog.Float64("cpu_load_percent", cpuLoad))

	return cpuLoad, nil
}
