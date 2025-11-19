package tray

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	shellquote "github.com/kballard/go-shellquote"
	"fyne.io/systray"
	"github.com/pgsdf/pgmount/config"
	"github.com/pgsdf/pgmount/device"
)

// Icon represents a system tray icon
type Icon struct {
	config        *config.Config
	deviceMgr     *device.Manager
	visible       bool
	updateChan    chan struct{}
	closeChan     chan struct{}
	quitChan      chan struct{}
	readyChan     chan struct{}
	menuMutex     sync.Mutex
	menuCloseChan chan struct{}
	onMountFunc   func(dev *device.Device) error
	onUnmountFunc func(dev *device.Device) error
	onQuitFunc    func()
}

// New creates a new tray icon
func New(cfg *config.Config, mgr *device.Manager) (*Icon, error) {
	icon := &Icon{
		config:        cfg,
		deviceMgr:     mgr,
		visible:       true,
		updateChan:    make(chan struct{}, 1),
		closeChan:     make(chan struct{}),
		quitChan:      make(chan struct{}),
		readyChan:     make(chan struct{}),
		menuCloseChan: make(chan struct{}),
	}

	// Start systray in a goroutine
	go systray.Run(icon.onReady, icon.onExit)

	// Wait for systray to be ready
	<-icon.readyChan

	log.Println("Tray icon initialized successfully")

	// Start update handler after systray is ready
	go icon.handleUpdates()

	return icon, nil
}

// onReady is called when systray is ready
func (i *Icon) onReady() {
	// Set icon and tooltip
	systray.SetIcon(getIcon())
	systray.SetTitle("PGMount")
	systray.SetTooltip("PGMount - Removable Media Manager")

	// Build initial menu
	i.rebuildMenu()

	// Signal that systray is ready
	close(i.readyChan)
}

// onExit is called when systray is exiting
func (i *Icon) onExit() {
	log.Println("Tray icon exiting")
}

// handleUpdates handles update requests
func (i *Icon) handleUpdates() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-i.closeChan:
			return
		case <-i.updateChan:
			i.rebuildMenu()
		case <-ticker.C:
			i.rebuildMenu()
		}
	}
}

// rebuildMenu rebuilds the entire menu
func (i *Icon) rebuildMenu() {
	// Lock to prevent concurrent menu modifications
	i.menuMutex.Lock()
	defer i.menuMutex.Unlock()

	// Get current devices
	devices, err := i.deviceMgr.Scan()
	if err != nil {
		log.Printf("Failed to scan devices: %v", err)
		return
	}

	// Filter for partitions, but also include whole disks if they have no partitions
	displayDevices := []*device.Device{}
	diskHasPartitions := make(map[string]bool)

	// First pass: collect all partitions and track which disks have partitions
	for _, dev := range devices {
		if dev.IsPartition && dev.IsRemovable {
			displayDevices = append(displayDevices, dev)
			// Extract parent disk name (e.g., "da0" from "da0p1")
			parentDisk := getParentDiskName(dev.Name)
			if parentDisk != "" {
				diskHasPartitions[parentDisk] = true
			}
		}
	}

	// Second pass: add whole disks that don't have partitions
	for _, dev := range devices {
		if !dev.IsPartition && dev.IsRemovable {
			if !diskHasPartitions[dev.Name] {
				displayDevices = append(displayDevices, dev)
				log.Printf("Adding unpartitioned disk to tray: %s", dev.Name)
			}
		}
	}

	// Stop all existing menu item handlers before clearing the menu
	close(i.menuCloseChan)

	// Clear existing menu
	systray.ResetMenu()

	// Create new channel for new menu item handlers
	i.menuCloseChan = make(chan struct{})

	// Add header
	systray.AddMenuItem("Devices", "Removable devices").Disable()
	systray.AddSeparator()

	// Add device menu items
	if len(displayDevices) == 0 {
		systray.AddMenuItem("No devices", "No removable devices found").Disable()
	} else {
		i.addDeviceMenuItems(displayDevices, i.menuCloseChan)
	}

	systray.AddSeparator()

	// Capture the menu close channel before starting handlers
	// to prevent race condition where handlers capture the wrong channel
	menuCloseChan := i.menuCloseChan

	// Add "Mount All"
	mMountAll := systray.AddMenuItem("Mount All", "Mount all available devices")
	go i.handleMenuItem(mMountAll, menuCloseChan, func() { i.onMountAll() })

	// Add "Unmount All"
	mUnmountAll := systray.AddMenuItem("Unmount All", "Unmount all mounted devices")
	go i.handleMenuItem(mUnmountAll, menuCloseChan, func() { i.onUnmountAll() })

	systray.AddSeparator()

	// Add "Refresh"
	mRefresh := systray.AddMenuItem("Refresh", "Refresh device list")
	go i.handleMenuItem(mRefresh, menuCloseChan, func() { i.onRefresh() })

	// Add "About"
	mAbout := systray.AddMenuItem("About", "About PGMount")
	go i.handleMenuItem(mAbout, menuCloseChan, func() { i.onAbout() })

	systray.AddSeparator()

	// Add "Quit"
	mQuit := systray.AddMenuItem("Quit", "Quit PGMount")
	go i.handleMenuItem(mQuit, menuCloseChan, func() { i.onQuit() })

	// Handle auto-hide
	if i.config.Tray.AutoHide {
		i.visible = len(displayDevices) > 0
	}
}

