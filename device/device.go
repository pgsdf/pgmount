package device

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// Device represents a removable storage device
type Device struct {
	Name         string // e.g., "da0", "da0p1"
	Path         string // e.g., "/dev/da0p1"
	Label        string
	UUID         string
	FSType       string
	Size         uint64
	MountPoint   string
	IsMounted    bool
	IsEncrypted  bool
	IsUnlocked   bool
	IsPartition  bool
	IsRemovable  bool
	PartitionNum int
}

// Manager handles device detection and management
type Manager struct {
	devices map[string]*Device
}

// NewManager creates a new device manager
func NewManager() *Manager {
	return &Manager{
		devices: make(map[string]*Device),
	}
}

// Scan scans for all available devices
func (m *Manager) Scan() ([]*Device, error) {
	var devices []*Device
	var err error

	// Detect OS and use appropriate scanning method
	switch runtime.GOOS {
	case "freebsd":
		devices, err = m.scanFreeBSD()
	case "linux":
		devices, err = m.scanLinux()
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	if err != nil {
		return nil, err
	}

	// Update internal device map
	for _, dev := range devices {
		m.devices[dev.Path] = dev
	}

	return devices, nil
}

// scanFreeBSD scans for devices on FreeBSD
func (m *Manager) scanFreeBSD() ([]*Device, error) {
	devices := []*Device{}

	// Get list of devices from geom
	cmd := exec.Command("geom", "disk", "list")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list disks: %w", err)
	}

	// Parse geom output
	diskDevices := m.parseGeomDiskList(string(output))

	// For each disk, check partitions
	for _, disk := range diskDevices {
		// Check if removable
		isRemovable := m.isRemovableDevice(disk.Name)
		disk.IsRemovable = isRemovable

		if isRemovable {
			devices = append(devices, disk)

			// Get partitions
			partitions, err := m.getPartitions(disk.Name)
			if err == nil {
				devices = append(devices, partitions...)
			}
		}
	}

	return devices, nil
}

// scanLinux scans for devices on Linux
func (m *Manager) scanLinux() ([]*Device, error) {
	devices := []*Device{}

	// Use lsblk to list block devices
	cmd := exec.Command("lsblk", "-J", "-o", "NAME,SIZE,TYPE,MOUNTPOINT,FSTYPE,LABEL,UUID,RM,HOTPLUG")
	output, err := cmd.Output()
	if err != nil {
		// Fallback to simpler method if lsblk JSON fails
		return m.scanLinuxFallback()
	}

	// Parse lsblk JSON output
	linuxDevices := m.parseLsblkJSON(string(output))

	// Filter for removable devices and their partitions
	for _, dev := range linuxDevices {
		if dev.IsRemovable {
			devices = append(devices, dev)
		}
	}

	return devices, nil
}

// scanLinuxFallback uses simple lsblk without JSON
func (m *Manager) scanLinuxFallback() ([]*Device, error) {
	devices := []*Device{}

	// Read from /sys/block to find removable devices
	blockDevices, err := os.ReadDir("/sys/block")
	if err != nil {
		return nil, fmt.Errorf("failed to read /sys/block: %w", err)
	}

	for _, entry := range blockDevices {
		deviceName := entry.Name()

		// Check if removable
		removablePath := filepath.Join("/sys/block", deviceName, "removable")
		removableData, err := os.ReadFile(removablePath)
		if err != nil {
			continue
		}

		isRemovable := strings.TrimSpace(string(removableData)) == "1"
		if !isRemovable {
			continue
		}

		// Create device entry
		dev := &Device{
			Name:        deviceName,
			Path:        "/dev/" + deviceName,
			IsRemovable: true,
			IsPartition: false,
		}

		// Get size
		sizePath := filepath.Join("/sys/block", deviceName, "size")
		if sizeData, err := os.ReadFile(sizePath); err == nil {
			if sectors, err := strconv.ParseUint(strings.TrimSpace(string(sizeData)), 10, 64); err == nil {
				dev.Size = sectors * 512 // Convert sectors to bytes
			}
		}

		devices = append(devices, dev)

		// Find partitions
		partitions := m.findLinuxPartitions(deviceName)
		devices = append(devices, partitions...)
	}

	return devices, nil
}

