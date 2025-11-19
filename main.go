package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pgsdf/pgmount/config"
	"github.com/pgsdf/pgmount/daemon"
	"github.com/pgsdf/pgmount/device"
	"github.com/pgsdf/pgmount/notify"
	"github.com/pgsdf/pgmount/tray"
)

const (
	Version = "1.0.0"
)

var (
	showVersion   = flag.Bool("version", false, "Show version information")
	configFile    = flag.String("config", "", "Path to configuration file")
	noConfig      = flag.Bool("no-config", false, "Don't use any config file")
	automount     = flag.Bool("automount", true, "Enable automounting new devices")
	noAutomount   = flag.Bool("no-automount", false, "Disable automounting new devices")
	notifications = flag.Bool("notify", true, "Enable pop-up notifications")
	noNotify      = flag.Bool("no-notify", false, "Disable pop-up notifications")
	showTray      = flag.Bool("tray", false, "Show tray icon")
	autoTray      = flag.Bool("auto-tray", false, "Auto-hide tray when no devices")
	noTray        = flag.Bool("no-tray", false, "Disable tray icon")
	verbose       = flag.Bool("verbose", false, "Verbose output")
	quiet         = flag.Bool("quiet", false, "Quiet output")
	mountAll      = flag.Bool("mount-all", false, "Mount all available devices")
	daemonMode    = flag.Bool("daemon", true, "Run as daemon")
)

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("pgmountd version %s\n", Version)
		fmt.Println("Automounter daemon for removable media on FreeBSD/GhostBSD")
		os.Exit(0)
	}

	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Apply command line flags
	applyFlags(cfg)

	// Initialize logger
	initLogger(cfg)

	log.Printf("Starting pgmountd v%s", Version)

	// Initialize notification system
	if cfg.Notifications.Enabled {
		if err := notify.Init(); err != nil {
			log.Printf("Warning: Failed to initialize notifications: %v", err)
			cfg.Notifications.Enabled = false
		}
		defer notify.Close()
	}

	// Initialize daemon
	d, err := daemon.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize daemon: %v", err)
	}

	// Mount all devices if requested
	if *mountAll {
		if err := d.MountAll(); err != nil {
			log.Fatalf("Failed to mount all devices: %v", err)
		}
		if !*daemonMode {
			return
		}
	}

	// Initialize tray icon if enabled
	var trayIcon *tray.Icon
	var trayStopChan chan struct{}
	if cfg.Tray.Enabled {
		trayIcon, err = tray.New(cfg, d.GetDeviceManager())
		if err != nil {
			log.Printf("Warning: Failed to initialize tray icon: %v", err)
			cfg.Tray.Enabled = false
		} else {
			// Set up callbacks for mount/unmount operations
			trayIcon.SetMountCallback(func(dev *device.Device) error {
				return d.MountDevice(dev)
			})
			trayIcon.SetUnmountCallback(func(dev *device.Device) error {
				return d.UnmountDevice(dev)
			})

			// Set up device changed callback to immediately update tray
			d.SetDeviceChangedCallback(func() {
				if trayIcon != nil {
					trayIcon.UpdateDevices()
				}
			})

			// Set up quit callback for proper cleanup
			trayIcon.SetQuitCallback(func() {
				log.Println("Quit requested from tray icon")
				if trayStopChan != nil {
					close(trayStopChan)
				}
				trayIcon.Close()
				d.Stop()
				os.Exit(0)
			})

			// Update tray with current devices periodically
			trayStopChan = make(chan struct{})
			go func() {
				ticker := time.NewTicker(5 * time.Second)
				defer ticker.Stop()

				for {
					select {
					case <-ticker.C:
						if trayIcon != nil {
							trayIcon.UpdateDevices()
						}
					case <-trayStopChan:
						return
					}
				}
			}()
		}
	}

	// Start daemon
	if err := d.Start(); err != nil {
		log.Fatalf("Failed to start daemon: %v", err)
	}

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("pgmountd daemon started. Press Ctrl+C to stop.")

	// Wait for signals
	<-sigChan

	log.Println("Shutting down...")

	// Cleanup
	if trayStopChan != nil {
		close(trayStopChan)
	}
	if trayIcon != nil {
		trayIcon.Close()
	}
	d.Stop()

	log.Println("pgmountd stopped")
}

func loadConfig() (*config.Config, error) {
	if *noConfig {
		return config.Default(), nil
	}

	path := *configFile
	if path == "" {
		// Use default config path
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		path = homeDir + "/.config/pgmount/config.yml"
	}

	// Check if config file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Printf("Config file not found at %s, using defaults", path)
		return config.Default(), nil
	}

	return config.Load(path)
}

func applyFlags(cfg *config.Config) {
	if *noAutomount {
		cfg.Automount = false
	} else if flag.Lookup("automount").Value.String() == "true" {
		cfg.Automount = *automount
	}

	if *noNotify {
		cfg.Notifications.Enabled = false
	} else if flag.Lookup("notify").Value.String() == "true" {
		cfg.Notifications.Enabled = *notifications
	}

	if *noTray {
		cfg.Tray.Enabled = false
	} else if *showTray {
		cfg.Tray.Enabled = true
		cfg.Tray.AutoHide = false
	} else if *autoTray {
		cfg.Tray.Enabled = true
		cfg.Tray.AutoHide = true
	}

	if *verbose {
		cfg.Verbose = true
	}
	if *quiet {
		cfg.Quiet = true
	}
}

func initLogger(cfg *config.Config) {
	if cfg.Quiet {
		log.SetOutput(os.NewFile(0, os.DevNull))
	}

	if cfg.Verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	} else {
		log.SetFlags(log.LstdFlags)
	}
}
