# PGMount Design Document

## Overview

PGMount is a PGSD/FreeBSD/GhostBSD automounter for removable media, inspired by udiskie. This document explains the design decisions, architecture, and how it adapts udiskie's concepts to PGSD.

## Design Goals

1. **Native PGSD Integration** - Use FreeBSD's native tools and APIs
2. **Performance** - Fast device detection and mounting using Go
3. **Simplicity** - Clean, maintainable codebase
4. **Compatibility** - Match udiskie's feature set where applicable
5. **Extensibility** - Easy to add new features and filesystem support

## Architecture

### System Architecture

```
┌─────────────────────────────────────────────────────┐
│                  User Interface                     │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐          │
│  │  Tray    │  │  CLI     │  │  Notify  │          │
│  │  Icon    │  │  Tools   │  │  System  │          │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘          │
└───────┼─────────────┼─────────────┼─────────────────┘
        │             │             │
        ▼             ▼             ▼
┌─────────────────────────────────────────────────────┐
│                    Daemon Core                      │
│  ┌────────────────────────────────────────────┐    │
│  │  Event Handler & Automount Logic           │    │
│  └────────────────────────────────────────────┘    │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐          │
│  │  Config  │  │  Device  │  │  Mount   │          │
│  │  Manager │  │  Manager │  │  Manager │          │
│  └──────────┘  └──────────┘  └──────────┘          │
└───────┬─────────────┬─────────────┬─────────────────┘
        │             │             │
        ▼             ▼             ▼
┌─────────────────────────────────────────────────────┐
│              PGSD System Layer                     │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐          │
│  │   devd   │  │  geom/   │  │  mount/  │          │
│  │  events  │  │  gpart   │  │  geli    │          │
│  └──────────┘  └──────────┘  └──────────┘          │
└─────────────────────────────────────────────────────┘
```

### Components

#### 1. Device Manager (`device/device.go`)

**Responsibilities:**
- Detect removable devices using `geom` and `camcontrol`
- Parse device metadata (label, UUID, filesystem type)
- Track device state (mounted, encrypted, etc.)

**FreeBSD-specific implementations:**
- Uses `geom disk list` to enumerate disks
- Uses `gpart show` to list partitions
- Uses `file -s` and `dumpe2fs` for filesystem detection
- Uses `glabel status` for device labels
- Checks `camcontrol devlist` for USB devices

**Difference from udiskie:**
- udiskie uses udisks2 DBus API for all device information
- PGMount directly calls FreeBSD command-line tools

#### 2. Daemon (`daemon/daemon.go`)

**Responsibilities:**
- Monitor for device events
- Handle automounting/unmounting
- Execute event hooks
- Manage GELI-encrypted devices

**Event Detection:**

udiskie approach (Linux):
```python
# Connects to udisks2 DBus signals
bus.add_signal_receiver(
    on_device_added,
    signal_name='InterfacesAdded',
    dbus_interface='org.freedesktop.DBus.ObjectManager'
)
```

PGMount approach (FreeBSD):
```go
// Option 1: Poll device state (current implementation)
ticker := time.NewTicker(2 * time.Second)
devices := deviceMgr.Scan()  // Detect changes

// Option 2: Monitor devd socket (future enhancement)
conn := net.Dial("unix", "/var/run/devd.seqpacket.pipe")
// Parse devd events
```

**Why polling instead of devd?**
- Simpler initial implementation
- More portable across PGSD versions
- devd integration planned for v2.0
- Performance impact is minimal (2-second interval)

#### 3. Configuration (`config/config.go`)

**Format:** YAML (same as udiskie)

**Compatibility:**
- Similar structure to udiskie's config
- Adapted for PGSD mount options
- GELI settings instead of LUKS
- Native PGSD filesystem types

Example comparison:

udiskie (Linux):
```yaml
device_config:
  - id_uuid: "ABCD-1234"
    options: [noexec, nosuid]
```

PGMount (FreeBSD):
```yaml
device_config:
  - id_uuid: "ABCD-1234"
    options: [noexec, nosuid]  # Same syntax!
mount_options:
  default:
    msdosfs: [locale=en_US.UTF-8]  # FreeBSD-specific
```

#### 4. Notifications (`notify/notify.go`)

**Implementation:** Command-line wrapper for `notify-send`

**Why command-line instead of CGo?**
- Simpler build process - no pkg-config or development headers needed
- No CGo compilation complexity
- Same functionality as CGo approach
- Works with all desktop environments that support libnotify
- Easier to cross-compile

**Comparison:**

udiskie:
```python
import gi
gi.require_version('Notify', '0.7')
from gi.repository import Notify

Notify.init("udiskie")
notification = Notify.Notification.new(summary, body, icon)
notification.show()
```

PGMount:
```go
import "os/exec"

func Send(summary, body, icon string, timeout int) error {
    args := []string{"-t", strconv.Itoa(timeout), "-i", icon, summary, body}
    return exec.Command("notify-send", args...).Run()
}
```

#### 5. Tray Icon (`tray/tray.go`)

**Status:** Stub implementation (planned for full GTK integration)

