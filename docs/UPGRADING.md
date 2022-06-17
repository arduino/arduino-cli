# Upgrading

Here you can find a list of migration guides to handle breaking changes between releases of the CLI.

## 0.24.0

### The flag `--debug-file path` in `daemon` command now overwrites the file instead of appending.

Previously running the Arduino CLI in deamon mode with the flag `--debug-file log.txt` will append to `log.txt`, now the
file is overwritten. Since the amount of log produced is very high this will help to keep the log size on a reasonable
amount if the flag is left enabled.

## 0.23.0

### Arduino IDE builtin libraries are now excluded from the build when running `arduino-cli` standalone

Previously the "builtin libraries" in the Arduino IDE 1.8.x were always included in the build process. This wasn't the
intended behaviour, `arduino-cli` should include them only if run as a daemon from the Arduino IDE. Now this is fixed,
but since it has been the default behaviour from a very long time we decided to report it here as a breaking change.

If a compilation fail for a missing bundled library, you can fix it just by installing the missing library from the
library manager as usual.

### gRPC: Changes in message `cc.arduino.cli.commands.v1.PlatformReference`

The gRPC message structure `cc.arduino.cli.commands.v1.PlatformReference` has been renamed to
`cc.arduino.cli.commands.v1.InstalledPlatformReference`, and some new fields have been added:

- `install_dir` is the installation directory of the platform
- `package_url` is the 3rd party platform URL of the platform

It is currently used only in `cc.arduino.cli.commands.v1.CompileResponse`, so the field type has been changed as well.
Old gRPC clients must only update gRPC bindings. They can safely ignore the new fields if not needed.

### golang API: `github.com/arduino/arduino-cli/cli/globals.DefaultIndexURL` has been moved under `github.com/arduino/arduino-cli/arduino/globals`

Legacy code should just update the import.

### golang API: PackageManager.DownloadPlatformRelease no longer need `label` parameter

```go
func (pm *PackageManager) DownloadPlatformRelease(platform *cores.PlatformRelease, config *downloader.Config, label string, progressCB rpc.DownloadProgressCB) error {
```

is now:

```go
func (pm *PackageManager) DownloadPlatformRelease(platform *cores.PlatformRelease, config *downloader.Config, progressCB rpc.DownloadProgressCB) error {
```

Just remove the `label` parameter from legacy code.

## 0.22.0

### `github.com/arduino/arduino-cli/arduino.MultipleBoardsDetectedError` field changed type

Now the `Port` field of the error is a `github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1.Port`, usually
imported as `rpc.Port`. The old `discovery.Port` can be converted to the new one using the `.ToRPC()` method.

### Function `github.com/arduino/arduino-cli/commands/upload.DetectConnectedBoard(...)` has been removed

Use `github.com/arduino/arduino-cli/commands/board.List(...)` to detect boards.

### Function `arguments.GetDiscoveryPort(...)` has been removed

NOTE: the functions in the `arguments` package doesn't have much use outside of the `arduino-cli` so we are considering
to remove them from the public golang API making them `internal`.

The old function:

```go
func (p *Port) GetDiscoveryPort(instance *rpc.Instance, sk *sketch.Sketch) *discovery.Port { }
```

is now replaced by the more powerful:

```go
func (p *Port) DetectFQBN(inst *rpc.Instance) (string, *rpc.Port) { }

func CalculateFQBNAndPort(portArgs *Port, fqbnArg *Fqbn, instance *rpc.Instance, sk *sketch.Sketch) (string, *rpc.Port) { }
```

### gRPC: `address` parameter has been removed from `commands.SupportedUserFieldsRequest`

The parameter is no more needed. Lagacy code will continue to work without modification (the value of the parameter will
be just ignored).

### The content of package `github.com/arduino/arduino-cli/httpclient` has been moved to a different path

In particular:

- `UserAgent` and `NetworkProxy` have been moved to `github.com/arduino/arduino-cli/configuration`
- the remainder of the package `github.com/arduino/arduino-cli/httpclient` has been moved to
  `github.com/arduino/arduino-cli/arduino/httpclient`

The old imports must be updated according to the list above.

### `commands.DownloadProgressCB` and `commands.TaskProgressCB` have been moved to package `github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1`

All references to these types must be updated with the new import.

### `commands.GetDownloaderConfig` has been moved to package `github.com/arduino/arduino-cli/arduino/httpclient`

All references to this function must be updated with the new import.

### `commands.Download` has been removed and replaced by `github.com/arduino/arduino-cli/arduino/httpclient.DownloadFile`

The old function must be replaced by the new one that is much more versatile.

