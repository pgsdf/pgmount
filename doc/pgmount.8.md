% PGMOUNT(8) PGMount 1.0.0
% Pacific Grove Software Distribution Foundation
% November 2025

# NAME

pgmount - Mount removable media devices

# SYNOPSIS

**pgmount** [*OPTIONS*] [*DEVICE*]

# DESCRIPTION

pgmount is a command-line utility for mounting removable media devices. It can mount individual devices or all available devices at once.

# OPTIONS

**-a**
:   Mount all available devices

**-v**
:   Verbose output

**-t** *FSTYPE*
:   Specify filesystem type (e.g., msdosfs, ntfs, ext4)

**-o** *OPTIONS*
:   Mount options (comma-separated)

**--config** *FILE*
:   Specify configuration file

**--no-config**
:   Don't use any configuration file

# ARGUMENTS

*DEVICE*
:   Device path to mount (e.g., /dev/da0p1)

# EXAMPLES

Mount a USB drive:

    pgmount /dev/da0p1

Mount with specific filesystem type:

    pgmount -t msdosfs /dev/da0p1

Mount with options:

    pgmount -o nosuid,noexec /dev/da0p1

Mount all available devices:

    pgmount -a

# EXIT STATUS

**0**
:   Success

**1**
:   Failure

# FILES

*~/.config/pgmount/config.yml*
:   User configuration file (for mount options)

*/media*
:   Default mount base directory

# SEE ALSO

**pgmountd**(8), **pgumount**(8), **pginfo**(8), **mount**(8)

# BUGS

Report bugs to: https://github.com/pgsdf/pgmount/issues

# COPYRIGHT

Copyright © 2025 Pacific Grove Software Distribution Foundation. BSD 2-Clause License.