// findLinuxPartitions finds partitions for a Linux block device
func (m *Manager) findLinuxPartitions(deviceName string) []*Device {
	partitions := []*Device{}

	deviceDir := filepath.Join("/sys/block", deviceName)
	entries, err := os.ReadDir(deviceDir)
	if err != nil {
		return partitions
	}

	for _, entry := range entries {
		partName := entry.Name()

		// Partitions are subdirectories that start with the device name
		if !strings.HasPrefix(partName, deviceName) {
			continue
		}

		// Skip if it's the device itself
		if partName == deviceName {
			continue
		}

		// Create partition device
		part := &Device{
			Name:        partName,
			Path:        "/dev/" + partName,
			IsPartition: true,
			IsRemovable: true,
		}

		// Get size
		sizePath := filepath.Join(deviceDir, partName, "size")
		if sizeData, err := os.ReadFile(sizePath); err == nil {
			if sectors, err := strconv.ParseUint(strings.TrimSpace(string(sizeData)), 10, 64); err == nil {
				part.Size = sectors * 512
			}
		}

		// Detect filesystem
		m.detectFilesystemLinux(part)

		// Check mount status
		m.checkMountStatus(part)

		partitions = append(partitions, part)
	}

	return partitions
}

// detectFilesystemLinux detects filesystem on Linux
func (m *Manager) detectFilesystemLinux(dev *Device) {
	// Try blkid to get filesystem info
	cmd := exec.Command("blkid", "-o", "export", dev.Path)
	output, err := cmd.Output()
	if err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(output)))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "TYPE=") {
				dev.FSType = strings.TrimPrefix(line, "TYPE=")
			} else if strings.HasPrefix(line, "LABEL=") {
				dev.Label = strings.TrimPrefix(line, "LABEL=")
			} else if strings.HasPrefix(line, "UUID=") {
				dev.UUID = strings.TrimPrefix(line, "UUID=")
			}
		}
	}

	// Fallback to file command
	if dev.FSType == "" {
		m.detectFilesystem(dev)
	}
}

// parseLsblkJSON parses lsblk JSON output
func (m *Manager) parseLsblkJSON(output string) []*Device {
	devices := []*Device{}

	// Simple JSON parsing for lsblk output
	// Format: {"blockdevices": [{"name": "sda", "size": "...", ...}, ...]}
	lines := strings.Split(output, "\n")
	var currentDevice *Device

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, `"name"`) {
			if idx := strings.Index(line, `"name"`); idx >= 0 {
				rest := line[idx+6:]
				if idx2 := strings.Index(rest, `"`); idx2 >= 0 {
					rest = rest[idx2+1:]
					if idx3 := strings.Index(rest, `"`); idx3 >= 0 {
						name := rest[:idx3]
						currentDevice = &Device{
							Name: name,
							Path: "/dev/" + name,
						}
					}
				}
			}
		}

		if currentDevice != nil {
			if strings.Contains(line, `"type"`) && strings.Contains(line, `"disk"`) {
				currentDevice.IsPartition = false
			} else if strings.Contains(line, `"type"`) && strings.Contains(line, `"part"`) {
				currentDevice.IsPartition = true
			}

			if strings.Contains(line, `"rm"`) && strings.Contains(line, `"1"`) {
				currentDevice.IsRemovable = true
			}

			if strings.Contains(line, `"hotplug"`) && strings.Contains(line, `"1"`) {
				currentDevice.IsRemovable = true
			}

			if strings.Contains(line, `"mountpoint"`) {
				if idx := strings.Index(line, `"mountpoint"`); idx >= 0 {
					rest := line[idx+13:]
					if idx2 := strings.Index(rest, `"`); idx2 >= 0 {
						rest = rest[idx2+1:]
						if idx3 := strings.Index(rest, `"`); idx3 >= 0 {
							mp := rest[:idx3]
							if mp != "" && mp != "null" {
								currentDevice.MountPoint = mp
								currentDevice.IsMounted = true
							}
						}
					}
				}
			}

			if strings.Contains(line, `"fstype"`) {
				if idx := strings.Index(line, `"fstype"`); idx >= 0 {
					rest := line[idx+9:]
					if idx2 := strings.Index(rest, `"`); idx2 >= 0 {
						rest = rest[idx2+1:]
						if idx3 := strings.Index(rest, `"`); idx3 >= 0 {
							fstype := rest[:idx3]
							if fstype != "" && fstype != "null" {
								currentDevice.FSType = fstype
							}
						}
					}
				}
			}

			if strings.Contains(line, `"label"`) {
				if idx := strings.Index(line, `"label"`); idx >= 0 {
					rest := line[idx+8:]
					if idx2 := strings.Index(rest, `"`); idx2 >= 0 {
						rest = rest[idx2+1:]
						if idx3 := strings.Index(rest, `"`); idx3 >= 0 {
							label := rest[:idx3]
							if label != "" && label != "null" {
								currentDevice.Label = label
							}
						}
					}
				}
			}

			if strings.Contains(line, `"uuid"`) {
				if idx := strings.Index(line, `"uuid"`); idx >= 0 {
					rest := line[idx+7:]
					if idx2 := strings.Index(rest, `"`); idx2 >= 0 {
						rest = rest[idx2+1:]
						if idx3 := strings.Index(rest, `"`); idx3 >= 0 {
							uuid := rest[:idx3]
							if uuid != "" && uuid != "null" {
								currentDevice.UUID = uuid
							}
						}
					}
				}
			}

			if strings.Contains(line, `"size"`) {
				if idx := strings.Index(line, `"size"`); idx >= 0 {
					rest := line[idx+7:]
					if idx2 := strings.Index(rest, `"`); idx2 >= 0 {
						rest = rest[idx2+1:]
						if idx3 := strings.Index(rest, `"`); idx3 >= 0 {
							sizeStr := rest[:idx3]
							if size, err := m.parseLinuxSize(sizeStr); err == nil {
								currentDevice.Size = size
							}
						}
					}
				}
			}

			// Check if we're at the end of a device object
			if strings.Contains(line, "}") && !strings.Contains(line, "},") {
				if currentDevice.Name != "" {
					devices = append(devices, currentDevice)
				}
				currentDevice = nil
			}
		}
	}

	return devices
}