### `packagemanager.PackageManager.DownloadToolRelease`, `packagemanager.PackageManager.DownloadPlatformRelease`, and `resources.DownloadResource.Download` functions change signature and behaviour

The following functions:

```go
func (pm *PackageManager) DownloadToolRelease(tool *cores.ToolRelease, config *downloader.Config) (*downloader.Downloader, error)
func (pm *PackageManager) DownloadPlatformRelease(platform *cores.PlatformRelease, config *downloader.Config) (*downloader.Downloader, error)
func (r *DownloadResource) Download(downloadDir *paths.Path, config *downloader.Config) (*downloader.Downloader, error)
```

now requires a label and a progress callback parameter, do not return the `Downloader` object anymore, and they
automatically handles the download internally:

```go
func (pm *PackageManager) DownloadToolRelease(tool *cores.ToolRelease, config *downloader.Config, label string, progressCB rpc.DownloadProgressCB) error
func (pm *PackageManager) DownloadPlatformRelease(platform *cores.PlatformRelease, config *downloader.Config, label string, progressCB rpc.DownloadProgressCB) error
func (r *DownloadResource) Download(downloadDir *paths.Path, config *downloader.Config, label string, downloadCB rpc.DownloadProgressCB) error
```

The new progress parameters must be added to legacy code, if progress reports are not needed an empty stub for `label`
and `progressCB` must be provided. There is no more need to execute the `downloader.Run()` or
`downloader.RunAndPoll(...)` method.

For example, the old legacy code like:

```go
downloader, err := pm.DownloadPlatformRelease(platformToDownload, config)
if err != nil {
    ...
}
if err := downloader.Run(); err != nil {
    ...
}
```

may be ported to the new version as:

```go
err := pm.DownloadPlatformRelease(platformToDownload, config, "", func(progress *rpc.DownloadProgress) {})
```

### `packagemanager.Load*` functions now returns `error` instead of `*status.Status`

The following functions signature:

```go
func (pm *PackageManager) LoadHardware() []*status.Status { ... }
func (pm *PackageManager) LoadHardwareFromDirectories(hardwarePaths paths.PathList) []*status.Status { ... }
func (pm *PackageManager) LoadHardwareFromDirectory(path *paths.Path) []*status.Status { ... }
func (pm *PackageManager) LoadToolsFromBundleDirectories(dirs paths.PathList) []*status.Status { ... }
func (pm *PackageManager) LoadDiscoveries() []*status.Status { ... }
```

have been changed to:

```go
func (pm *PackageManager) LoadHardware() []error { ... }
func (pm *PackageManager) LoadHardwareFromDirectories(hardwarePaths paths.PathList) []error { ... }
func (pm *PackageManager) LoadHardwareFromDirectory(path *paths.Path) []error { ... }
func (pm *PackageManager) LoadToolsFromBundleDirectories(dirs paths.PathList) []error { ... }
func (pm *PackageManager) LoadDiscoveries() []error { ... }
```

These function no longer returns a gRPC status, so the errors can be handled as any other `error`.

### Removed `error` return from `discovery.New(...)` function

The `discovery.New(...)` function never fails, so the error has been removed, the old signature:

```go
func New(id string, args ...string) (*PluggableDiscovery, error) { ... }
```

is now:

```go
func New(id string, args ...string) *PluggableDiscovery { ... }
```

## 0.21.0

### `packagemanager.NewPackageManager` function change

A new argument `userAgent` has been added to `packagemanager.NewPackageManager`, the new function signature is:

```go
func NewPackageManager(indexDir, packagesDir, downloadDir, tempDir *paths.Path, userAgent string) *PackageManager {
```

The userAgent string must be in the format `"ProgramName/Version"`, for example `"arduino-cli/0.20.1"`.

### `commands.Create` function change

A new argument `extraUserAgent` has been added to `commands.Create`, the new function signature is:

```go
func Create(req *rpc.CreateRequest, extraUserAgent ...string) (*rpc.CreateResponse, error) {
```

`extraUserAgent` is an array of strings, so multiple user agent may be provided. Each user agent must be in the format
`"ProgramName/Version"`, for example `"arduino-cli/0.20.1"`.

### `commands.Compile` function change

A new argument `progressCB` has been added to `commands.Compile`, the new function signature is:

```go
func Compile(
	ctx context.Context,
	req *rpc.CompileRequest,
	outStream, errStream io.Writer,
	progressCB commands.TaskProgressCB,
	debug bool
) (r *rpc.CompileResponse, e error) {
```

