#!/bin/sh
#
# Compares the compile time of a sketch between a build of master and a build of a branch.
# Use it to measure the cost of a change in the build path, for example the preprocessor flags.
#
# Requirements:
# - hyperfine (https://github.com/sharkdp/hyperfine), for example `apt install hyperfine`.
# - Two builds in the current dir, named arduino-cli-master and arduino-cli-branch.
#   On each branch, build with: task build EXTRA_FLAGS="-o arduino-cli-<master|branch>"
#   For another target, set GOOS and GOARCH, for example GOOS=linux GOARCH=arm64.
# - The platform of the FQBN and the libraries of the sketch, installed in the default data dir.
#
# Usage:
#   sh <path to>/compile.sh <fqbn> <sketch dir>   (run it from the dir with the two builds)
#
# The cold case cleans the build dir before each run. The warm case changes nothing, so nothing compiles
# again. The edit case appends a comment to the sketch before each run, so only the sketch compiles
# again. The script restores the sketch at exit.

MAIN="$2/$(basename "$2").ino"
BUILD=$(mktemp -d)
cp "$MAIN" "$BUILD/orig.ino"
trap 'cp "$BUILD/orig.ino" "$MAIN"; rm -rf "$BUILD"' EXIT
trap 'exit 1' INT TERM

hyperfine --warmup 1 --runs 5 \
  -n master-cold "./arduino-cli-master compile -b $1 --clean --build-path $BUILD/master $2" \
  -n branch-cold "./arduino-cli-branch compile -b $1 --clean --build-path $BUILD/branch $2"

hyperfine --warmup 1 --runs 10 \
  -n master-warm "./arduino-cli-master compile -b $1 --build-path $BUILD/master $2" \
  -n branch-warm "./arduino-cli-branch compile -b $1 --build-path $BUILD/branch $2"

hyperfine --warmup 1 --runs 10 \
  --prepare "echo '// bench' >> $MAIN" \
  -n master-edit "./arduino-cli-master compile -b $1 --build-path $BUILD/master $2" \
  -n branch-edit "./arduino-cli-branch compile -b $1 --build-path $BUILD/branch $2"
