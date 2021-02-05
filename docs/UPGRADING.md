# Upgrading

Here you can find a list of migration guides to handle breaking changes between releases of the CLI.

## Unreleased

### Change type of `CompileReq.ExportBinaries` message in gRPC interface

This change affects only the gRPC consumers.

In the `CompileReq` message the `export_binaries` property type has been changed from `bool` to
`google.protobuf.BoolValue`. This has been done to handle settings bindings by gRPC consumers and the CLI in the same
way so that they an identical behaviour.

## 0.15.0

### Rename `telemetry` settings to `metrics`

All instances of the term `telemetry` in the code and the documentation has been changed to `metrics`. This has been
done to clarify that no data is currently gathered from users of the CLI.

To handle this change the users must edit their config file, usually `arduino-cli.yaml`, and change the `telemetry` key
to `metrics`. The modification must be done by manually editing the file using a text editor, it can't be done via CLI.
No other action is necessary.

The default folders for the `arduino-cli.yaml` are:

- Linux: `/home/<your_username>/.arduino15/arduino-cli.yaml`
- OS X: `/Users/<your_username>/Library/Arduino15/arduino-cli.yaml`
- Windows: `C:\Users\<your_username>\AppData\Arduino15\arduino-cli.yaml`

## 0.14.0

### Changes in `debug` command

Previously it was required:

- To provide a debug command line recipe in `platform.txt` like `tools.reciped-id.debug.pattern=.....` that will start a
  `gdb` session for the selected board.
- To add a `debug.tool` definition in the `boards.txt` to recall that recipe, for example `myboard.debug.tool=recipe-id`

Now:

- Only the configuration needs to be supplied, the `arduino-cli` or the GUI tool will figure out how to call and setup
  the `gdb` session. An example of configuration is the following:

```
debug.executable={build.path}/{build.project_name}.elf
debug.toolchain=gcc
debug.toolchain.path={runtime.tools.arm-none-eabi-gcc-7-2017q4.path}/bin/
debug.toolchain.prefix=arm-none-eabi-
debug.server=openocd
debug.server.openocd.path={runtime.tools.openocd-0.10.0-arduino7.path}/bin/
debug.server.openocd.scripts_dir={runtime.tools.openocd-0.10.0-arduino7.path}/share/openocd/scripts/
debug.server.openocd.script={runtime.platform.path}/variants/{build.variant}/{build.openocdscript}
```

The `debug.server.XXXX` subkeys are optional and also "free text", this means that the configuration may be extended as
needed by the specific server. For now only `openocd` is supported. Anyway, if this change works, any other kind of
server may be fairly easily added.

The `debug.xxx=yyy` definitions above may be supplied and overlayed in the usual ways:

- on `platform.txt`: definition here will be shared through all boards in the platform
- on `boards.txt` as part of a board definition: they will override the global platform definitions
- on `programmers.txt`: they will override the boards and global platform definitions if the programmer is selected

### Binaries export must now be explicitly specified

Previously, if the `--build-path` was not specified, compiling a Sketch would copy the generated binaries in
`<sketch_folder>/build/<fqbn>/`, uploading to a board required that path to exist and contain the necessary binaries.

The `--dry-run` flag was removed.

The default, `compile` does not copy generated binaries to the sketch folder. The `--export-binaries` (`-e`) flag was
introduced to copy the binaries from the build folder to the sketch one. `--export-binaries` is not required when using
the `--output-dir` flag. A related configuration key and environment variable has been added to avoid the need to always
specify the `--export-binaries` flag: `sketch.always_export_binaries` and `ARDUINO_SKETCH_ALWAYS_EXPORT_BINARIES`.

If `--input-dir` or `--input-file` is not set when calling `upload` the command will search for the deterministically
created build directory in the temp folder and use the binaries found there.

The gRPC interface has been updated accordingly, `dryRun` is removed.

### Programmers can't be listed anymore using `burn-bootloader -P list`

The `-P` flag is used to select the programmer used to burn the bootloader on the specified board. Using `-P list` to
list all the possible programmers for the current board was hackish.

This way has been removed in favour of `board details <fqbn> --list-programmers`.

### `lib install --git-url` and `--zip-file` must now be explicitly enabled

With the introduction of the `--git-url` and `--zip-file` flags the new config key `library.enable_unsafe_install` has
been added to enable them.

This changes the ouput of the `config dump` command.

### Change behaviour of `--config-file` flag with `config` commands

To create a new config file with `config init` one must now use `--dest-dir` or the new `--dest-file` flags. Previously
the config file would always be overwritten by this command, now it fails if the it already exists, to force the
previous behaviour the user must set the `--overwrite` flag.
