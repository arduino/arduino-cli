#!/bin/sh

# Data dir in the home dir, as the default data dir, not in /tmp (often a tmpfs on Linux)
export ARDUINO_DIRECTORIES_DATA=$(mktemp -d "$HOME/.arduino15-bench.XXXXXX")
trap 'rm -rf "$ARDUINO_DIRECTORIES_DATA"' EXIT

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
