package device

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

	// Update internal device map
	for _, dev := range devices {
		m.devices[dev.Path] = dev
	}

	return devices, nil
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