if a callback function is provided the `Compile` command will call it periodically with progress reports with the
percentage of compilation completed, otherwise, if the parameter is `nil`, no progress reports will be performed.

### `github.com/arduino/arduino-cli/cli/arguments.ParseReferences` function change

The `parseArch` parameter was removed since it was unused and was always true. This means that the architecture gets
always parsed by the function.

### `github.com/arduino/arduino-cli/cli/arguments.ParseReference` function change

The `parseArch` parameter was removed since it was unused and was always true. This means that the architecture gets
always parsed by the function. Furthermore the function now should also correctly interpret `packager:arch` spelled with
the wrong casing.

### `github.com/arduino/arduino-cli/executils.NewProcess` and `executils.NewProcessFromPath` function change

A new argument `extraEnv` has been added to `executils.NewProcess` and `executils.NewProcessFromPath`, the new function
signature is:

```go
func NewProcess(extraEnv []string, args ...string) (*Process, error) {
```

```go
func NewProcessFromPath(extraEnv []string, executable *paths.Path, args ...string) (*Process, error) {
```

The `extraEnv` params allow to pass environment variables (in addition to the default ones) to the spawned process.

### `github.com/arduino/arduino-cli/i18n.Init(...)` now requires an empty string to be passed for autodetection of locale

For automated detection of locale, change the call from:

```go
i18n.Init()
```

to

```go
i18n.Init("")
```

### `github.com/arduino/arduino-cli/legacy/i18n` module has been removed (in particular the `i18n.Logger`)

The `i18n.Logger` is no longer available. It was mainly used in the legacy builder struct field `Context.Logger`.

The `Context.Logger` field has been replaced with plain `io.Writer` fields `Contex.Stdout` and `Context.Stderr`. All
existing logger functionality has been dropped, for example the Java-Style formatting with tags like `{0} {1}...` must
be replaced with one of the equivalent golang printf-based alternatives and logging levels must be replaced with direct
writes to `Stdout` or `Stderr`.

## 0.20.0

### `board details` arguments change

The `board details` command now accepts only the `--fqbn` or `-b` flags to specify the FQBN.

The previously deprecated `board details <FQBN>` syntax is no longer supported.

### `board attach` arguments change

The `board attach` command now uses `--port` and `-p` flags to set board port and `--board` and `-b` flags to select its
FQBN.

The previous syntax `board attach <port>|<FQBN> [sketchPath]` is no longer supported.

### `--timeout` flag in `board list` command has been replaced by `--discovery-timeout`

The flag `--timeout` in the `board list` command is no longer supported.

## 0.19.0

### `board list` command JSON output change

The `board list` command JSON output has been changed quite a bit, from:

```
$ arduino-cli board list --format json
[
  {
    "address": "/dev/ttyACM1",
    "protocol": "serial",
    "protocol_label": "Serial Port (USB)",
    "boards": [
      {
        "name": "Arduino Uno",
        "fqbn": "arduino:avr:uno",
        "vid": "0x2341",
        "pid": "0x0043"
      }
    ],
    "serial_number": "954323132383515092E1"
  }
]
```

to:

```
$ arduino-cli board list --format json
[
  {
    "matching_boards": [
      {
        "name": "Arduino Uno",
        "fqbn": "arduino:avr:uno"
      }
    ],
    "port": {
      "address": "/dev/ttyACM1",
      "label": "/dev/ttyACM1",
      "protocol": "serial",
      "protocol_label": "Serial Port (USB)",
      "properties": {
        "pid": "0x0043",
        "serialNumber": "954323132383515092E1",
        "vid": "0x2341"
      }
    }
  }
]
```

The `boards` array has been renamed `matching_boards`, each contained object will now contain only `name` and `fqbn`.
Properties that can be used to identify a board are now moved to the new `properties` object, it can contain any key
name. `pid` and `vid` have been moved to `properties`, `serial_number` has been renamed `serialNumber` and moved to
`properties`. The new `label` field is the name of the `port` if it should be displayed in a GUI.

### gRPC interface `DebugConfigRequest`, `UploadRequest`, `UploadUsingProgrammerRequest`, `BurnBootloaderRequest`, `DetectedPort` field changes

`DebugConfigRequest`, `UploadRequest`, `UploadUsingProgrammerRequest` and `BurnBootloaderRequest` had their `port` field
change from type `string` to `Port`.

`Port` contains the following information:

