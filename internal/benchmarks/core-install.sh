#!/bin/sh
#
# Compares the time of `core install arduino:avr` between a build of master and a build of a branch.
# Use it to measure the cost of a change in the install path, for example a disk sync after a write.
#
# Requirements:
# - hyperfine (https://github.com/sharkdp/hyperfine), for example `apt install hyperfine`.
# - Two builds in the current dir, named arduino-cli-master and arduino-cli-branch.
#   On each branch, build with: task build EXTRA_FLAGS="-o arduino-cli-<master|branch>"
#   For another target, set GOOS and GOARCH, for example GOOS=linux GOARCH=arm64.
# - About 1 GB of free space in the home dir, plus 500 MB for the worst case.
#
# Usage:
#   sh <path to>/core-install.sh   (run it from the dir with the two builds)
#
# The data dir is in the home dir, so the test uses the same disk as a real install.
# The script removes it at exit. After a crash, remove ~/.arduino15-bench.* by hand.
# The normal case runs 20 times. The worst case runs 10 times, with 500 MB of dirty data before each install.
# The worst case shows the cost of a change that waits for the disk.

# Data dir in the home dir, as the default data dir, not in /tmp (often a tmpfs on Linux)
export ARDUINO_DIRECTORIES_DATA=$(mktemp -d "$HOME/.arduino15-bench.XXXXXX")
trap 'rm -rf "$ARDUINO_DIRECTORIES_DATA"' EXIT
trap 'exit 1' INT TERM

./arduino-cli-master core update-index
./arduino-cli-master core download arduino:avr

# Normal case
hyperfine --warmup 1 --runs 20 \
  --prepare './arduino-cli-master core uninstall arduino:avr; true' \
  -n master './arduino-cli-master core install arduino:avr' \
  -n branch './arduino-cli-branch core install arduino:avr'

# Worst case: 500 MB of dirty data in the page cache before each install
hyperfine --warmup 1 --runs 10 \
  --prepare "./arduino-cli-master core uninstall arduino:avr; dd if=/dev/zero of=$ARDUINO_DIRECTORIES_DATA/dirty bs=1048576 count=500 2>/dev/null" \
  -n master-dirty './arduino-cli-master core install arduino:avr' \
  -n branch-dirty './arduino-cli-branch core install arduino:avr'
