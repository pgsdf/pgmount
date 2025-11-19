# TODO List

## Version 1.0 (Initial Release)

- [x] Core device detection
- [x] Basic automounting
- [x] Desktop notifications
- [x] Configuration file support
- [x] CLI tools (mount/umount/info)
- [x] GELI encryption support
- [x] Event hooks
- [x] Improved tray icon implementation (stub with menu structure)
- [x] Basic test suite for config module
- [x] Man pages for all commands
- [x] Contributing guidelines
- [ ] FreeBSD port/package (pending)

## Version 1.1 (Enhancements)

### High Priority

- [ ] **devd Integration**
  - [ ] Replace polling with devd socket monitoring
  - [ ] Parse devd event messages
  - [ ] Handle attach/detach events in real-time

- [ ] **GTK Tray Icon**
  - [ ] Implement using gotk3 bindings
  - [ ] Device list menu
  - [ ] Mount/unmount actions
  - [ ] Auto-hide when no devices
  - [ ] Custom icons per device type

- [ ] **Password Caching**
  - [ ] Implement secure password cache
  - [ ] Timeout-based expiration
  - [ ] Per-device password storage
  - [ ] Keyring integration (optional)

- [ ] **Better Filesystem Detection**
  - [ ] Use libblkid (if available)
  - [ ] Better ext2/3/4 detection
  - [ ] exFAT support
  - [ ] APFS detection (future)

### Medium Priority

- [ ] **Loop Device Support**
  - [ ] mdconfig integration
  - [ ] ISO mounting
  - [ ] Auto-unmount and detach

- [ ] **Logging Improvements**
  - [ ] Structured logging
  - [ ] Log rotation
  - [ ] Syslog integration
  - [ ] Debug vs normal modes

- [ ] **Performance Optimizations**
  - [ ] Cache device information
  - [ ] Async device scanning
  - [ ] Reduce syscalls

- [ ] **Documentation**
  - [ ] Man pages for all commands
  - [ ] Architecture diagrams
  - [ ] Video tutorials
  - [ ] FAQ expansion

### Low Priority

- [ ] **Internationalization**
  - [ ] i18n framework
  - [ ] Translation strings
  - [ ] Localized notifications

- [ ] **Additional Filesystems**
  - [ ] F2FS support
  - [ ] Btrfs support (via FUSE)
  - [ ] ReiserFS support

## Version 1.2 (Advanced Features)

### Network Filesystems

- [ ] **SMB/CIFS Support**
  - [ ] Auto-mount SMB shares
  - [ ] Credential management
  - [ ] Discovery via Avahi/mDNS

- [ ] **NFS Support**
  - [ ] Auto-mount NFS exports
  - [ ] v3 and v4 support
  - [ ] Kerberos authentication

- [ ] **SSH/SFTP Support**
  - [ ] SSHFS integration
  - [ ] Key-based auth
  - [ ] Connection pooling

### Configuration

- [ ] **Multiple Profiles**
  - [ ] Work/Home profiles
  - [ ] Profile switching
  - [ ] Per-profile settings

- [ ] **GUI Configuration Tool**
  - [ ] GTK application
  - [ ] Device whitelist/blacklist
  - [ ] Mount options editor
  - [ ] Event hook builder

### Notifications

- [ ] **Notification Actions**
  - [ ] Click to open file manager
  - [ ] Mount/unmount buttons
  - [ ] Custom actions
  - [ ] Action feedback

- [ ] **Progress Notifications**
  - [ ] Mount progress
  - [ ] Large file copy detection
  - [ ] Estimated time remaining

## Version 2.0 (Major Features)

### Cross-Platform BSD Support

- [ ] **OpenBSD Support**
  - [ ] Port device detection
  - [ ] Adapt to OpenBSD tools
  - [ ] Package for OpenBSD

- [ ] **NetBSD Support**
  - [ ] Port device detection
  - [ ] Adapt to NetBSD tools
  - [ ] Package for NetBSD

- [ ] **DragonFly BSD Support**
  - [ ] Port device detection
  - [ ] HAMMER filesystem support

### Plugin System