// addDeviceMenuItems adds device-specific menu items
func (i *Icon) addDeviceMenuItems(devices []*device.Device, menuCloseChan chan struct{}) {
	for _, dev := range devices {
		// Create a copy for the closure
		device := dev

		// Create device menu item
		displayName := device.GetDisplayName()
		if device.IsMounted {
			displayName += " ●"
		}

		// Mark whole disks (unpartitioned) with a special indicator
		if !device.IsPartition {
			displayName += " [Raw Disk]"
		}

		mDevice := systray.AddMenuItem(displayName, device.Path)

		// Handle whole disk vs partition differently
		if !device.IsPartition {
			// Whole disk (no partitions) - can't be mounted directly
			mInfo := mDevice.AddSubMenuItem("No partitions found", "This disk has no partition table")
			mInfo.Disable()
			mDevice.AddSubMenuItem("Format/partition this disk using Disk Utility", "Use gpart or other tools").Disable()
		} else if device.IsMounted {
			// Mounted partition
			// Add "Open" option
			mOpen := mDevice.AddSubMenuItem("Open in File Manager", "Open device in file manager")
			go i.handleMenuItem(mOpen, menuCloseChan, func() { i.onOpenDevice(device) })

			// Add "Unmount" option
			mUnmount := mDevice.AddSubMenuItem("Unmount", "Unmount device")
			go i.handleMenuItem(mUnmount, menuCloseChan, func() { i.onUnmountDevice(device) })

			// Add "Eject" option
			if device.IsRemovable {
				mEject := mDevice.AddSubMenuItem("Eject", "Eject device")
				go i.handleMenuItem(mEject, menuCloseChan, func() { i.onEjectDevice(device) })
			}
		} else {
			// Unmounted partition
			// Add "Mount" option
			mMount := mDevice.AddSubMenuItem("Mount", "Mount device")
			go i.handleMenuItem(mMount, menuCloseChan, func() { i.onMountDevice(device) })
		}

		// Add device info
		infoText := fmt.Sprintf("%s", device.Path)
		if device.FSType != "" {
			infoText += fmt.Sprintf(" • %s", device.FSType)
		}
		if device.Size > 0 {
			infoText += fmt.Sprintf(" • %s", formatSize(device.Size))
		}
		mDevice.AddSubMenuItem(infoText, "Device information").Disable()
	}
}

// handleMenuItem handles menu item clicks
func (i *Icon) handleMenuItem(item *systray.MenuItem, menuCloseChan chan struct{}, action func()) {
	for {
		select {
		case <-item.ClickedCh:
			action()
		case <-menuCloseChan:
			// Menu is being rebuilt, stop this handler
			return
		case <-i.closeChan:
			return
		}
	}
}

// Show makes the tray icon visible
func (i *Icon) Show() {
	// systray doesn't support Show/Hide in v1.11.0
	// The icon visibility is controlled by systray.Run()
	i.visible = true
	log.Println("Tray icon visibility: shown")
}

// Hide hides the tray icon
func (i *Icon) Hide() {
	// systray doesn't support Show/Hide in v1.11.0
	// We track the state for auto-hide logic but can't actually hide the icon
	i.visible = false
	log.Println("Tray icon visibility: hidden (state tracked, icon remains visible)")
}

// UpdateDevices triggers a menu update
func (i *Icon) UpdateDevices() {
	select {
	case i.updateChan <- struct{}{}:
	default:
		// Update already pending
	}
}