```
// Port represents a board port that may be used to upload or to monitor a board
message Port {
  // Address of the port (e.g., `/dev/ttyACM0`).
  string address = 1;
  // The port label to show on the GUI (e.g. "ttyACM0")
  string label = 2;
  // Protocol of the port (e.g., `serial`, `network`, ...).
  string protocol = 3;
  // A human friendly description of the protocol (e.g., "Serial Port (USB)"
  string protocol_label = 4;
  // A set of properties of the port
  map<string, string> properties = 5;
}
```

The gRPC interface message `DetectedPort` has been changed from:

```
message DetectedPort {
  // Address of the port (e.g., `serial:///dev/ttyACM0`).
  string address = 1;
  // Protocol of the port (e.g., `serial`).
  string protocol = 2;
  // A human friendly description of the protocol (e.g., "Serial Port (USB)").
  string protocol_label = 3;
  // The boards attached to the port.
  repeated BoardListItem boards = 4;
  // Serial number of connected board
  string serial_number = 5;
}
```

to:

```
message DetectedPort {
  // The possible boards attached to the port.
  repeated BoardListItem matching_boards = 1;
  // The port details
  Port port = 2;
}
```

The properties previously contained directly in the message are now stored in the `port` property.

These changes are necessary for the pluggable discovery.

### gRPC interface `BoardListItem` change

The `vid` and `pid` fields of the `BoardListItem` message have been removed. They used to only be available when
requesting connected board lists, now that information is stored in the `port` field of `DetectedPort`.

### Change public library interface

#### `github.com/arduino/arduino-cli/i18n` package

The behavior of the `Init` function has changed. The user specified locale code is no longer read from the
`github.com/arduino/arduino-cli/configuration` package and now must be passed directly to `Init` as a string:

```go
i18n.Init("it")
```

Omit the argument for automated locale detection:

```go
i18n.Init()
```

#### `github.com/arduino/arduino-cli/arduino/builder` package

`GenBuildPath()` function has been moved to `github.com/arduino/arduino-cli/arduino/sketch` package. The signature is
unchanged.

`EnsureBuildPathExists` function from has been completely removed, in its place use
`github.com/arduino/go-paths-helper.MkDirAll()`.

`SketchSaveItemCpp` function signature is changed from `path string, contents []byte, destPath string` to
`path *paths.Path, contents []byte, destPath *paths.Path`. `paths` is `github.com/arduino/go-paths-helper`.

`SketchLoad` function has been removed, in its place use `New` from `github.com/arduino/arduino-cli/arduino/sketch`
package.

```diff
-      SketchLoad("/some/path", "")
+      sketch.New(paths.New("some/path))
}
```

If you need to set a custom build path you must instead set it after creating the Sketch.

```diff
-      SketchLoad("/some/path", "/my/build/path")
+      s, err := sketch.New(paths.New("some/path))
+      s.BuildPath = paths.new("/my/build/path")
}
```

`SketchCopyAdditionalFiles` function signature is changed from
`sketch *sketch.Sketch, destPath string, overrides map[string]string` to
`sketch *sketch.Sketch, destPath *paths.Path, overrides map[string]string`.

#### `github.com/arduino/arduino-cli/arduino/sketch` package

`Item` struct has been removed, use `go-paths-helper.Path` in its place.

`NewItem` has been removed too, use `go-paths-helper.New` in its place.

`GetSourceBytes` has been removed, in its place use `go-paths-helper.Path.ReadFile`. `GetSourceStr` too has been
removed, in its place:

```diff
-      s, err := item.GetSourceStr()
+      data, err := file.ReadFile()
+      s := string(data)
}
```

`ItemByPath` type and its member functions have been removed, use `go-paths-helper.PathList` in its place.

`Sketch.LocationPath` has been renamed to `FullPath` and its type changed from `string` to `go-paths-helper.Path`.

`Sketch.MainFile` type has changed from `*Item` to `go-paths-helper.Path`. `Sketch.OtherSketchFiles`,
`Sketch.AdditionalFiles` and `Sketch.RootFolderFiles` type has changed from `[]*Item` to `go-paths-helper.PathList`.

`New` signature has been changed from `sketchFolderPath, mainFilePath, buildPath string, allFilesPaths []string` to
`path *go-paths-helper.Path`.

`CheckSketchCasing` function is now private, the check is done internally by `New`.

`InvalidSketchFoldernameError` has been renamed `InvalidSketchFolderNameError`.

#### `github.com/arduino/arduino-cli/arduino/sketches` package

`Sketch` struct has been merged with `sketch.Sketch` struct.

`Metadata` and `BoardMetadata` structs have been moved to `github.com/arduino/arduino-cli/arduino/sketch` package.

`NewSketchFromPath` has been deleted, use `sketch.New` in its place.

`ImportMetadata` is now private called internally by `sketch.New`.

`ExportMetadata` has been moved to `github.com/arduino/arduino-cli/arduino/sketch` package.

`BuildPath` has been removed, use `sketch.Sketch.BuildPath` in its place.

`CheckForPdeFiles` has been moved to `github.com/arduino/arduino-cli/arduino/sketch` package.

#### `github.com/arduino/arduino-cli/legacy/builder/types` package

`Sketch` has been removed, use `sketch.Sketch` in its place.

`SketchToLegacy` and `SketchFromLegacy` have been removed, nothing replaces them.

`Context.Sketch` types has been changed from `Sketch` to `sketch.Sketch`.

### Change in `board details` response (gRPC and JSON output)

The `board details` output WRT board identification properties has changed, before it was:

```
$ arduino-cli board details arduino:samd:mkr1000
Board name:                Arduino MKR1000
FQBN:                      arduino:samd:mkr1000
Board version:             1.8.11
Debugging supported:       ✔

Official Arduino board:    ✔

Identification properties: VID:0x2341 PID:0x824e
                           VID:0x2341 PID:0x024e
                           VID:0x2341 PID:0x804e
                           VID:0x2341 PID:0x004e
[...]

$ arduino-cli board details arduino:samd:mkr1000 --format json
[...]
  "identification_prefs": [
    {
      "usb_id": {
        "vid": "0x2341",
        "pid": "0x804e"
      }
    },
    {
      "usb_id": {
        "vid": "0x2341",
        "pid": "0x004e"
      }
    },
    {
      "usb_id": {
        "vid": "0x2341",
        "pid": "0x824e"
      }
    },
    {
      "usb_id": {
        "vid": "0x2341",
        "pid": "0x024e"
      }
    }
  ],
[...]
```

now the properties have been renamed from `identification_prefs` to `identification_properties` and they are no longer
specific to USB but they can theoretically be any set of key/values:

```
$ arduino-cli board details arduino:samd:mkr1000
Board name:                Arduino MKR1000
FQBN:                      arduino:samd:mkr1000
Board version:             1.8.11
Debugging supported:       ✔

Official Arduino board:    ✔

Identification properties: vid=0x2341
                           pid=0x804e

Identification properties: vid=0x2341
                           pid=0x004e

Identification properties: vid=0x2341
                           pid=0x824e

Identification properties: vid=0x2341
                           pid=0x024e
[...]

$ arduino-cli board details arduino:samd:mkr1000 --format json
[...]
  "identification_properties": [
    {
      "properties": {
        "pid": "0x804e",
        "vid": "0x2341"
      }
    },
    {
      "properties": {
        "pid": "0x004e",
        "vid": "0x2341"
      }
    },
    {
      "properties": {
        "pid": "0x824e",
        "vid": "0x2341"
      }
    },
    {
      "properties": {
        "pid": "0x024e",
        "vid": "0x2341"
      }
    }
  ]
}
```

### Change of behaviour of gRPC `Init` function

Previously the `Init` function was used to both create a new `CoreInstance` and initialize it, so that the internal
package and library managers were already populated with all the information available from `*_index.json` files,
installed platforms and libraries and so on.

Now the initialization phase is split into two, first the client must create a new `CoreInstance` with the `Create`
function, that does mainly two things:

- create all folders necessary to correctly run the CLI if not already existing
- create and return a new `CoreInstance`

The `Create` function will only fail if folders creation is not successful.

The returned instance is relatively unusable since no library and no platform is loaded, some functions that don't need
that information can still be called though.

The `Init` function has been greatly overhauled and it doesn't fail completely if one or more platforms or libraries
fail to load now.

Also the option `library_manager_only` has been removed, the package manager is always initialized and platforms are
loaded.

The `Init` was already a server-side streaming function but it would always return one and only one response, this has
been modified so that each response is either an error or a notification on the initialization process so that it works
more like an actual stream of information.

Previously a client would call the function like so:

```typescript
const initReq = new InitRequest()
initReq.setLibraryManagerOnly(false)
const initResp = await new Promise<InitResponse>((resolve, reject) => {
  let resp: InitResponse | undefined = undefined
  const stream = client.init(initReq)
  stream.on("data", (data: InitResponse) => (resp = data))
  stream.on("end", () => resolve(resp!))
  stream.on("error", (err) => reject(err))
})

