package cdtools

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// ScanDevices returns CD/DVD drive device paths via macOS diskutil.
// The drive must have a disc inserted to appear in diskutil list.
func ScanDevices() ([]string, error) {
	out, err := exec.Command("diskutil", "list").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("diskutil list: %w", err)
	}

	var devices []string
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "/dev/disk") {
			continue
		}
		disk := strings.Fields(line)[0]
		if isOpticalDisk(disk) {
			// cdrdao uses raw device nodes (/dev/rdiskN) for direct sector access
			devices = append(devices, strings.Replace(disk, "/dev/disk", "/dev/rdisk", 1))
		}
	}
	return devices, nil
}

func isOpticalDisk(disk string) bool {
	out, _ := exec.Command("diskutil", "info", disk).CombinedOutput()
	return bytes.Contains(out, []byte("Optical Drive Type"))
}

// DetectDevice returns the first optical drive device found, or an error if none is detected.
func DetectDevice() (string, error) {
	devices, err := ScanDevices()
	if err != nil {
		return "", err
	}
	if len(devices) == 0 {
		return "", fmt.Errorf("no CD/DVD drives found — is the drive connected and a disc inserted?")
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

// Unmount unmounts the disc without ejecting it so cdrdao can access it.
// Accepts /dev/diskN or /dev/rdiskN.
func Unmount(device string) error {
	disk := strings.Replace(device, "/dev/rdisk", "/dev/disk", 1)
	out, err := exec.Command("diskutil", "unmount", disk).CombinedOutput()
	if err != nil {
		// "not mounted" is fine — blank discs and already-unmounted discs land here
		if bytes.Contains(out, []byte("not mounted")) {
			return nil
		}
		return fmt.Errorf("diskutil unmount %s: %w", disk, err)
	}
	return nil
}

// ScanCdrdao returns the cdrdao device address for the first optical drive
// found via "cdrdao scanbus". Must be called after Unmount — cdrdao scanbus
// fails when the disc is still mounted.
func ScanCdrdao() (string, error) {
	out, err := exec.Command("cdrdao", "scanbus").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("cdrdao scanbus: %w\n%s", err, bytes.TrimSpace(out))
	}
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		if idx := strings.Index(line, " : "); idx > 0 {
			return strings.TrimSpace(line[:idx]), nil
		}
	}
	return "", fmt.Errorf("no optical drives found via cdrdao scanbus")
}

// Eject ejects the disc in the given device.
func Eject(device string) error {
	return exec.Command("diskutil", "eject", device).Run()
}