// parseLinuxSize parses Linux size strings like "8G", "128M", etc.
func (m *Manager) parseLinuxSize(sizeStr string) (uint64, error) {
	sizeStr = strings.TrimSpace(sizeStr)
	if sizeStr == "" || sizeStr == "null" {
		return 0, fmt.Errorf("empty size")
	}

	// Handle numeric-only sizes (in bytes)
	if val, err := strconv.ParseUint(sizeStr, 10, 64); err == nil {
		return val, nil
	}

	// Handle sizes with units (K, M, G, T)
	multiplier := uint64(1)
	numStr := sizeStr

	if len(sizeStr) > 0 {
		lastChar := sizeStr[len(sizeStr)-1]
		switch lastChar {
		case 'K', 'k':
			multiplier = 1024
			numStr = sizeStr[:len(sizeStr)-1]
		case 'M', 'm':
			multiplier = 1024 * 1024
			numStr = sizeStr[:len(sizeStr)-1]
		case 'G', 'g':
			multiplier = 1024 * 1024 * 1024
			numStr = sizeStr[:len(sizeStr)-1]
		case 'T', 't':
			multiplier = 1024 * 1024 * 1024 * 1024
			numStr = sizeStr[:len(sizeStr)-1]
		}
	}

	if val, err := strconv.ParseFloat(numStr, 64); err == nil {
		return uint64(val * float64(multiplier)), nil
	}

	return 0, fmt.Errorf("failed to parse size: %s", sizeStr)
}

// GetDevice returns a device by path
func (m *Manager) GetDevice(path string) (*Device, bool) {
	dev, ok := m.devices[path]
	return dev, ok
}

// GetMountedDevices returns all mounted devices
func (m *Manager) GetMountedDevices() []*Device {
	mounted := []*Device{}
	for _, dev := range m.devices {
		if dev.IsMounted {
			mounted = append(mounted, dev)
		}
	}
	return mounted
}

// parseGeomDiskList parses the output of 'geom disk list'
func (m *Manager) parseGeomDiskList(output string) []*Device {
	devices := []*Device{}
	var current *Device

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "Geom name:") {
			if current != nil {
				devices = append(devices, current)
			}
			name := strings.TrimSpace(strings.TrimPrefix(line, "Geom name:"))
			current = &Device{
				Name: name,
				Path: "/dev/" + name,
			}
		} else if current != nil {
			if strings.HasPrefix(line, "Mediasize:") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					fmt.Sscanf(parts[1], "%d", &current.Size)
				}
			}
		}
	}

	if current != nil {
		devices = append(devices, current)
	}

	return devices
}

// getPartitions returns partitions for a disk
func (m *Manager) getPartitions(diskName string) ([]*Device, error) {
	partitions := []*Device{}

	// Use gpart to list partitions
	cmd := exec.Command("gpart", "show", "-p", diskName)
	output, err := cmd.Output()
	if err != nil {
		// Disk might not have partitions
		return partitions, nil
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		fields := strings.Fields(line)

		// Look for partition entries
		if len(fields) >= 4 && !strings.HasPrefix(line, "=>") {
			// Field format: start size index type [label]
			if len(fields[0]) > 0 && fields[0][0] >= '0' && fields[0][0] <= '9' {
				partName := ""
				for i := 3; i < len(fields); i++ {
					if strings.HasPrefix(fields[i], diskName) {
						partName = fields[i]
						break
					}
				}

				if partName != "" {
					part := &Device{
						Name:        partName,
						Path:        "/dev/" + partName,
						IsPartition: true,
						IsRemovable: true,
					}

					// Get filesystem info
					m.detectFilesystem(part)

					// Check if mounted
					m.checkMountStatus(part)

					partitions = append(partitions, part)
				}
			}
		}
	}

	return partitions, nil
}

