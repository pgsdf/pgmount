package config

import (
	"os"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.Automount != true {
		t.Error("Default automount should be true")
	}

	if cfg.MountBase != "/media" {
		t.Errorf("Default mount base should be /media, got %s", cfg.MountBase)
	}

	if !cfg.Notifications.Enabled {
		t.Error("Default notifications should be enabled")
	}

	if cfg.Notifications.DeviceMounted != 5.0 {
		t.Errorf("Default device_mounted notification should be 5.0, got %f", cfg.Notifications.DeviceMounted)
	}
}

func TestLoadConfig(t *testing.T) {
	// Create temporary config file
	tmpfile, err := os.CreateTemp("", "ghostmount-test-*.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	configContent := `
automount: false
mount_base: /mnt/usb
file_manager: thunar

notifications:
  enabled: true
  device_mounted: 3.0

device_config:
  - id_label: "TEST_USB"
    automount: true
    options:
      - noexec
      - nosuid
`

	if _, err := tmpfile.Write([]byte(configContent)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	cfg, err := Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Automount {
		t.Error("Automount should be false")
	}

	if cfg.MountBase != "/mnt/usb" {
		t.Errorf("Mount base should be /mnt/usb, got %s", cfg.MountBase)
	}

	if cfg.FileManager != "thunar" {
		t.Errorf("File manager should be thunar, got %s", cfg.FileManager)
	}

	if cfg.Notifications.DeviceMounted != 3.0 {
		t.Errorf("Device mounted notification should be 3.0, got %f", cfg.Notifications.DeviceMounted)
	}

	if len(cfg.Devices) != 1 {
		t.Errorf("Should have 1 device config, got %d", len(cfg.Devices))
	}

	if cfg.Devices[0].IDLabel != "TEST_USB" {
		t.Errorf("Device label should be TEST_USB, got %s", cfg.Devices[0].IDLabel)
	}
}

func TestGetDeviceConfig(t *testing.T) {
	cfg := Default()
	cfg.Devices = []DeviceConfig{
		{
			IDLabel: "MY_USB",
			Options: []string{"noexec"},
		},
		{
			IDUUID: "1234-5678",
			Options: []string{"ro"},
		},
	}

	// Test finding by label
	devCfg := cfg.GetDeviceConfig("MY_USB", "", "")
	if devCfg == nil {
		t.Error("Should find device by label")
	}
	if len(devCfg.Options) != 1 || devCfg.Options[0] != "noexec" {
		t.Error("Should have correct options")
	}

	// Test finding by UUID
	devCfg = cfg.GetDeviceConfig("", "1234-5678", "")
	if devCfg == nil {
		t.Error("Should find device by UUID")
	}

	// Test not finding device
	devCfg = cfg.GetDeviceConfig("NONEXISTENT", "", "")
	if devCfg != nil {
		t.Error("Should not find nonexistent device")
	}
}

func TestShouldIgnoreDevice(t *testing.T) {
	cfg := Default()
	cfg.Devices = []DeviceConfig{
		{
			IDLabel: "IGNORE_ME",
			Ignore:  true,
		},
	}

	if !cfg.ShouldIgnoreDevice("IGNORE_ME", "", "") {
		t.Error("Should ignore device")
	}

	if cfg.ShouldIgnoreDevice("KEEP_ME", "", "") {
		t.Error("Should not ignore device")
	}
}

func TestShouldAutomountDevice(t *testing.T) {
	cfg := Default()
	cfg.Automount = true

	automountFalse := false
	automountTrue := true

	cfg.Devices = []DeviceConfig{
		{
			IDLabel:   "NO_AUTO",
			Automount: &automountFalse,
		},
		{
			IDLabel:   "YES_AUTO",
			Automount: &automountTrue,
		},
	}

	// Device with automount=false
	if cfg.ShouldAutomountDevice("NO_AUTO", "", "") {
		t.Error("Should not automount NO_AUTO")
	}

	// Device with automount=true
	if !cfg.ShouldAutomountDevice("YES_AUTO", "", "") {
		t.Error("Should automount YES_AUTO")
	}

	// Unknown device should use global setting
	if !cfg.ShouldAutomountDevice("UNKNOWN", "", "") {
		t.Error("Should use global automount setting for unknown device")
	}

	// Test with global automount=false
	cfg.Automount = false
	if cfg.ShouldAutomountDevice("UNKNOWN", "", "") {
		t.Error("Should not automount unknown device when global is false")
	}
}

func TestGetMountOptions(t *testing.T) {
	cfg := Default()
	cfg.Devices = []DeviceConfig{
		{
			IDLabel: "CUSTOM",
			Options: []string{"noexec", "nosuid"},
		},
	}

	// Test device-specific options
	opts := cfg.GetMountOptions("vfat", "CUSTOM", "", "")
	if len(opts) != 2 || opts[0] != "noexec" || opts[1] != "nosuid" {
		t.Errorf("Should get device-specific options, got %v", opts)
	}

	// Test filesystem default options
	opts = cfg.GetMountOptions("vfat", "OTHER", "", "")
	if len(opts) == 0 {
		t.Error("Should get default vfat options")
	}

	// Test unknown filesystem
	opts = cfg.GetMountOptions("unknown", "OTHER", "", "")
	if len(opts) != 0 {
		t.Error("Should return empty options for unknown filesystem")
	}
}
