
# Docker container requirements

Each image that we can use for build our dependencies

## SSH Server

- SSH server must be enabled on standard port (22)
- `permitRootLogin` must be enabled in the `sshd` configuration
- password for user `root` must be `1234`

## CMake

- CMake >= 3.21 must be installed in the system and reachable for user `root`

## Bash

- Standard `bash` utility must be installed and reachable for user `root`

## lsb_release and uname

`lsb_release` and `uname` are used to construct platform string.

`lsb_release` must support

- `-s` - short print that is easily parsable by machine
- `-r` - release version (for Debian 11 it prints "11" if used with `-s` switch)
- `-i` - distribution ID (for Debian it prints "Debian" if used with `-s` switch)

`uname` must support

- `-m` - machine. For example "x86_64"

# Host system

Docker container forward port 22 of the sshd daemon in the container to the
port 1122 of the host system.
