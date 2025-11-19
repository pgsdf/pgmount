package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/pgsdf/pgmount/device"
)

var (
	unmountAll = flag.Bool("a", false, "Unmount all mounted devices")
	detach     = flag.Bool("detach", false, "Also detach/eject the device after unmounting")
	force      = flag.Bool("f", false, "Force unmount")
	verbose    = flag.Bool("v", false, "Verbose output")
)

func main() {
	flag.Parse()

	// Initialize device manager
	mgr := device.NewManager()

	if *unmountAll {
		// Unmount all devices
		devices, err := mgr.Scan()
		if err != nil {
			log.Fatalf("Failed to scan devices: %v", err)
		}

		unmounted := 0
		for _, dev := range devices {
			if dev.IsMounted {
				if err := unmountDevice(dev); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to unmount %s: %v\n", dev.Path, err)
				} else {
					unmounted++
					fmt.Printf("Unmounted %s\n", dev.Path)

					if *detach {
						if err := detachDevice(dev); err != nil {
							fmt.Fprintf(os.Stderr, "Failed to detach %s: %v\n", dev.Path, err)
						}
					}
				}
			}
		}

		fmt.Printf("Unmounted %d device(s)\n", unmounted)
		return
	}

	// Unmount specific device or mount point
	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: pgumount [-a] [--detach] [-f] <device|mountpoint>\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	target := flag.Arg(0)

	// Scan for the device
	devices, err := mgr.Scan()
	if err != nil {
		log.Fatalf("Failed to scan devices: %v", err)
	}

	var targetDev *device.Device
	for _, dev := range devices {
		if dev.Path == target || dev.Name == target ||
			"/dev/"+dev.Name == target || dev.MountPoint == target {
			targetDev = dev
			break
		}
	}

	if targetDev == nil {
		log.Fatalf("Device not found: %s", target)
	}

	if !targetDev.IsMounted {
		log.Fatalf("Device not mounted: %s", targetDev.Path)
	}

	if err := unmountDevice(targetDev); err != nil {
		log.Fatalf("Failed to unmount device: %v", err)
	}

	fmt.Printf("Unmounted %s\n", targetDev.Path)

	if *detach {
		if err := detachDevice(targetDev); err != nil {
			log.Fatalf("Failed to detach device: %v", err)
		}
		fmt.Printf("Detached %s\n", targetDev.Path)
	}
}

func unmountDevice(dev *device.Device) error {
	args := []string{}

	if *force {
		args = append(args, "-f")
	}

	args = append(args, dev.MountPoint)

	if *verbose {
		log.Printf("Running: umount %v", args)
	}

	cmd := exec.Command("umount", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("unmount failed: %w (output: %s)", err, string(output))
	}

	// Remove mount point directory if empty
	os.Remove(dev.MountPoint)

	return nil
}

func detachDevice(dev *device.Device) error {
	// For USB devices, we can use camcontrol to eject/detach
	if *verbose {
		log.Printf("Detaching device %s", dev.Name)
	}

	// Get the disk name (strip partition number)
	diskName := dev.Name
	// Simple heuristic: if name ends with a digit preceded by 'p', strip it
	if len(diskName) > 2 {
		if diskName[len(diskName)-2] == 'p' {
			diskName = diskName[:len(diskName)-2]
		}
	}

	// Try to eject using camcontrol
	cmd := exec.Command("camcontrol", "eject", diskName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Try alternative method
		cmd = exec.Command("usbconfig", "-d", diskName, "power_off")
		output, err = cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("detach failed: %w (output: %s)", err, string(output))
		}
	}

	return nil
}