// detectFilesystem detects the filesystem type and metadata
func (m *Manager) detectFilesystem(dev *Device) {
	// Try file -s to detect filesystem
	cmd := exec.Command("file", "-s", dev.Path)
	output, err := cmd.Output()
	if err == nil {
		fsInfo := string(output)

		// Parse filesystem type
		if strings.Contains(fsInfo, "ext2") || strings.Contains(fsInfo, "ext3") || strings.Contains(fsInfo, "ext4") {
			if strings.Contains(fsInfo, "ext4") {
				dev.FSType = "ext4"
			} else if strings.Contains(fsInfo, "ext3") {
				dev.FSType = "ext3"
			} else {
				dev.FSType = "ext2"
			}
		} else if strings.Contains(fsInfo, "FAT") {
			dev.FSType = "msdosfs"
		} else if strings.Contains(fsInfo, "NTFS") {
			dev.FSType = "ntfs"
		} else if strings.Contains(fsInfo, "UFS") {
			dev.FSType = "ufs"
		} else if strings.Contains(fsInfo, "ZFS") {
			dev.FSType = "zfs"
		}

		// Check for GELI encryption
		if strings.Contains(fsInfo, "GELI") {
			dev.IsEncrypted = true
		}

		// Extract UUID and label if available
		m.extractMetadata(dev)
	}
}

// extractMetadata extracts UUID and label
func (m *Manager) extractMetadata(dev *Device) {
	// Try to get label using glabel
	cmd := exec.Command("glabel", "status")
	output, err := cmd.Output()
	if err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(output)))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, dev.Name) {
				fields := strings.Fields(line)
				if len(fields) >= 1 {
					dev.Label = fields[0]
				}
			}
		}
	}

	// For ext filesystems, try dumpe2fs
	if strings.HasPrefix(dev.FSType, "ext") {
		cmd := exec.Command("dumpe2fs", "-h", dev.Path)
		output, err := cmd.Output()
		if err == nil {
			scanner := bufio.NewScanner(strings.NewReader(string(output)))
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "Filesystem UUID:") {
					dev.UUID = strings.TrimSpace(strings.TrimPrefix(line, "Filesystem UUID:"))
				} else if strings.HasPrefix(line, "Filesystem volume name:") {
					label := strings.TrimSpace(strings.TrimPrefix(line, "Filesystem volume name:"))
					if label != "<none>" && label != "" {
						dev.Label = label
					}
				}
			}
		}
	}
}

// checkMountStatus checks if a device is mounted
func (m *Manager) checkMountStatus(dev *Device) {
	file, err := os.Open("/etc/mtab")
	if err != nil {
		// Try /proc/mounts on some systems, or use mount command
		cmd := exec.Command("mount")
		output, err := cmd.Output()
		if err != nil {
			return
		}
		m.parseMountOutput(dev, string(output))
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, dev.Path+" ") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				dev.MountPoint = fields[1]
				dev.IsMounted = true
			}
		}
	}
}

// parseMountOutput parses mount command output
func (m *Manager) parseMountOutput(dev *Device, output string) {
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, dev.Path) {
			// Format: /dev/da0p1 on /media/USB (ufs, local)
			parts := strings.Split(line, " on ")
			if len(parts) >= 2 {
				mountInfo := strings.Split(parts[1], " (")
				if len(mountInfo) >= 1 {
					dev.MountPoint = strings.TrimSpace(mountInfo[0])
					dev.IsMounted = true
				}
			}
		}
	}
}

// isRemovableDevice checks if a device is removable
func (m *Manager) isRemovableDevice(name string) bool {
	// Check if it's a USB device (da*, umass*)
	if strings.HasPrefix(name, "da") || strings.HasPrefix(name, "umass") {
		return true
	}

	// Check CAM device info
	cmd := exec.Command("camcontrol", "devlist")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// Look for USB mass storage in device list
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, name) {
			if strings.Contains(strings.ToLower(line), "usb") ||
				strings.Contains(strings.ToLower(line), "mass storage") {
				return true
			}
		}
	}

	return false
}

// GetDisplayName returns a user-friendly display name
func (d *Device) GetDisplayName() string {
	if d.Label != "" {
		return d.Label
	}
	if d.UUID != "" {
		return d.UUID[:8] + "..."
	}
	return d.Name
}

// GetMountDirectory returns the preferred mount directory name
func (d *Device) GetMountDirectory(base string) string {
	name := d.GetDisplayName()
	// Sanitize the name
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, " ", "_")
	return filepath.Join(base, name)
}