const instance = initResp.getInstance()
if (!instance) {
  throw new Error("Could not retrieve instance from the initialize response.")
}
```

Now something similar should be done.

```typescript
const createReq = new CreateRequest()
const instance = client.create(createReq)

if (!instance) {
  throw new Error("Could not retrieve instance from the initialize response.")
}

const initReq = new InitRequest()
initReq.setInstance(instance)
const initResp = client.init(initReq)
initResp.on("data", (o: InitResponse) => {
  const downloadProgress = o.getDownloadProgress()
  if (downloadProgress) {
    // Handle download progress
  }
  const taskProgress = o.getTaskProgress()
  if (taskProgress) {
    // Handle task progress
  }
  const err = o.getError()
  if (err) {
    // Handle error
  }
})

await new Promise<void>((resolve, reject) => {
  initResp.on("error", (err) => reject(err))
  initResp.on("end", resolve)
})
```

Previously if even one platform or library failed to load everything else would fail too, that doesn't happen anymore.
Now it's easier for both the CLI and the gRPC clients to handle gracefully platforms or libraries updates that might
break the initialization step and make everything unusable.

### Removal of gRPC `Rescan` function

The `Rescan` function has been removed, in its place the `Init` function must be used.

### Change of behaviour of gRPC `UpdateIndex` and `UpdateLibrariesIndex` functions

Previously both `UpdateIndex` and `UpdateLibrariesIndex` functions implicitly called `Rescan` so that the internal
`CoreInstance` was updated with the eventual new information obtained in the update.

This behaviour is now removed and the internal `CoreInstance` must be explicitly updated by the gRPC client using the
`Init` function.

### Removed rarely used golang API

The following function from the `github.com/arduino/arduino-cli/arduino/libraries` module is no longer available:

```go
func (lm *LibrariesManager) UpdateIndex(config *downloader.Config) (*downloader.Downloader, error) {
```

We recommend using the equivalent gRPC API to perform the update of the index.

## 0.18.0

### Breaking changes in gRPC API and CLI JSON output.

Starting from this release we applied a more rigorous and stricter naming conventions in gRPC API following the official
guidelines: https://developers.google.com/protocol-buffers/docs/style

We also started using a linter to implement checks for gRPC API style errors.

This provides a better consistency and higher quality API but inevitably introduces breaking changes.

### gRPC API breaking changes

Consumers of the gRPC API should regenerate their bindings and update all structures naming where necessary. Most of the
changes are trivial and falls into the following categories:

- Service names have been suffixed with `...Service` (for example `ArduinoCore` -> `ArduinoCoreService`)
- Message names suffix has been changed from `...Req`/`...Resp` to `...Request`/`...Response` (for example
  `BoardDetailsReq` -> `BoardDetailsRequest`)
- Enumerations now have their class name prefixed (for example the enumeration value `FLAT` in `LibraryLayout` has been
  changed to `LIBRARY_LAYOUT_FLAT`)
- Use of lower-snake case on all fields (for example: `ID` -> `id`, `FQBN` -> `fqbn`, `Name` -> `name`,
  `ArchiveFilename` -> `archive_filename`)
- Package names are now versioned (for example `cc.arduino.cli.commands` -> `cc.arduino.cli.commands.v1`)
- Repeated responses are now in plural form (`identification_pref` -> `identification_prefs`, `platform` -> `platforms`)

### arduino-cli JSON output breaking changes

Consumers of the JSON output of the CLI must update their clients if they use one of the following commands:

- in `core search` command the following fields have been renamed:

  - `Boards` -> `boards`
  - `Email` -> `email`
  - `ID` -> `id`
  - `Latest` -> `latest`
  - `Maintainer` -> `maintainer`
  - `Name` -> `name`
  - `Website` -> `website`

  The new output is like:

  ```
  $ arduino-cli core search Due --format json
  [
    {
      "id": "arduino:sam",
      "latest": "1.6.12",
      "name": "Arduino SAM Boards (32-bits ARM Cortex-M3)",
      "maintainer": "Arduino",
      "website": "http://www.arduino.cc/",
      "email": "packages@arduino.cc",
      "boards": [
        {
          "name": "Arduino Due (Native USB Port)",
          "fqbn": "arduino:sam:arduino_due_x"
        },
        {
          "name": "Arduino Due (Programming Port)",
          "fqbn": "arduino:sam:arduino_due_x_dbg"
        }
      ]
    }
  ]
  ```

- in `board details` command the following fields have been renamed:

  - `identification_pref` -> `identification_prefs`
  - `usbID` -> `usb_id`
  - `PID` -> `pid`
  - `VID` -> `vid`
  - `websiteURL` -> `website_url`
  - `archiveFileName` -> `archive_filename`
  - `propertiesId` -> `properties_id`
  - `toolsDependencies` -> `tools_dependencies`

  The new output is like:

  ```
  $ arduino-cli board details arduino:avr:uno --format json
  {
    "fqbn": "arduino:avr:uno",
    "name": "Arduino Uno",
    "version": "1.8.3",
    "properties_id": "uno",
    "official": true,
    "package": {
      "maintainer": "Arduino",
      "url": "https://downloads.arduino.cc/packages/package_index.json",
      "website_url": "http://www.arduino.cc/",
      "email": "packages@arduino.cc",
      "name": "arduino",
      "help": {
        "online": "http://www.arduino.cc/en/Reference/HomePage"
      }
    },
    "platform": {
      "architecture": "avr",
      "category": "Arduino",
      "url": "http://downloads.arduino.cc/cores/avr-1.8.3.tar.bz2",
      "archive_filename": "avr-1.8.3.tar.bz2",
      "checksum": "SHA-256:de8a9b982477762d3d3e52fc2b682cdd8ff194dc3f1d46f4debdea6a01b33c14",
      "size": 4941548,
      "name": "Arduino AVR Boards"
    },
    "tools_dependencies": [
      {
        "packager": "arduino",
        "name": "avr-gcc",
        "version": "7.3.0-atmel3.6.1-arduino7",
        "systems": [
          {
            "checksum": "SHA-256:3903553d035da59e33cff9941b857c3cb379cb0638105dfdf69c97f0acc8e7b5",
            "host": "arm-linux-gnueabihf",
            "archive_filename": "avr-gcc-7.3.0-atmel3.6.1-arduino7-arm-linux-gnueabihf.tar.bz2",
            "url": "http://downloads.arduino.cc/tools/avr-gcc-7.3.0-atmel3.6.1-arduino7-arm-linux-gnueabihf.tar.bz2",
            "size": 34683056
          },
          { ... }
        ]
      },
      { ... }
    ],
    "identification_prefs": [
      {
        "usb_id": {
          "vid": "0x2341",
          "pid": "0x0043"
        }
      },
      { ... }
    ],
    "programmers": [
      {
        "platform": "Arduino AVR Boards",
        "id": "parallel",
        "name": "Parallel Programmer"
      },
      { ... }
    ]
  }
  ```

- in `board listall` command the following fields have been renamed:

  - `FQBN` -> `fqbn`
  - `Email` -> `email`
  - `ID` -> `id`
  - `Installed` -> `installed`
  - `Latest` -> `latest`
  - `Name` -> `name`
  - `Maintainer` -> `maintainer`
  - `Website` -> `website`

  The new output is like:

  ```
  $ arduino-cli board listall Uno --format json
  {
    "boards": [
      {
        "name": "Arduino Uno",
        "fqbn": "arduino:avr:uno",
        "platform": {
          "id": "arduino:avr",
          "installed": "1.8.3",
          "latest": "1.8.3",
          "name": "Arduino AVR Boards",
          "maintainer": "Arduino",
          "website": "http://www.arduino.cc/",
          "email": "packages@arduino.cc"
        }
      }
    ]
  }
  ```

- in `board search` command the following fields have been renamed:

  - `FQBN` -> `fqbn`
  - `Email` -> `email`
  - `ID` -> `id`
  - `Installed` -> `installed`
  - `Latest` -> `latest`
  - `Name` -> `name`
  - `Maintainer` -> `maintainer`
  - `Website` -> `website`

  The new output is like:

  ```
  $ arduino-cli board search Uno --format json
  [
    {
      "name": "Arduino Uno",
      "fqbn": "arduino:avr:uno",
      "platform": {
        "id": "arduino:avr",
        "installed": "1.8.3",
        "latest": "1.8.3",
        "name": "Arduino AVR Boards",
        "maintainer": "Arduino",
        "website": "http://www.arduino.cc/",
        "email": "packages@arduino.cc"
      }
    }
  ]
  ```

- in `lib deps` command the following fields have been renamed:

  - `versionRequired` -> `version_required`
  - `versionInstalled` -> `version_installed`

  The new output is like:

  ```
  $ arduino-cli lib deps Arduino_MKRIoTCarrier --format json
  {
    "dependencies": [
      {
        "name": "Adafruit seesaw Library",
        "version_required": "1.3.1"
      },
      {
        "name": "SD",
        "version_required": "1.2.4",
        "version_installed": "1.2.3"
      },
      { ... }
    ]
  }
  ```

- in `lib search` command the following fields have been renamed:

  - `archivefilename` -> `archive_filename`
  - `cachepath` -> `cache_path`

  The new output is like:

  ```
  $ arduino-cli lib search NTPClient --format json
  {
    "libraries": [
      {
        "name": "NTPClient",
        "releases": {
          "1.0.0": {
            "author": "Fabrice Weinberg",
            "version": "1.0.0",
            "maintainer": "Fabrice Weinberg \u003cfabrice@weinberg.me\u003e",
            "sentence": "An NTPClient to connect to a time server",
            "paragraph": "Get time from a NTP server and keep it in sync.",
            "website": "https://github.com/FWeinb/NTPClient",
            "category": "Timing",
            "architectures": [
              "esp8266"
            ],
            "types": [
              "Arduino"
            ],
            "resources": {
              "url": "https://downloads.arduino.cc/libraries/github.com/arduino-libraries/NTPClient-1.0.0.zip",
              "archive_filename": "NTPClient-1.0.0.zip",
              "checksum": "SHA-256:b1f2907c9d51ee253bad23d05e2e9c1087ab1e7ba3eb12ee36881ba018d81678",
              "size": 6284,
              "cache_path": "libraries"
            }
          },
          "2.0.0": { ... },
          "3.0.0": { ... },
          "3.1.0": { ... },
          "3.2.0": { ... }
        },
        "latest": {
          "author": "Fabrice Weinberg",
          "version": "3.2.0",
          "maintainer": "Fabrice Weinberg \u003cfabrice@weinberg.me\u003e",
          "sentence": "An NTPClient to connect to a time server",
          "paragraph": "Get time from a NTP server and keep it in sync.",
          "website": "https://github.com/arduino-libraries/NTPClient",
          "category": "Timing",
          "architectures": [
            "*"
          ],
          "types": [
            "Arduino"
          ],
          "resources": {
            "url": "https://downloads.arduino.cc/libraries/github.com/arduino-libraries/NTPClient-3.2.0.zip",
            "archive_filename": "NTPClient-3.2.0.zip",
            "checksum": "SHA-256:122d00df276972ba33683aff0f7fe5eb6f9a190ac364f8238a7af25450fd3e31",
            "size": 7876,
            "cache_path": "libraries"
          }
        }
      }
    ],
    "status": 1
  }
  ```

- in `board list` command the following fields have been renamed:

  - `FQBN` -> `fqbn`
  - `VID` -> `vid`
  - `PID` -> `pid`

  The new output is like:

  ```
  $ arduino-cli board list --format json
  [
    {
      "address": "/dev/ttyACM0",
      "protocol": "serial",
      "protocol_label": "Serial Port (USB)",
      "boards": [
        {
          "name": "Arduino Nano 33 BLE",
          "fqbn": "arduino:mbed:nano33ble",
          "vid": "0x2341",
          "pid": "0x805a"
        },
        {
          "name": "Arduino Nano 33 BLE",
          "fqbn": "arduino-dev:mbed:nano33ble",
          "vid": "0x2341",
          "pid": "0x805a"
        },
        {
          "name": "Arduino Nano 33 BLE",
          "fqbn": "arduino-dev:nrf52:nano33ble",
          "vid": "0x2341",
          "pid": "0x805a"
        },
        {
          "name": "Arduino Nano 33 BLE",
          "fqbn": "arduino-beta:mbed:nano33ble",
          "vid": "0x2341",
          "pid": "0x805a"
        }
      ],
      "serial_number": "BECC45F754185EC9"
    }
  ]
  $ arduino-cli board list -w --format json
  {
    "type": "add",
    "address": "/dev/ttyACM0",
    "protocol": "serial",
    "protocol_label": "Serial Port (USB)",
    "boards": [
      {
        "name": "Arduino Nano 33 BLE",
        "fqbn": "arduino-dev:nrf52:nano33ble",
        "vid": "0x2341",
        "pid": "0x805a"
      },
      {
        "name": "Arduino Nano 33 BLE",
        "fqbn": "arduino-dev:mbed:nano33ble",
        "vid": "0x2341",
        "pid": "0x805a"
      },
      {
        "name": "Arduino Nano 33 BLE",
        "fqbn": "arduino-beta:mbed:nano33ble",
        "vid": "0x2341",
        "pid": "0x805a"
      },
      {
        "name": "Arduino Nano 33 BLE",
        "fqbn": "arduino:mbed:nano33ble",
        "vid": "0x2341",
        "pid": "0x805a"
      }
    ],
    "serial_number": "BECC45F754185EC9"
  }
  {
    "type": "remove",
    "address": "/dev/ttyACM0"
  }
  ```

## 0.16.0

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
- Windows: `C:\Users\<your_username>\AppData\Local\Arduino15\arduino-cli.yaml`

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
