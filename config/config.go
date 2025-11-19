package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Automount     bool                `yaml:"automount"`
	Verbose       bool                `yaml:"verbose"`
	Quiet         bool                `yaml:"quiet"`
	MountBase     string              `yaml:"mount_base"`
	FileManager   string              `yaml:"file_manager"`
	Notifications NotificationConfig  `yaml:"notifications"`
	Tray          TrayConfig          `yaml:"tray"`
	Devices       []DeviceConfig      `yaml:"device_config"`
	EventHooks    map[string]string   `yaml:"event_hooks"`
	MountOptions  MountOptionsConfig  `yaml:"mount_options"`
	GELI          GELIConfig          `yaml:"geli"`
}

// NotificationConfig contains notification settings
type NotificationConfig struct {
	Enabled          bool    `yaml:"enabled"`
	Timeout          float64 `yaml:"timeout"`
	DeviceMounted    float64 `yaml:"device_mounted"`
	DeviceUnmounted  float64 `yaml:"device_unmounted"`
	DeviceAdded      float64 `yaml:"device_added"`
	DeviceRemoved    float64 `yaml:"device_removed"`
	DeviceUnlocked   float64 `yaml:"device_unlocked"`
	DeviceLocked     float64 `yaml:"device_locked"`
	JobFailed        float64 `yaml:"job_failed"`
}

// TrayConfig contains tray icon settings
type TrayConfig struct {
	Enabled  bool   `yaml:"enabled"`
	AutoHide bool   `yaml:"auto_hide"`
	IconName string `yaml:"icon_name"`
}

// DeviceConfig contains per-device configuration
type DeviceConfig struct {
	IDLabel    string   `yaml:"id_label"`
	IDUUID     string   `yaml:"id_uuid"`
	DevicePath string   `yaml:"device_path"`
	Ignore     bool     `yaml:"ignore"`
	Automount  *bool    `yaml:"automount,omitempty"`
	Options    []string `yaml:"options"`
}

// MountOptionsConfig contains default mount options
type MountOptionsConfig struct {
	Default map[string][]string `yaml:"default"`
}

// GELIConfig contains GELI encryption settings
type GELIConfig struct {
	Enabled      bool              `yaml:"enabled"`
	PasswordCmd  string            `yaml:"password_cmd"`
	CacheTimeout int               `yaml:"cache_timeout"`
	KeyFiles     map[string]string `yaml:"keyfiles"`
}

// Default returns a default configuration
func Default() *Config {
	return &Config{
		Automount:  true,
		Verbose:    false,
		Quiet:      false,
		MountBase:  "/media",
		FileManager: "xdg-open",
		Notifications: NotificationConfig{
			Enabled:         true,
			Timeout:         1.5,
			DeviceMounted:   5.0,
			DeviceUnmounted: -1.0,
			DeviceAdded:     -1.0,
			DeviceRemoved:   -1.0,
			DeviceUnlocked:  -1.0,
			DeviceLocked:    -1.0,
			JobFailed:       -1.0,
		},
		Tray: TrayConfig{
			Enabled:  false,
			AutoHide: true,
			IconName: "drive-removable-media",
		},
		Devices:    []DeviceConfig{},
		EventHooks: make(map[string]string),
		MountOptions: MountOptionsConfig{
			Default: map[string][]string{
				"vfat":  {"locale=en_US.UTF-8", "longnames"},
				"ntfs":  {"locale=en_US.UTF-8"},
				"ext2":  {},
				"ext3":  {},
				"ext4":  {},
				"ufs":   {},
				"zfs":   {},
				"msdos": {"locale=en_US.UTF-8", "longnames"},
			},
		},
		GELI: GELIConfig{
			Enabled:      true,
			PasswordCmd:  "",
			CacheTimeout: 0,
			KeyFiles:     make(map[string]string),
		},
	}
}

// Load reads and parses the configuration file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

// Save writes the configuration to a file
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetDeviceConfig returns the configuration for a specific device
func (c *Config) GetDeviceConfig(label, uuid, path string) *DeviceConfig {
	for i := range c.Devices {
		dev := &c.Devices[i]
		if dev.IDLabel != "" && dev.IDLabel == label {
			return dev
		}
		if dev.IDUUID != "" && dev.IDUUID == uuid {
			return dev
		}
		if dev.DevicePath != "" && dev.DevicePath == path {
			return dev
		}
	}
	return nil
}

// ShouldIgnoreDevice checks if a device should be ignored
func (c *Config) ShouldIgnoreDevice(label, uuid, path string) bool {
	devCfg := c.GetDeviceConfig(label, uuid, path)
	if devCfg != nil {
		return devCfg.Ignore
	}
	return false
}

// ShouldAutomountDevice checks if a device should be automounted
func (c *Config) ShouldAutomountDevice(label, uuid, path string) bool {
	devCfg := c.GetDeviceConfig(label, uuid, path)
	if devCfg != nil && devCfg.Automount != nil {
		return *devCfg.Automount
	}
	return c.Automount
}

// GetMountOptions returns mount options for a device
func (c *Config) GetMountOptions(fstype string, label, uuid, path string) []string {
	// Check device-specific options first
	devCfg := c.GetDeviceConfig(label, uuid, path)
	if devCfg != nil && len(devCfg.Options) > 0 {
		return devCfg.Options
	}

	// Return default options for filesystem type
	if opts, ok := c.MountOptions.Default[fstype]; ok {
		return opts
	}

	return []string{}
}
