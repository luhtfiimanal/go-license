package licverify

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// HardwareInfo contains information about the current hardware
type HardwareInfo struct {
	MACAddresses []string
	DiskIDs      []string
	Hostname     string
	CPUInfo      string
}

// GetHardwareInfo collects hardware information from the system
func GetHardwareInfo() (*HardwareInfo, error) {
	macs, err := getMACAddresses()
	if err != nil {
		return nil, fmt.Errorf("failed to get MAC addresses: %v", err)
	}

	diskIDs, err := getDiskIDs()
	if err != nil {
		return nil, fmt.Errorf("failed to get disk IDs: %v", err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %v", err)
	}

	cpuInfo, err := getCPUInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU info: %v", err)
	}

	return &HardwareInfo{
		MACAddresses: macs,
		DiskIDs:      diskIDs,
		Hostname:     hostname,
		CPUInfo:      cpuInfo,
	}, nil
}

// getMACAddresses returns all non-loopback MAC addresses
func getMACAddresses() ([]string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var macAddresses []string
	for _, iface := range interfaces {
		// Only include up, non-loopback interfaces with a valid hardware address
		if iface.Flags&net.FlagUp != 0 &&
			iface.Flags&net.FlagLoopback == 0 &&
			len(iface.HardwareAddr.String()) > 0 {
			macAddresses = append(macAddresses, iface.HardwareAddr.String())
		}
	}

	if len(macAddresses) == 0 {
		return nil, errors.New("no valid MAC addresses found")
	}

	return macAddresses, nil
}

// getDiskIDs returns disk identifiers based on the current OS
func getDiskIDs() ([]string, error) {
	switch runtime.GOOS {
	case "linux":
		return getLinuxDiskIDs()
	case "windows":
		return getWindowsDiskIDs()
	case "darwin":
		return getMacOSDiskIDs()
	default:
		return []string{"unknown-os-disk-id"}, nil
	}
}

// getLinuxDiskIDs gets disk serial numbers on Linux
func getLinuxDiskIDs() ([]string, error) {
	// Try to get disk serial numbers using lsblk
	cmd := exec.Command("lsblk", "-no", "SERIAL", "-d")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()

	if err != nil {
		// Fallback to disk by-id
		cmd = exec.Command("ls", "-la", "/dev/disk/by-id/")
		cmd.Stdout = &out
		err = cmd.Run()
		if err != nil {
			return []string{"linux-disk-id-fallback"}, nil
		}
	}

	lines := strings.Split(out.String(), "\n")
	var diskIDs []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			diskIDs = append(diskIDs, line)
		}
	}

	if len(diskIDs) == 0 {
		return []string{"linux-disk-id-fallback"}, nil
	}

	return diskIDs, nil
}

// getWindowsDiskIDs gets disk serial numbers on Windows
func getWindowsDiskIDs() ([]string, error) {
	// Use wmic to get disk serial numbers
	cmd := exec.Command("wmic", "diskdrive", "get", "SerialNumber")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()

	if err != nil {
		return []string{"windows-disk-id-fallback"}, nil
	}

	lines := strings.Split(out.String(), "\n")
	var diskIDs []string

	// Skip the first line which is the header
	for i, line := range lines {
		if i == 0 {
			continue
		}

		line = strings.TrimSpace(line)
		if line != "" {
			diskIDs = append(diskIDs, line)
		}
	}

	if len(diskIDs) == 0 {
		return []string{"windows-disk-id-fallback"}, nil
	}

	return diskIDs, nil
}

// getMacOSDiskIDs gets disk serial numbers on macOS
func getMacOSDiskIDs() ([]string, error) {
	// Use diskutil to get disk info
	cmd := exec.Command("diskutil", "info", "/dev/disk0")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()

	if err != nil {
		return []string{"macos-disk-id-fallback"}, nil
	}

	// Parse output to find serial number
	lines := strings.Split(out.String(), "\n")
	var diskIDs []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Serial Number") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				serial := strings.TrimSpace(parts[1])
				diskIDs = append(diskIDs, serial)
			}
		}
	}

	if len(diskIDs) == 0 {
		return []string{"macos-disk-id-fallback"}, nil
	}

	return diskIDs, nil
}

// getCPUInfo gets CPU information
func getCPUInfo() (string, error) {
	switch runtime.GOOS {
	case "linux":
		return getLinuxCPUInfo()
	case "windows":
		return getWindowsCPUInfo()
	case "darwin":
		return getMacOSCPUInfo()
	default:
		return runtime.GOARCH + "-unknown-cpu", nil
	}
}

// getLinuxCPUInfo gets CPU information on Linux
func getLinuxCPUInfo() (string, error) {
	// Read CPU info from /proc/cpuinfo
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return "linux-cpu-fallback", nil
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "model name") || strings.HasPrefix(line, "processor") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				return strings.TrimSpace(parts[1]), nil
			}
		}
	}

	return "linux-cpu-fallback", nil
}

// getWindowsCPUInfo gets CPU information on Windows
func getWindowsCPUInfo() (string, error) {
	cmd := exec.Command("wmic", "cpu", "get", "ProcessorId")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()

	if err != nil {
		return "windows-cpu-fallback", nil
	}

	lines := strings.Split(out.String(), "\n")
	if len(lines) >= 2 {
		return strings.TrimSpace(lines[1]), nil
	}

	return "windows-cpu-fallback", nil
}

// getMacOSCPUInfo gets CPU information on macOS
func getMacOSCPUInfo() (string, error) {
	cmd := exec.Command("sysctl", "-n", "machdep.cpu.brand_string")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()

	if err != nil {
		return "macos-cpu-fallback", nil
	}

	return strings.TrimSpace(out.String()), nil
}
