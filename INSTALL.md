# Installation Guide

## Quick Start

### FreeBSD

```bash
# Install dependencies
pkg install go libnotify

# Clone repository
git clone https://github.com/pgsdf/pgmount.git
cd pgmount

# Build and install
make
sudo make install

# Create config directory
mkdir -p ~/.config/pgmount
cp /usr/local/share/examples/pgmount/config.example.yml \
   ~/.config/pgmount/config.yml
```

### PGSD and GhostBSD

GhostBSD will include pgmountd in the package repository:

```bash
# Install from package (coming soon)
pkg install pgmount

# Create config
mkdir -p ~/.config/pgmount
cp /usr/local/share/examples/pgmount/config.example.yml \
   ~/.config/pgmount/config.yml
```

## Dependencies

### Runtime Requirements

- **PGSD** or **FreeBSD 14.0+** or **GhostBSD**
- **libnotify** - provides `notify-send` command for notifications

### Build Requirements

- **Go 1.21+** - Go programming language
- **make** - build automation

### Optional Dependencies

- **dumpe2fs** (e2fsprogs) - for ext2/3/4 filesystem detection
- **ntfs-3g** - for NTFS write support
- **exfat-utils** - for exFAT support
- **fusefs-lkl** - for better Linux filesystem support

## Installation Methods

### Method 1: From Source (Recommended for Development)

```bash
# Install build dependencies
pkg install go libnotify

# Clone the repository
git clone https://github.com/pgsdf/pgmount.git
cd pgmount

# Build (automatically downloads dependencies)
make

# Test (optional)
make test

# Install system-wide
sudo make install

# Or install to custom location
make PREFIX=/opt/pgmount install
```

### Method 2: Using Go Install

```bash
# Install directly from source
go install github.com/pgsdf/pgmount@latest
go install github.com/pgsdf/pgmount/cmd/pgmount@latest
go install github.com/pgsdf/pgmount/cmd/pgumount@latest
go install github.com/pgsdf/pgmount/cmd/pginfo@latest

# Binaries will be in ~/go/bin/
# Add to PATH if needed:
export PATH="$PATH:$HOME/go/bin"
```

### Method 3: Package Installation (Future)

```bash
# Once available in PGSD or FreeBSD or GhostBSD ports
pkg install pgmount

# Or from ports
cd /usr/ports/sysutils/pgmount
make install clean
```

## Post-Installation Setup

### 1. User Permissions

Add your user to the `operator` group to allow mounting:

```bash
sudo pw groupmod operator -m $USER
```

**Important:** Log out and log back in for group changes to take effect.

### 2. Enable User Mounting

Enable the `vfs.usermount` sysctl:

```bash
# Temporary (until reboot)
sudo sysctl vfs.usermount=1

# Permanent (add to /etc/sysctl.conf)
echo 'vfs.usermount=1' | sudo tee -a /etc/sysctl.conf
```

### 3. Create Mount Directory

```bash
# Create base mount directory
sudo mkdir -p /media
sudo chmod 1777 /media  # Sticky bit for security
```

### 4. Configuration File

Create your configuration file:

```bash
# Create config directory
mkdir -p ~/.config/pgmount

# Copy example config
cp /usr/local/share/examples/pgmount/config.example.yml \
   ~/.config/pgmount/config.yml

# Edit as needed
${EDITOR:-vi} ~/.config/pgmount/config.yml
```

### 5. Autostart (Optional)

#### For MATE Desktop

Create `~/.config/autostart/pgmount.desktop`:

```ini
[Desktop Entry]
Type=Application
Name=PGMount
Comment=Automounter for removable media
Exec=pgmount --tray
X-GNOME-Autostart-enabled=true
Hidden=false
NoDisplay=false
```

#### For Openbox

Add to `~/.config/openbox/autostart`:

```bash
pgmountd --tray &
```

#### For i3

Add to `~/.config/i3/config`:

```
exec --no-startup-id pgmount
```

## Verification

Test that everything is working:

```bash
# Check version
pgmountd --version

# List devices
pginfo

# Test daemon (in foreground)
pgmountd --verbose

# In another terminal, insert a USB drive and watch the output
```

## Troubleshooting Installation

### Issue: "command not found: go"

Install Go:

```bash
pkg install go
```

### Issue: Notifications don't work

Install libnotify (provides notify-send):

```bash
pkg install libnotify
```

### Issue: Build fails with dependency errors

Download Go modules:

```bash
go mod download
```

### Issue: "Permission denied" when mounting

1. Check group membership:
   ```bash
   groups $USER
   # Should include 'operator'
   ```

2. If not in operator group:
   ```bash
   sudo pw groupmod operator -m $USER
   # Log out and back in
   ```

3. Check vfs.usermount:
   ```bash
   sysctl vfs.usermount
   # Should be 1
   ```

### Issue: Notifications don't appear

1. Test libnotify:
   ```bash
   notify-send "Test" "Testing notifications"
   ```

2. Check desktop environment supports notifications

3. Verify notification daemon is running:
   ```bash
   ps aux | grep -i notify
   ```

## Uninstallation

```bash
# Remove binaries
sudo make uninstall

# Remove configuration (optional)
rm -rf ~/.config/pgmount

# Remove Go dependencies (optional)
go clean -modcache
```

## Building from Git

For the latest development version:

```bash
# Clone repository
git clone https://github.com/pgsdf/pgmount.git
cd pgmount

# Checkout development branch
git checkout develop

# Build
make

# Install
sudo make install
```

## Building for Distribution

To create a distributable package:

```bash
# Build with version information
make VERSION=1.0.0

# Create tarball
make dist

# This creates: pgmount-1.0.0.tar.gz
```

## Next Steps

After installation:

1. Read the [README.md](README.md) for usage information
2. See [DESIGN.md](DESIGN.md) for architecture details
3. Check example configuration in `config.example.yml`
4. Run `pgmountd --help` for command-line options

## Getting Help

If you encounter issues:

1. Check the [Troubleshooting](#troubleshooting-installation) section above
2. Review the [README.md](README.md) troubleshooting section
3. Search existing [GitHub issues](https://github.com/pgsdf/pgmount/issues)
4. Create a new issue with:
   - PGSD/FreeBSD/GhostBSD version
   - Go version (`go version`)
   - Full error message
   - Steps to reproduce