// Close closes the tray icon
func (i *Icon) Close() {
	close(i.closeChan)
	systray.Quit()
	log.Println("Tray icon closed")
}

// SetMountCallback sets the callback for mounting devices
func (i *Icon) SetMountCallback(fn func(dev *device.Device) error) {
	i.onMountFunc = fn
}

// SetUnmountCallback sets the callback for unmounting devices
func (i *Icon) SetUnmountCallback(fn func(dev *device.Device) error) {
	i.onUnmountFunc = fn
}

// SetQuitCallback sets the callback for quit action
func (i *Icon) SetQuitCallback(fn func()) {
	i.onQuitFunc = fn
}

// Menu action handlers

func (i *Icon) onMountDevice(dev *device.Device) {
	log.Printf("Tray: Mount device %s", dev.Path)

	if i.onMountFunc != nil {
		if err := i.onMountFunc(dev); err != nil {
			log.Printf("Failed to mount %s: %v", dev.GetDisplayName(), err)
			i.showNotification("Mount Failed", fmt.Sprintf("Failed to mount %s: %v", dev.GetDisplayName(), err))
		} else {
			i.showNotification("Device Mounted", fmt.Sprintf("%s mounted successfully", dev.GetDisplayName()))
			i.UpdateDevices()
		}
	} else {
		// Fallback: call pgmount command
		cmd := exec.Command("pgmount", dev.Path)
		if err := cmd.Run(); err != nil {
			log.Printf("Failed to mount %s: %v", dev.GetDisplayName(), err)
			i.showNotification("Mount Failed", fmt.Sprintf("Failed to mount %s: %v", dev.GetDisplayName(), err))
		} else {
			i.showNotification("Device Mounted", fmt.Sprintf("%s mounted successfully", dev.GetDisplayName()))
			i.UpdateDevices()
		}
	}
}

func (i *Icon) onUnmountDevice(dev *device.Device) {
	log.Printf("Tray: Unmount device %s", dev.Path)

	if i.onUnmountFunc != nil {
		if err := i.onUnmountFunc(dev); err != nil {
			log.Printf("Failed to unmount %s: %v", dev.GetDisplayName(), err)
			i.showNotification("Unmount Failed", fmt.Sprintf("Failed to unmount %s: %v", dev.GetDisplayName(), err))
		} else {
			i.showNotification("Device Unmounted", fmt.Sprintf("%s unmounted successfully", dev.GetDisplayName()))
			i.UpdateDevices()
		}
	} else {
		// Fallback: call pgumount command
		cmd := exec.Command("pgumount", dev.Path)
		if err := cmd.Run(); err != nil {
			log.Printf("Failed to unmount %s: %v", dev.GetDisplayName(), err)
			i.showNotification("Unmount Failed", fmt.Sprintf("Failed to unmount %s: %v", dev.GetDisplayName(), err))
		} else {
			i.showNotification("Device Unmounted", fmt.Sprintf("%s unmounted successfully", dev.GetDisplayName()))
			i.UpdateDevices()
		}
	}
}

func (i *Icon) onEjectDevice(dev *device.Device) {
	log.Printf("Tray: Eject device %s", dev.Path)

	// First unmount
	i.onUnmountDevice(dev)

	// Then eject
	cmd := exec.Command("pgumount", "--detach", dev.Path)
	if err := cmd.Run(); err != nil {
		log.Printf("Failed to eject %s: %v", dev.GetDisplayName(), err)
		i.showNotification("Eject Failed", fmt.Sprintf("Failed to eject %s: %v", dev.GetDisplayName(), err))
	} else {
		i.showNotification("Device Ejected", fmt.Sprintf("%s ejected successfully", dev.GetDisplayName()))
		i.UpdateDevices()
	}
}

