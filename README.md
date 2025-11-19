# PGMount

A modern automounter for removable media on PGSD/FreeBSD/GhostBSD.

## Requirements

### Runtime Dependencies

- PGSD or FreeBSD 14.0+ or GhostBSD 25.01+
- libnotify 

### Build Dependencies

- Go 1.21 or later

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/pgsdf/pgmount.git
cd pgmount

# Install dependencies
pkg install go gtk3 libnotify pkgconf

# Build (automatically downloads Go modules)
make

# Install
sudo make install
```

### Package Installation

```bash
# Coming soon: pkg install pgmount
```

## Usage

### Running the Daemon

Start pgmountd as a daemon with automounting enabled:

```bash
pgmountd
```

Common options:

```bash
# Run with notifications and tray icon
pgmountd --notify --tray

# Run with auto-hiding tray icon
pgmountd --auto-tray

# Disable automounting
pgmountd --no-automount

# Use custom config file
pgmountd --config /path/to/config.yml

# Mount all devices on startup
pgmountd --mount-all
```

### Autostart

To start pgmountd automatically, add it to your window manager's autostart:

**For Openbox** (`~/.config/openbox/autostart`):
```bash
pgmountd --tray &
```

**For MATE** (`~/.config/autostart/pgmount.desktop`):
```ini
[Desktop Entry]
Type=Application
Name=PGMount
Exec=pgmountd --tray
X-GNOME-Autostart-enabled=true
```

### Manual Mounting

Mount a specific device:

```bash
# Mount by device path
pgmount /dev/da0p1

# Mount with specific filesystem type
pgmount -t msdosfs /dev/da0p1

# Mount with options
pgmount -o nosuid,noexec /dev/da0p1

# Mount all available devices
pgmount -a
```

### Manual Unmounting

Unmount a device:

```bash
# Unmount by device path
pgumount /dev/da0p1

# Unmount by mount point
pgumount /media/USB_DRIVE

# Unmount and detach (safe removal)
pgumount --detach /dev/da0p1

# Force unmount
pgumount -f /dev/da0p1

# Unmount all
pgumount -a
```

### Listing Devices

List all removable devices:

```bash
# Basic list
pginfo

# Show all devices (including non-partitions)
pginfo -a

# Verbose output with details
pginfo -v
```

## Configuration

PGMount uses a YAML configuration file located at `~/.config/pgmount/config.yml`.

### Example Configuration

```yaml
# Enable/disable automounting
automount: true

# Verbose logging
verbose: false

# Base directory for mount points
mount_base: /media

# File manager to open mounted directories
file_manager: xdg-open

# Notification settings
notifications:
  enabled: true
  timeout: 1.5  # Default timeout in seconds
  device_mounted: 5.0    # Mount notification timeout
  device_unmounted: -1   # -1 uses default timeout
  device_added: -1
  device_removed: -1
  device_unlocked: 3.0
  device_locked: -1
  job_failed: 10.0       # Error notification timeout

# Tray icon settings
tray:
  enabled: true
  auto_hide: true  # Auto-hide when no devices available
  icon_name: drive-removable-media

# GELI encryption settings
geli:
  enabled: true
  password_cmd: ""  # Custom password prompt command
  cache_timeout: 0  # Password cache timeout (0 = disabled)
  keyfiles:
    # Map device UUID to keyfile path
    "12345678-1234-1234-1234-123456789abc": "/path/to/keyfile"

# Per-device configuration
device_config:
  # Configuration for specific device by UUID
  - id_uuid: "ABCD-1234"
    automount: false
    ignore: false
    options:
      - noexec
      - nosuid

  # Configuration by label
  - id_label: "MY_USB"
    automount: true
    options:
      - locale=en_US.UTF-8

  # Configuration by device path
  - device_path: "/dev/da0p1"
    ignore: true  # Never mount this device

# Default mount options by filesystem type
mount_options:
  default:
    vfat:
      - locale=en_US.UTF-8
      - longnames
    ntfs:
      - locale=en_US.UTF-8
    ext2: []
    ext3: []
    ext4: []
    ufs: []
    zfs: []
    msdos:
      - locale=en_US.UTF-8
      - longnames

