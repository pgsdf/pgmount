package daemon

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/pgsdf/pgmount/config"
	"github.com/pgsdf/pgmount/device"
	"github.com/pgsdf/pgmount/notify"
)

// Daemon handles automounting and device events
type Daemon struct {
	config      *config.Config
	deviceMgr   *device.Manager
	devdPipe    *os.File
	stopChan    chan struct{}
	wg          sync.WaitGroup
	mu          sync.Mutex
	mounted     map[string]*device.Device
}

// New creates a new daemon instance
func New(cfg *config.Config) (*Daemon, error) {
	return &Daemon{
		config:    cfg,
		deviceMgr: device.NewManager(),
		stopChan:  make(chan struct{}),
		mounted:   make(map[string]*device.Device),
	}, nil
}

// Start starts the daemon
func (d *Daemon) Start() error {
	log.Println("Starting daemon...")

	// Initial scan for existing devices
	devices, err := d.deviceMgr.Scan()
	if err != nil {
		return fmt.Errorf("failed to scan devices: %w", err)
	}

	log.Printf("Found %d removable devices", len(devices))

	// Start devd event monitor
	d.wg.Add(1)
	go d.monitorDevd()

	return nil
}

// Stop stops the daemon
func (d *Daemon) Stop() {
	log.Println("Stopping daemon...")
	close(d.stopChan)
	if d.devdPipe != nil {
		d.devdPipe.Close()
	}
	d.wg.Wait()
}

// MountAll mounts all available devices
func (d *Daemon) MountAll() error {
	devices, err := d.deviceMgr.Scan()
	if err != nil {
		return fmt.Errorf("failed to scan devices: %w", err)
	}

	for _, dev := range devices {
		if dev.IsPartition && !dev.IsMounted {
			if err := d.mountDevice(dev); err != nil {
				log.Printf("Failed to mount %s: %v", dev.Path, err)
			}
		}
	}

	return nil
}

// monitorDevd monitors devd for device events
func (d *Daemon) monitorDevd() {
	defer d.wg.Done()

	// Open devd socket pipe
	// In practice, we'd connect to devd's socket at /var/run/devd.seqpacket.pipe
	// For now, we'll simulate by monitoring system logs or using a simpler approach
	
	// Alternative: poll for device changes
	d.pollDevices()
}

// pollDevices periodically checks for device changes
func (d *Daemon) pollDevices() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	knownDevices := make(map[string]bool)

	for {
		select {
		case <-d.stopChan:
			return
		case <-ticker.C:
			devices, err := d.deviceMgr.Scan()
			if err != nil {
				log.Printf("Failed to scan devices: %v", err)
				continue
			}

			currentDevices := make(map[string]bool)

			// Check for new devices
			for _, dev := range devices {
				currentDevices[dev.Path] = true

				if !knownDevices[dev.Path] {
					// New device detected
					d.onDeviceAdded(dev)
					knownDevices[dev.Path] = true
				}
			}

			// Check for removed devices
			for path := range knownDevices {
				if !currentDevices[path] {
					d.onDeviceRemoved(path)
					delete(knownDevices, path)
				}
			}
		}
	}
}

// onDeviceAdded handles device addition
func (d *Daemon) onDeviceAdded(dev *device.Device) {
	log.Printf("Device added: %s (%s)", dev.Path, dev.GetDisplayName())

	// Check if device should be ignored
	if d.config.ShouldIgnoreDevice(dev.Label, dev.UUID, dev.Path) {
		log.Printf("Ignoring device %s", dev.Path)
		return
	}

	// Send notification
	if d.config.Notifications.Enabled && d.config.Notifications.DeviceAdded > 0 {
		notify.Send("Device Added", fmt.Sprintf("%s connected", dev.GetDisplayName()),
			int(d.config.Notifications.DeviceAdded*1000))
	}

	// Execute event hook
	d.executeEventHook("device_added", dev)

	// Auto-mount if enabled
	if dev.IsPartition && d.config.ShouldAutomountDevice(dev.Label, dev.UUID, dev.Path) {
		if err := d.mountDevice(dev); err != nil {
			log.Printf("Failed to automount %s: %v", dev.Path, err)
			
			if d.config.Notifications.Enabled && d.config.Notifications.JobFailed > 0 {
				notify.Send("Mount Failed", fmt.Sprintf("Failed to mount %s: %v", dev.GetDisplayName(), err),
					int(d.config.Notifications.JobFailed*1000))
			}
		}
	}
}

