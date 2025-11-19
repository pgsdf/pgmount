% PGINFO(8) PGMount 1.0.0
% Pacific Grove Software Distribution Foundation
% November 2025

# NAME

pginfo - List removable media devices

# SYNOPSIS

**pginfo** [*OPTIONS*]

# DESCRIPTION

pginfo lists all removable media devices and their status. It shows device paths, labels, filesystem types, and mount status.

# OPTIONS

**-a**
:   Show all devices (including non-removable and parent devices)

**-v**
:   Verbose output with detailed information

# OUTPUT

The default output shows:

- **DEVICE**: Device path (e.g., /dev/da0p1)
- **LABEL**: Device label or name
- **MOUNTED**: Whether the device is currently mounted (Yes/No)
- **MOUNT POINT**: Where the device is mounted (if mounted)

Verbose output additionally shows:

- **UUID**: Device UUID
- **FSTYPE**: Filesystem type
- **SIZE**: Device size
- **ENCRYPTED**: Whether the device is encrypted (GELI)

# EXAMPLES

List all removable devices:

    pginfo

Show detailed information:

    pginfo -v

Include all devices (not just removable):

    pginfo -a

Verbose output for all devices:

    pginfo -av

# EXIT STATUS

**0**
:   Success

**1**
:   Failure

# SEE ALSO

**pgmountd**(8), **pgmount**(8), **pgumount**(8), **geom**(8), **gpart**(8)

# BUGS

Report bugs to: https://github.com/pgsdf/pgmount/issues

# COPYRIGHT

Copyright © 2025 Pacific Grove Software Distibution Foundation. BSD 2-Clause License.
