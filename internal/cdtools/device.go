package cdtools

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// ScanDevices returns all CD/DVD drive device paths found by cdrdao.
func ScanDevices() ([]string, error) {
	out, err := exec.Command("cdrdao", "scan-bus").CombinedOutput()
	if err != nil {
		// cdrdao returns non-zero even on success when printing device list;
		// only fail if output is empty.
		if len(out) == 0 {
			return nil, fmt.Errorf("cdrdao scan-bus failed: %w", err)
		}
	}

	var devices []string
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		// Lines with device paths look like: /dev/diskN: Vendor Model, Rev X.XX
		if strings.HasPrefix(line, "/dev/") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) >= 1 {
				devices = append(devices, strings.TrimSpace(parts[0]))
			}
		}
	}
	return devices, nil
}

// DetectDevice returns the first optical drive device found, or an error if
// none is detected.
func DetectDevice() (string, error) {
	devices, err := ScanDevices()
	if err != nil {
		return "", err
	}
	if len(devices) == 0 {
		return "", fmt.Errorf("no CD/DVD drives found — is a drive connected?")
	}
	return devices[0], nil
}

// CheckDependencies verifies that required tools are on PATH.
func CheckDependencies() error {
	if _, err := exec.LookPath("cdrdao"); err != nil {
		return fmt.Errorf("cdrdao not found — install it with: brew install cdrdao")
	}
	return nil
}

// Eject ejects the disc in the given device using diskutil.
func Eject(device string) error {
	return exec.Command("diskutil", "eject", device).Run()
}