// onDeviceRemoved handles device removal
func (d *Daemon) onDeviceRemoved(path string) {
	log.Printf("Device removed: %s", path)

	d.mu.Lock()
	dev, ok := d.mounted[path]
	d.mu.Unlock()

	if ok {
		// Device was mounted, unmount it
		if err := d.unmountDevice(dev); err != nil {
			log.Printf("Failed to unmount %s: %v", path, err)
		}
	}

	// Send notification
	if d.config.Notifications.Enabled && d.config.Notifications.DeviceRemoved > 0 {
		displayName := path
		if dev != nil {
			displayName = dev.GetDisplayName()
		}
		notify.Send("Device Removed", fmt.Sprintf("%s disconnected", displayName),
			int(d.config.Notifications.DeviceRemoved*1000))
	}
}

// mountDevice mounts a device
func (d *Daemon) mountDevice(dev *device.Device) error {
	if dev.IsMounted {
		return fmt.Errorf("device already mounted at %s", dev.MountPoint)
	}

	// Handle encrypted devices
	if dev.IsEncrypted && !dev.IsUnlocked {
		if err := d.unlockDevice(dev); err != nil {
			return fmt.Errorf("failed to unlock device: %w", err)
		}
	}

	// Determine mount point
	mountPoint := dev.GetMountDirectory(d.config.MountBase)

	// Create mount point if it doesn't exist
	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		return fmt.Errorf("failed to create mount point: %w", err)
	}

	// Get mount options
	opts := d.config.GetMountOptions(dev.FSType, dev.Label, dev.UUID, dev.Path)

	// Build mount command
	args := []string{}
	if len(opts) > 0 {
		args = append(args, "-o", strings.Join(opts, ","))
	}
	if dev.FSType != "" && dev.FSType != "auto" {
		args = append(args, "-t", dev.FSType)
	}
	args = append(args, dev.Path, mountPoint)

	log.Printf("Mounting %s at %s (fstype: %s)", dev.Path, mountPoint, dev.FSType)

	cmd := exec.Command("mount", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mount failed: %w (output: %s)", err, string(output))
	}

	dev.MountPoint = mountPoint
	dev.IsMounted = true

	d.mu.Lock()
	d.mounted[dev.Path] = dev
	d.mu.Unlock()

	log.Printf("Successfully mounted %s at %s", dev.Path, mountPoint)

	// Send notification
	if d.config.Notifications.Enabled && d.config.Notifications.DeviceMounted > 0 {
		notify.Send("Device Mounted", fmt.Sprintf("%s mounted at %s", dev.GetDisplayName(), mountPoint),
			int(d.config.Notifications.DeviceMounted*1000))
	}

	// Execute event hook
	d.executeEventHook("device_mounted", dev)

	// Open in file manager if configured
	if d.config.FileManager != "" {
		go d.openInFileManager(mountPoint)
	}

	return nil
}

// unmountDevice unmounts a device
func (d *Daemon) unmountDevice(dev *device.Device) error {
	if !dev.IsMounted {
		return fmt.Errorf("device not mounted")
	}

	log.Printf("Unmounting %s from %s", dev.Path, dev.MountPoint)

	cmd := exec.Command("umount", dev.MountPoint)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("unmount failed: %w (output: %s)", err, string(output))
	}

	mountPoint := dev.MountPoint
	dev.MountPoint = ""
	dev.IsMounted = false

	d.mu.Lock()
	delete(d.mounted, dev.Path)
	d.mu.Unlock()

	// Remove mount point directory if empty
	os.Remove(mountPoint)

	log.Printf("Successfully unmounted %s", dev.Path)

	// Send notification
	if d.config.Notifications.Enabled && d.config.Notifications.DeviceUnmounted > 0 {
		notify.Send("Device Unmounted", fmt.Sprintf("%s unmounted", dev.GetDisplayName()),
			int(d.config.Notifications.DeviceUnmounted*1000))
	}

	// Execute event hook
	d.executeEventHook("device_unmounted", dev)

	return nil
}

