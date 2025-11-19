% PGUMOUNT(8) PGMount 1.0.0
% Pacific Grove Software Distribution Foundation
% November 2024

# NAME

pgumount - Unmount removable media devices

# SYNOPSIS

**pgumount** [*OPTIONS*] [*DEVICE|MOUNTPOINT*]

# DESCRIPTION

pgumount is a command-line utility for safely unmounting removable media devices. It can unmount by device path or mount point, and optionally detach/eject the device for safe removal.

# OPTIONS

**-a**
:   Unmount all mounted removable devices

**-v**
:   Verbose output

**-f**
:   Force unmount

**--detach**
:   Also detach/eject the device after unmounting (safe removal)

# ARGUMENTS

*DEVICE|MOUNTPOINT*
:   Device path (e.g., /dev/da0p1) or mount point (e.g., /media/USB_DRIVE)

# EXAMPLES

Unmount a USB drive by device path:

    pgumount /dev/da0p1

Unmount by mount point:

    pgumount /media/MY_USB

Safe removal (unmount and detach):

    pgumount --detach /dev/da0p1

Force unmount:

    pgumount -f /dev/da0p1

Unmount all devices:

    pgumount -a

# EXIT STATUS

**0**
:   Success

**1**
:   Failure

# SEE ALSO

**pgmountd**(8), **pgmount**(8), **pginfo**(8), **umount**(8), **camcontrol**(8)

# BUGS

Report bugs to: https://github.com/pgsdf/pgmount/issues

# COPYRIGHT

Copyright © 2025 Pacific Grove Software Distribution Foundation. BSD 2-Clause License.
