# Changelog

All notable changes to PGMount will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-11-17

### Added
- Initial release of PGMount
- Automatic mounting of removable media
- Desktop notifications using notify-send
- Command-line utilities (pgmount, pgumount, pginfo)
- GELI encryption support for encrypted devices
- YAML configuration file support
- Per-device configuration (mount options, automount behavior, ignore list)
- Event hooks for custom automation
- Filesystem support: FAT, NTFS, ext2/3/4, UFS, ZFS
- FreeBSD-native device detection using geom, gpart, camcontrol
- Polling-based device monitoring (2-second interval)
- File manager integration
- System tray icon (stub implementation)
- Comprehensive documentation (README, INSTALL, DESIGN, QUICKSTART)
- Man pages for all commands
- Unit tests for configuration module
- Contributing guidelines
- BSD 2-Clause license

### Technical Details
- Written in Go 1.21+
- No CGo dependencies for easier building
- Command-line notification system (notify-send wrapper)
- Pure Go implementation for portability

## [Unreleased]

### Planned for v1.1
- Full GTK tray icon implementation with gotk3
- devd socket monitoring for real-time events
- Password caching for GELI devices
- Improved filesystem detection
- Additional unit and integration tests

### Planned for v1.2
- Loop device support (mdconfig)
- Network filesystem support (NFS, SMB)
- GUI configuration tool
- Multiple configuration profiles
- Notification actions (click to open, mount/unmount buttons)

### Planned for v2.0
- Cross-platform BSD support (OpenBSD, NetBSD)
- Plugin system for custom filesystems
- DBus service for desktop integration
- Advanced GELI features (multiple keyfiles, TPM)

---

## Version History

### [1.0.0] - 2025-11-17
First public release of PGMount, providing udiskie-like functionality for PGSD/FreeBSD/GhostBSD.