- [ ] **Plugin Architecture**
  - [ ] Plugin API design
  - [ ] Filesystem plugins
  - [ ] Notification plugins
  - [ ] Mount strategy plugins

- [ ] **Community Plugins**
  - [ ] Plugin repository
  - [ ] Plugin manager
  - [ ] Security sandboxing

### DBus Service

- [ ] **DBus Interface**
  - [ ] Implement DBus service
  - [ ] Standard mount interface
  - [ ] Desktop integration
  - [ ] File manager integration

- [ ] **Desktop Integration**
  - [ ] GNOME Files integration
  - [ ] KDE Dolphin integration
  - [ ] PCManFM integration

### Advanced Encryption

- [ ] **Multiple Encryption Methods**
  - [ ] GELI with detached metadata
  - [ ] Multiple keyfile support
  - [ ] TPM integration (if available)
  - [ ] Smart card support

- [ ] **Container Formats**
  - [ ] TrueCrypt/VeraCrypt containers
  - [ ] LUKS containers (via FUSE)
  - [ ] Encrypted disk images

## Wishlist (No Specific Version)

### Nice to Have

- [ ] Web interface for remote management
- [ ] Mobile app for notifications
- [ ] Cloud storage integration (Nextcloud, etc.)
- [ ] Backup automation on device insertion
- [ ] Smart mounting based on usage patterns
- [ ] Battery-aware operations (laptops)
- [ ] Bandwidth limiting for network mounts
- [ ] Snapshot support for ZFS devices
- [ ] Automatic filesystem repair
- [ ] Usage statistics and reports

### Research Items

- [ ] Machine learning for mount predictions
- [ ] Blockchain for device trust verification (?)
- [ ] Integration with FreeBSD jails
- [ ] Bhyve VM disk passthrough
- [ ] RDMA support for network filesystems
- [ ] GPU acceleration for encryption

## Code Quality

### Testing

- [ ] **Unit Tests**
  - [ ] Config parsing tests
  - [ ] Device detection tests
  - [ ] Mount logic tests
  - [ ] 80%+ code coverage

- [ ] **Integration Tests**
  - [ ] Full mount/unmount cycle
  - [ ] Multi-device scenarios
  - [ ] Error handling
  - [ ] Race condition tests

- [ ] **Performance Tests**
  - [ ] Benchmark device scanning
  - [ ] Benchmark mount operations
  - [ ] Memory leak detection
  - [ ] Long-running stability tests

### Code Improvements

- [ ] **Refactoring**
  - [ ] Reduce code duplication
  - [ ] Improve error handling
  - [ ] Better abstractions
  - [ ] Interface definitions

- [ ] **Documentation**
  - [ ] Inline code documentation
  - [ ] Package documentation
  - [ ] API documentation
  - [ ] Architecture documentation

- [ ] **CI/CD**
  - [ ] GitHub Actions setup
  - [ ] Automated testing
  - [ ] Automated releases
  - [ ] Code quality checks

## Community

- [ ] **Website**
  - [ ] Project website
  - [ ] Documentation site
  - [ ] Download page
  - [ ] Blog for updates

- [ ] **Community Building**
  - [ ] Discord/IRC channel
  - [ ] Forum or mailing list
  - [ ] Contributing guidelines
  - [ ] Code of conduct

- [ ] **Ecosystem**
  - [ ] Example configurations repository
  - [ ] Plugin development guide
  - [ ] Third-party integrations
  - [ ] Related tools

## Bugs & Known Issues

### Current Issues

- [ ] Tray icon not implemented
- [ ] Device polling has 2-second latency
- [ ] Some filesystems require manual module loading
- [ ] GELI password not cached

### Reported Issues

(To be filled in as issues are reported)

## Notes

- Items marked [x] are completed
- Items marked [ ] are pending
- Priority can change based on user feedback
- Version numbers are tentative
- Community contributions welcome for any item!

## Contributing

Want to help? Pick an item from the TODO list and:

1. Create an issue to discuss the feature
2. Fork the repository
3. Implement the feature
4. Submit a pull request

See [CONTRIBUTING.md](CONTRIBUTING.md) for details (to be created).