func (i *Icon) onOpenDevice(dev *device.Device) {
	log.Printf("Tray: Open device %s", dev.Path)

	if dev.MountPoint == "" {
		return
	}

	// Validate that the path is absolute and clean to prevent command injection
	absPath, err := filepath.Abs(dev.MountPoint)
	if err != nil {
		log.Printf("Failed to get absolute path: %v", err)
		i.showNotification("Open Failed", fmt.Sprintf("Invalid path: %v", err))
		return
	}

	// Clean the path to remove any .. or other traversal attempts
	cleanPath := filepath.Clean(absPath)

	// Verify the path exists
	if _, err := os.Stat(cleanPath); err != nil {
		log.Printf("Path does not exist: %v", err)
		i.showNotification("Open Failed", "Mount point does not exist")
		return
	}

	fileManager := i.config.FileManager
	if fileManager == "" {
		fileManager = "xdg-open"
	}

	// Parse file manager command to handle arguments safely
	parts, err := shellquote.Split(fileManager)
	if err != nil || len(parts) == 0 {
		log.Printf("Invalid file manager command: %v", err)
		i.showNotification("Open Failed", "Invalid file manager configuration")
		return
	}

	// Execute file manager with path as separate argument
	args := append(parts[1:], cleanPath)
	cmd := exec.Command(parts[0], args...)
	if err := cmd.Start(); err != nil {
		log.Printf("Failed to open file manager: %v", err)
		i.showNotification("Open Failed", fmt.Sprintf("Failed to open file manager: %v", err))
	}
}

func (i *Icon) onMountAll() {
	log.Println("Tray: Mount All clicked")

	devices, err := i.deviceMgr.Scan()
	if err != nil {
		log.Printf("Failed to scan devices: %v", err)
		return
	}

	mounted := 0
	for _, dev := range devices {
		if dev.IsPartition && !dev.IsMounted && dev.IsRemovable {
			if i.onMountFunc != nil {
				if err := i.onMountFunc(dev); err != nil {
					log.Printf("Failed to mount %s: %v", dev.Path, err)
				} else {
					mounted++
				}
			}
		}
	}

	i.UpdateDevices()

	if mounted > 0 {
		i.showNotification("Mount All", fmt.Sprintf("Mounted %d device(s)", mounted))
	}
}

func (i *Icon) onUnmountAll() {
	log.Println("Tray: Unmount All clicked")

	devices, err := i.deviceMgr.Scan()
	if err != nil {
		log.Printf("Failed to scan devices: %v", err)
		return
	}

	unmounted := 0
	for _, dev := range devices {
		if dev.IsPartition && dev.IsMounted && dev.IsRemovable {
			if i.onUnmountFunc != nil {
				if err := i.onUnmountFunc(dev); err != nil {
					log.Printf("Failed to unmount %s: %v", dev.Path, err)
				} else {
					unmounted++
				}
			}
		}
	}

	i.UpdateDevices()

	if unmounted > 0 {
		i.showNotification("Unmount All", fmt.Sprintf("Unmounted %d device(s)", unmounted))
	}
}

func (i *Icon) onRefresh() {
	log.Println("Tray: Refresh clicked")
	i.UpdateDevices()
}

func (i *Icon) onAbout() {
	log.Println("Tray: About clicked")
	i.showNotification("About PGMount", "PGMount v1.0.0\nAutomount daemon for FreeBSD/GhostBSD\nPacific Grove Software Distribution Foundation")
}

func (i *Icon) onQuit() {
	log.Println("Tray: Quit clicked")
	if i.onQuitFunc != nil {
		i.onQuitFunc()
	} else {
		// Fallback to closing the tray icon if no callback is set
		i.Close()
	}
}

// Helper functions

func (i *Icon) showNotification(title, message string) {
	// Use notify-send if available
	cmd := exec.Command("notify-send", title, message)
	if err := cmd.Run(); err != nil {
		log.Printf("Failed to send notification: %v", err)
	}
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

// getParentDiskName extracts the parent disk name from a partition name
// For example: "da0p1" -> "da0", "ada0s1" -> "ada0"
func getParentDiskName(partitionName string) string {
	// Handle common FreeBSD partition naming schemes:
	// - da0p1, da0p2 (GPT partitions)
	// - da0s1, da0s2 (MBR slices)
	// - ada0p1, ada0p2
	// - mmcsd0s1, etc.

	// Find the first occurrence of 'p' or 's' followed by a digit
	for i := 0; i < len(partitionName); i++ {
		if (partitionName[i] == 'p' || partitionName[i] == 's') && i > 0 {
			// Check if next character is a digit
			if i+1 < len(partitionName) && partitionName[i+1] >= '0' && partitionName[i+1] <= '9' {
				return partitionName[:i]
			}
		}
	}

	return ""
}

// getIcon returns the icon data for the tray
func getIcon() []byte {
	// Simple drive icon as PNG (embedded as base64 or bytes)
	// For now, return empty - systray will use default
	// You can embed an icon here or load from file
	return []byte{}
}
