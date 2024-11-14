package resourceutil

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// intFromFile reads a file at the specified path and attempts to parse its contents as an integer.
func intFromFile(path string) (int, error) {
	// Read the data
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("failed to read int at path %s, error: %w", path, err)
	}

	dataStr := strings.TrimSpace(string(data))
	dataInt, err := strconv.Atoi(dataStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse int at path %s, error: %w", path, err)
	}

	return dataInt, nil
}

// GetBatterySOC retrieves the State of Charge (SOC) of the battery as a percentage.
func GetBatterySOC(batteryName string) (int, error) {
	if batteryName == "" {
		return 0, fmt.Errorf("battery name cannot be empty")
	}

	path := fmt.Sprintf("/sys/class/power_supply/%s/capacity", batteryName)
	capacity, err := intFromFile(path)
	if err != nil {
		return 0, fmt.Errorf("failed to get battery SOC for %s: %w", batteryName, err)
	}

	return capacity, nil
}

// GetBatterySOH retrieves the State of Health (SOH) of the battery as a percentage.
//
// SOH is calculated as the ratio of the battery's current maximum energy capacity
// as a percentage of its original design capacity.
func GetBatterySOH(batteryName string) (int, error) {
	if batteryName == "" {
		return 0, fmt.Errorf("battery name cannot be empty")
	}

	energyFullPath := fmt.Sprintf("/sys/class/power_supply/%s/energy_full", batteryName)
	energyFullDesignPath := fmt.Sprintf("/sys/class/power_supply/%s/energy_full_design", batteryName)

	energyFull, err := intFromFile(energyFullPath)
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve energy_full for battery %s: %w", batteryName, err)
	}

	energyFullDesign, err := intFromFile(energyFullDesignPath)
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve energy_full_design for battery %s: %w", batteryName, err)
	}

	if energyFullDesign == 0 {
		return 0, fmt.Errorf("energy_full_design is zero, cannot calculate SOH for battery %s", batteryName)
	}

	stateOfHealth := 100 * energyFull / energyFullDesign
	return stateOfHealth, nil
}