# Event hooks - execute commands on device events
event_hooks:
  device_added: "echo 'Device {device} added' >> /tmp/pgmount.log"
  device_mounted: "notify-send 'Mounted' '{label} mounted at {mount_point}'"
  device_unmounted: "echo 'Unmounted {device}'"
```

### Configuration Variables

Event hooks support the following variables:

- `{device}` - Device path (e.g., `/dev/da0p1`)
- `{label}` - Device label
- `{uuid}` - Device UUID
- `{mount_point}` - Mount point path

## Filesystem Support

PGMount supports the following filesystems:

| Filesystem | FreeBSD Mount Type | Notes |
|------------|-------------------|-------|
| FAT12/16/32 | `msdosfs` | Full support |
| NTFS | `ntfs` | Read-only by default, use ntfs-3g for write |
| ext2/3/4 | `ext2fs` | Requires ext2fs kernel module |
| UFS | `ufs` | Native FreeBSD filesystem |
| ZFS | `zfs` | Native FreeBSD filesystem |
| exFAT | `exfat` | Requires exfat-utils |

## GELI Encryption

PGMount supports GELI-encrypted devices (FreeBSD equivalent of LUKS):

### Unlocking with Password

Devices will prompt for password when inserted:

```bash
# Daemon will automatically prompt
pgmount
```

### Unlocking with Keyfile

Configure keyfiles in `config.yml`:

```yaml
geli:
  enabled: true
  keyfiles:
    "device-uuid": "/path/to/keyfile"
```

### Manual Unlock

```bash
# Unlock manually
geli attach -k /path/to/keyfile /dev/da0p1

# Then mount
pgmountd /dev/da0p1.eli
```

## Troubleshooting

### Device Not Auto-Mounting

1. Check if automounting is enabled:
   ```bash
   pginfo -v
   ```

2. Verify device is not in ignore list (config.yml)

3. Check logs for errors:
   ```bash
   pgmountd --verbose
   ```

### Permission Denied Errors

Ensure your user is in the `operator` group:

```bash
sudo pw groupmod operator -m $USER
```

Then log out and back in.

### Notifications Not Working

1. Ensure libnotify is installed:
   ```bash
   pkg install libnotify
   ```

2. Test notifications:
   ```bash
   notify-send "Test" "Testing notifications"
   ```

### Tray Icon Not Appearing

The tray icon requires GTK+3. Currently, the tray implementation is a stub. Full GTK integration is planned for future releases.

## Comparison with udiskie

| Feature | udiskie (Linux) | pgmountd (FreeBSD) |
|---------|----------------|----------------------|
| Automounting | ✓ (udisks2) | ✓ (native FreeBSD) |
| Notifications | ✓ | ✓ |
| Tray Icon | ✓ | ⚠ (planned) |
| CLI Tools | ✓ | ✓ |
| Encryption | ✓ (LUKS) | ✓ (GELI) |
| Configuration | ✓ (YAML) | ✓ (YAML) |
| Event Hooks | ✓ | ✓ |
| Loop Devices | ✓ | ⚠ (planned) |

## Development

### Building from Source

```bash
# Get dependencies
go mod download

# Run tests
go test ./...

# Build all binaries
make all

# Install
make install
```

### Project Structure

```
pgmount/
├── main.go              # Main daemon
├── config/              # Configuration handling
│   └── config.go
├── device/              # Device detection and management
│   └── device.go
├── daemon/              # Automount daemon
│   └── daemon.go
├── notify/              # Desktop notifications
│   └── notify.go
├── tray/                # System tray icon
│   └── tray.go
└── cmd/                 # Command-line utilities
    ├── pgmount/
    ├── pgumount/
    └── pginfo/
```

## License

BSD 2-Clause License

Copyright (c) 2025, Pacific Grove Software Distribution Foundation  
Author: Vester (Vic) Thacker  
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this
   list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
   this list of conditions and the following disclaimer in the documentation
   and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

## Contact and Support

- **Telegram**: https://t.me/PGSD_Foundation
- **Issues**: https://github.com/pgsdf/pgmount/issues