// unlockDevice unlocks a GELI encrypted device
func (d *Daemon) unlockDevice(dev *device.Device) error {
	if !d.config.GELI.Enabled {
		return fmt.Errorf("GELI support is disabled")
	}

	log.Printf("Unlocking encrypted device %s", dev.Path)

	// Check for keyfile
	keyfile, hasKeyfile := d.config.GELI.KeyFiles[dev.UUID]
	
	var cmd *exec.Cmd
	if hasKeyfile {
		// Use keyfile
		cmd = exec.Command("geli", "attach", "-k", keyfile, dev.Path)
	} else {
		// Prompt for password
		password, err := d.getPassword(dev)
		if err != nil {
			return fmt.Errorf("failed to get password: %w", err)
		}

		cmd = exec.Command("geli", "attach", dev.Path)
		cmd.Stdin = strings.NewReader(password + "\n")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("geli attach failed: %w (output: %s)", err, string(output))
	}

	dev.IsUnlocked = true

	log.Printf("Successfully unlocked %s", dev.Path)

	// Send notification
	if d.config.Notifications.Enabled && d.config.Notifications.DeviceUnlocked > 0 {
		notify.Send("Device Unlocked", fmt.Sprintf("%s unlocked", dev.GetDisplayName()),
			int(d.config.Notifications.DeviceUnlocked*1000))
	}

	return nil
}

// getPassword prompts for a password
func (d *Daemon) getPassword(dev *device.Device) (string, error) {
	if d.config.GELI.PasswordCmd != "" {
		// Use custom password command
		cmd := exec.Command("sh", "-c", d.config.GELI.PasswordCmd)
		output, err := cmd.Output()
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(output)), nil
	}

	// Use built-in password prompt
	fmt.Printf("Enter password for %s: ", dev.GetDisplayName())
	reader := bufio.NewReader(os.Stdin)
	password, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(password), nil
}

// executeEventHook executes an event hook if configured
func (d *Daemon) executeEventHook(event string, dev *device.Device) {
	if hookCmd, ok := d.config.EventHooks[event]; ok {
		// Replace placeholders
		cmd := strings.ReplaceAll(hookCmd, "{device}", dev.Path)
		cmd = strings.ReplaceAll(cmd, "{label}", dev.Label)
		cmd = strings.ReplaceAll(cmd, "{uuid}", dev.UUID)
		cmd = strings.ReplaceAll(cmd, "{mount_point}", dev.MountPoint)

		log.Printf("Executing event hook for %s: %s", event, cmd)

		go func() {
			execCmd := exec.Command("sh", "-c", cmd)
			if err := execCmd.Run(); err != nil {
				log.Printf("Event hook failed: %v", err)
			}
		}()
	}
}

// GetDeviceManager returns the device manager
func (d *Daemon) GetDeviceManager() *device.Manager {
	return d.deviceMgr
}

// MountDevice mounts a specific device (public method for tray integration)
func (d *Daemon) MountDevice(dev *device.Device) error {
	return d.mountDevice(dev)
}

// UnmountDevice unmounts a specific device (public method for tray integration)
func (d *Daemon) UnmountDevice(dev *device.Device) error {
	return d.unmountDevice(dev)
}

// openInFileManager opens a path in the configured file manager
func (d *Daemon) openInFileManager(path string) {
	cmd := exec.Command(d.config.FileManager, path)
	if err := cmd.Start(); err != nil {
		log.Printf("Failed to open file manager: %v", err)
	}
}