**Planned implementation:**
- Use `gotk3` library for GTK+3 bindings
- Create StatusIcon or AppIndicator
- Menu with device list and actions
- Auto-hide when no devices available

**Challenges:**
- GTK bindings in Go are less mature than Python
- Need to ensure compatibility with different desktop environments
- May require goroutine for GTK main loop

## Key Differences from udiskie

### 1. Backend System

| Aspect | udiskie (Linux) | PGMount (FreeBSD) |
|--------|----------------|----------------------|
| Device detection | udisks2 DBus | geom/camcontrol |
| Event system | DBus signals | devd/polling |
| Mount system | udisks2 | mount(8) |
| Encryption | LUKS via udisks2 | GELI via geli(8) |

### 2. Filesystem Support

**Linux (udiskie):**
- Relies on udisks2's filesystem support
- Automatically handles various Linux filesystems

**FreeBSD (PGMount):**
- Must explicitly handle each filesystem type
- Some filesystems require kernel modules (ext2fs, etc.)
- Different mount options and syntax

### 3. Permissions

**Linux:**
- udisks2 uses PolicyKit for permissions
- Users in specific groups can mount

**FreeBSD:**
- Requires user in `operator` group
- No PolicyKit, uses traditional Unix permissions
- May need vfs.usermount=1 sysctl

### 4. Encryption

**udiskie (LUKS):**
```python
# udisks2 handles LUKS automatically
device.Unlock(passphrase, options)
device.Mount(options)
```

**PGMount (GELI):**
```go
// Manual GELI handling
exec.Command("geli", "attach", device)
exec.Command("mount", device + ".eli", mountpoint)
```

## Implementation Decisions

### Why Go?

1. **Performance:** Compiled binary, fast startup
2. **Concurrency:** Goroutines for event handling
3. **Static binary:** Easy deployment
4. **Cross-platform:** Could support other BSDs
5. **Type safety:** Fewer runtime errors

### Why Not Python (like udiskie)?

While Python would allow code reuse from udiskie:

**Cons:**
- udisks2 doesn't exist on FreeBSD
- Python DBus bindings less critical on FreeBSD
- Go's concurrency better for event handling
- Static binary easier for embedded systems

**Pros:**
- Could reuse some logic structure
- Easier GTK bindings (PyGObject)

### Device Detection Strategy

**Considered approaches:**

1. **devd socket monitoring** (ideal)
   - Pros: Real-time events, efficient
   - Cons: Complex implementation, version-specific

2. **Polling** (current)
   - Pros: Simple, reliable, portable
   - Cons: 2-second delay, some overhead

3. **Kernel events (kqueue)**
   - Pros: Real-time, efficient
   - Cons: Complex, requires /dev monitoring

**Decision:** Start with polling, migrate to devd in v2.0

### Mount Point Strategy

**Options:**

1. `/media` (chosen)
   - Standard on many systems
   - User-writable
   - Expected location

2. `/mnt`
   - Traditional FreeBSD
   - Often requires root

3. `/run/media/$USER`
   - Modern Linux standard
   - Good isolation
   - Not standard on FreeBSD

**Decision:** `/media` for compatibility and ease of use

## Future Enhancements

### Version 1.1
- [ ] Full GTK tray icon implementation
- [ ] devd socket monitoring
- [ ] Password caching for GELI
- [ ] Better filesystem detection

### Version 1.2
- [ ] Loop device support (mdconfig)
- [ ] Remote filesystem support (NFS, SMB)
- [ ] Multiple configuration profiles
- [ ] GUI configuration tool

### Version 2.0
- [ ] Port to other BSDs (OpenBSD, NetBSD)
- [ ] Plugin system for custom filesystems
- [ ] DBus service (for desktop integration)
- [ ] Advanced notification actions

## Testing Strategy

### Unit Tests
- Device detection logic
- Configuration parsing
- Mount option handling

### Integration Tests
- Full mount/unmount cycle
- GELI encryption workflow
- Event hook execution

### Manual Tests
- Various USB drives
- Different filesystems
- Encrypted devices
- Multiple devices simultaneously

## Performance Considerations

### Memory Usage
- Go runtime: ~10MB
- Per-device tracking: ~1KB
- Total: ~15MB typical

### CPU Usage
- Polling: <1% CPU
- Event processing: negligible
- Mount operations: kernel-bound

### Disk I/O
- Device scanning: ~10 reads/scan
- Polling interval: 2 seconds
- Total: ~5 reads/second

## Security Considerations

### Mount Options
- Default to `nosuid` for removable media
- Support `noexec` for security
- Configurable per-device

### GELI Passwords
- Never stored in memory longer than needed
- Optional password caching with timeout
- Support keyfiles for automation

### Event Hooks
- Run with user permissions
- Shell injection protection via proper quoting
- Configurable via YAML only (no runtime injection)

## Conclusion

PGMount successfully adapts udiskie's excellent design to FreeBSD while maintaining:
- Similar user experience
- Compatible configuration format
- Core feature parity
- Native FreeBSD integration

The use of Go provides better performance and simpler deployment while preserving the spirit of udiskie's approach to automounting.
