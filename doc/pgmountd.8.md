% PGMOUNTD(8) PGMount 1.0.0
% Pacific Grove Software Distribution Foundation
% November 2025

# NAME

pgmountd - Automounter daemon for removable media on PGSD/FreeBSD/GhostBSD

# SYNOPSIS

**pgmountd** [*OPTIONS*]

# DESCRIPTION

pgmountd is a daemon that automatically mounts removable media such as USB drives, external hard drives, and other storage devices when they are inserted. It provides desktop notifications and can integrate with system tray icons.

# OPTIONS

**-version**
:   Show version information and exit

**--config** *FILE*
:   Specify configuration file (default: ~/.config/pgmount/config.yml)

**--no-config**
:   Don't use any configuration file, use defaults

**--automount**
:   Enable automounting new devices (default: true)

**--no-automount**
:   Disable automounting new devices

**--notify**
:   Enable pop-up notifications (default: true)

**--no-notify**
:   Disable pop-up notifications

**--tray**
:   Show tray icon

**--auto-tray**
:   Show tray icon that auto-hides when no devices available

**--no-tray**
:   Disable tray icon (default)

**--mount-all**
:   Mount all available devices on startup

**--daemon**
:   Run as daemon (default: true)

**--verbose**
:   Verbose output

**--quiet**
:   Quiet output (suppress non-error messages)

# CONFIGURATION

The configuration file is located at **~/.config/pgmount/config.yml** by default. It uses YAML format.

Example configuration:

```yaml
automount: true
mount_base: /media
file_manager: xdg-open

notifications:
  enabled: true
  device_mounted: 5.0

device_config:
  - id_label: "MY_USB"
    options: [noexec, nosuid]
```

See **/usr/local/share/examples/pgmount/config.example.yml** for a complete example.

# FILES

*~/.config/pgmount/config.yml*
:   User configuration file

*/usr/local/share/examples/pgmount/config.example.yml*
:   Example configuration file

*/media*
:   Default mount base directory

# ENVIRONMENT

**DISPLAY** or **WAYLAND_DISPLAY**
:   Required for desktop notifications and tray icon

# EXAMPLES

Start pgmountd with notifications:

    pgmountd --notify &

Start with tray icon that auto-hides:

    pgmountd --auto-tray &

Use custom configuration:

    pgmountd --config /path/to/config.yml

Mount all devices and exit:

    pgmountd --mount-all --daemon=false

# SEE ALSO

**pgmount**(8), **pgumount**(8), **pginfo**(8), **mount**(8), **geli**(8)

# BUGS

Report bugs to: https://github.com/pgsdf/pgmount/issues

# COPYRIGHT

Copyright © 2025 Pacific Grove Software Distribution Foundation. BSD 2-Clause License.
