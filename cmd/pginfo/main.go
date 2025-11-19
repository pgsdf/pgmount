package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/pgsdf/pgmount/device"
)

var (
	showAll = flag.Bool("a", false, "Show all devices")
	verbose = flag.Bool("v", false, "Verbose output")
)

func main() {
	flag.Parse()

	// Initialize device manager
	mgr := device.NewManager()

	// Scan for devices
	devices, err := mgr.Scan()
	if err != nil {
		log.Fatalf("Failed to scan devices: %v", err)
	}

	if len(devices) == 0 {
		fmt.Println("No removable devices found")
		return
	}

	// Create tabwriter for formatted output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	if *verbose {
		fmt.Fprintln(w, "DEVICE\tLABEL\tUUID\tFSTYPE\tSIZE\tMOUNTED\tMOUNT POINT\tENCRYPTED")
		fmt.Fprintln(w, "------\t-----\t----\t------\t----\t-------\t-----------\t---------")
	} else {
		fmt.Fprintln(w, "DEVICE\tLABEL\tMOUNTED\tMOUNT POINT")
		fmt.Fprintln(w, "------\t-----\t-------\t-----------")
	}

	for _, dev := range devices {
		// Skip non-partitions unless showing all
		if !dev.IsPartition && !*showAll {
			continue
		}

		if *verbose {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%v\t%s\t%v\n",
				dev.Path,
				dev.Label,
				truncateString(dev.UUID, 8),
				dev.FSType,
				formatSize(dev.Size),
				dev.IsMounted,
				dev.MountPoint,
				dev.IsEncrypted,
			)
		} else {
			mounted := "No"
			if dev.IsMounted {
				mounted = "Yes"
			}

			label := dev.Label
			if label == "" {
				label = dev.Name
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				dev.Path,
				label,
				mounted,
				dev.MountPoint,
			)
		}
	}

	w.Flush()
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func formatSize(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
