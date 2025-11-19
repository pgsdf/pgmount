package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/pgsdf/pgmount/config"
	"github.com/pgsdf/pgmount/device"
)

var (
	mountAll   = flag.Bool("a", false, "Mount all available devices")
	verbose    = flag.Bool("v", false, "Verbose output")
	configFile = flag.String("config", "", "Path to configuration file")
	noConfig   = flag.Bool("no-config", false, "Don't use any config file")
	fsType     = flag.String("t", "", "Filesystem type")
	options    = flag.String("o", "", "Mount options (comma-separated)")
)

func main() {
	flag.Parse()

	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize device manager
	mgr := device.NewManager()

	if *mountAll {
		// Mount all devices
		devices, err := mgr.Scan()
		if err != nil {
			log.Fatalf("Failed to scan devices: %v", err)
		}

		mounted := 0
		for _, dev := range devices {
			if dev.IsPartition && !dev.IsMounted {
				if err := mountDevice(cfg, dev); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to mount %s: %v\n", dev.Path, err)
				} else {
					mounted++
					fmt.Printf("Mounted %s at %s\n", dev.Path, dev.MountPoint)
				}
			}
		}

		fmt.Printf("Mounted %d device(s)\n", mounted)
		return
	}

	// Mount specific device
	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: pgmount [-a] [-t fstype] [-o options] <device>\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	devicePath := flag.Arg(0)

	// Scan for the device
	devices, err := mgr.Scan()
	if err != nil {
		log.Fatalf("Failed to scan devices: %v", err)
	}

	var targetDev *device.Device
	for _, dev := range devices {
		if dev.Path == devicePath || dev.Name == devicePath ||
			"/dev/"+dev.Name == devicePath {
			targetDev = dev
			break
		}
	}

	if targetDev == nil {
		log.Fatalf("Device not found: %s", devicePath)
	}

	if targetDev.IsMounted {
		log.Fatalf("Device already mounted at %s", targetDev.MountPoint)
	}

	if err := mountDevice(cfg, targetDev); err != nil {
		log.Fatalf("Failed to mount device: %v", err)
	}

	fmt.Printf("Mounted %s at %s\n", targetDev.Path, targetDev.MountPoint)
}

func loadConfig() (*config.Config, error) {
	if *noConfig {
		return config.Default(), nil
	}

	path := *configFile
	if path == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		path = homeDir + "/.config/pgmount/config.yml"
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return config.Default(), nil
	}

	return config.Load(path)
}

func mountDevice(cfg *config.Config, dev *device.Device) error {
	// Determine mount point
	mountPoint := dev.GetMountDirectory(cfg.MountBase)

	// Create mount point if it doesn't exist
	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		return fmt.Errorf("failed to create mount point: %w", err)
	}

	// Get mount options
	opts := cfg.GetMountOptions(dev.FSType, dev.Label, dev.UUID, dev.Path)

	// Override with command-line options
	if *options != "" {
		opts = strings.Split(*options, ",")
	}

	// Override filesystem type
	fs := dev.FSType
	if *fsType != "" {
		fs = *fsType
	}

	// Build mount command
	args := []string{}
	if len(opts) > 0 {
		args = append(args, "-o", strings.Join(opts, ","))
	}
	if fs != "" && fs != "auto" {
		args = append(args, "-t", fs)
	}
	args = append(args, dev.Path, mountPoint)

	if *verbose {
		log.Printf("Running: mount %s", strings.Join(args, " "))
	}

	cmd := exec.Command("mount", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mount failed: %w (output: %s)", err, string(output))
	}

	dev.MountPoint = mountPoint
	dev.IsMounted = true

	return nil
}
