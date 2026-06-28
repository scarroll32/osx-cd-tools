package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/scarroll32/osx-cd-tools/internal/cdtools"
	"github.com/spf13/cobra"
)

var (
	flagDevice    string
	flagSpeed     int
	flagKeepImage bool
	flagImageDir  string
)

var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Copy an audio CD (read source, then write to blank)",
	Long: `Copies an audio CD using cdrdao for maximum quality.

Workflow:
  1. Insert the SOURCE disc and press Enter
  2. cd-tools reads a raw TOC+binary image of the disc
  3. Eject the source disc, insert a blank CD-R, press Enter
  4. cd-tools burns the image to the blank disc`,
	RunE: runCopy,
}

func init() {
	copyCmd.Flags().StringVarP(&flagDevice, "device", "d", "", "CD drive device path (auto-detected if omitted)")
	copyCmd.Flags().IntVarP(&flagSpeed, "speed", "s", 4, "Read/write speed multiplier (lower = higher quality, 1–8 recommended)")
	copyCmd.Flags().BoolVar(&flagKeepImage, "keep-image", false, "Keep the disc image files after copying")
	copyCmd.Flags().StringVar(&flagImageDir, "image-dir", "", "Directory to store the disc image (implies --keep-image, defaults to a temp dir)")
	rootCmd.AddCommand(copyCmd)
}

func runCopy(cmd *cobra.Command, args []string) error {
	if err := cdtools.CheckDependencies(); err != nil {
		return err
	}

	device := flagDevice
	if device == "" {
		fmt.Println("Scanning for CD drives...")
		var err error
		device, err = cdtools.DetectDevice()
		if err != nil {
			return fmt.Errorf("%w\n\nUse --device to specify the drive path (e.g. --device /dev/disk2)", err)
		}
		fmt.Printf("Found drive: %s\n", device)
	}

	// Resolve image directory.
	imageDir := flagImageDir
	if imageDir != "" {
		flagKeepImage = true
		if err := os.MkdirAll(imageDir, 0o755); err != nil {
			return fmt.Errorf("cannot create image directory: %w", err)
		}
	} else {
		var err error
		imageDir, err = os.MkdirTemp("", "cd-copy-*")
		if err != nil {
			return fmt.Errorf("cannot create temp directory: %w", err)
		}
		if !flagKeepImage {
			defer os.RemoveAll(imageDir)
		}
	}

	tocFile := filepath.Join(imageDir, "disc.toc")
	binFile := filepath.Join(imageDir, "disc.bin")

	// ── Step 1: Read source disc ──────────────────────────────────────────────
	fmt.Println("\n── Step 1 of 2: Read source disc ───────────────────────────────")
	fmt.Println("Insert the SOURCE audio CD into the drive, then press Enter...")
	waitForEnter()

	fmt.Printf("Reading disc at %dx speed (raw mode for maximum quality)...\n", flagSpeed)
	if err := readDisc(device, tocFile, binFile, flagSpeed); err != nil {
		return fmt.Errorf("disc read failed: %w", err)
	}
	fmt.Println("Source disc read successfully.")

	if flagKeepImage {
		fmt.Printf("Image saved to: %s\n", imageDir)
	}

	fmt.Println("Ejecting source disc...")
	if err := cdtools.Eject(device); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not eject disc automatically (%v) — please eject manually.\n", err)
	}

	// ── Step 2: Write to blank disc ───────────────────────────────────────────
	fmt.Println("\n── Step 2 of 2: Write to blank disc ────────────────────────────")
	fmt.Println("Insert a blank CD-R into the drive, then press Enter...")
	waitForEnter()

	fmt.Printf("Writing disc at %dx speed...\n", flagSpeed)
	if err := writeDisc(device, tocFile, flagSpeed); err != nil {
		return fmt.Errorf("disc write failed: %w", err)
	}

	fmt.Println("Ejecting written disc...")
	_ = cdtools.Eject(device)

	fmt.Println("\nDisc copy complete!")
	return nil
}

// readDisc runs cdrdao read-cd in --read-raw mode for a bit-perfect audio copy.
func readDisc(device, tocFile, binFile string, speed int) error {
	args := []string{
		"read-cd",
		"--read-raw",                          // raw 2352-byte sectors — preserves all audio data
		"--device", device,
		"--speed", fmt.Sprintf("%d", speed),
		"--datafile", binFile,
		tocFile,
	}
	return runVisible("cdrdao", args...)
}

// writeDisc runs cdrdao write with the TOC file produced by readDisc.
func writeDisc(device, tocFile string, speed int) error {
	args := []string{
		"write",
		"--device", device,
		"--speed", fmt.Sprintf("%d", speed),
		"--eject",                             // auto-eject after writing
		tocFile,
	}
	return runVisible("cdrdao", args...)
}

// runVisible runs a command with its stdout/stderr piped to the terminal so
// the user can see progress output from cdrdao.
func runVisible(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return nil
}

func waitForEnter() {
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
}
