# Backward compatibility policy for arduino-cli.

The arduino-cli project follows a strict semantic versioning policy. We are committing not to make breaking changes in
minor releases of Arduino CLI 1.x.x.

The release rules are the following:

- Alpha phase `0.0.X`: In this phase, the software is going through a quick iteration of the API, each release (with
  increments of X) may contain massive and breaking changes.
- Beta phase `0.Y.X`: The software is usable, but the API is still not settled and is under continuous testing and
  review. Breaking changes are expected. Bug fixes and new features are made as patch releases (with increments of X).
  Breaking changes due to API refinements are made as minor releases (with increments of Y).
- Production release-candidate `1.0.0-rc.X`: in this phase, the software is considered ready for release and distributed
  to the users for final testing. Release candidates (with increments of X) are possible for bug fixes only.
- **Production release `1.Y.X`**: For the production releases backward compatibility is guaranteed, and all the breaking
  changes are cumulated for the next major release (2.0.0). Bug fixes are made as patch releases (with increments of X);
  New features are released as minor releases (with increments of Y).
- Next major release development `>=2.0.0` and up: see below.

## Backward compatibility guarantees and definition of "breaking change"

There are three main user facing API in the arduino-cli project:

- the standalone command-line API
- the gRPC API
- the golang API

Let's examine the backward compatibility rules for each one of these categories.

### Breaking changes in the command-line app

Changes in the command-line interface are considered breaking if:

- a command, a positional argument, or a flag is removed or renamed
- a command, a positional argument, or a flag behavior is changed
- an optional positional argument or a flag is made mandatory
- a positional argument or a flag format is changed

The following changes to the command-line syntax are NOT considered breaking changes:

- a new command is added
- a new optional positional argument is added
- a new optional flag is added

Any change in the **human-readable** text output is **NOT** considered a breaking change. In general, the human-readable
text is subject to translation and small adjustments in natural language syntax and presentation.

We will consider breaking changes only in the **machine-readable** output of the commands using the `--json` flag. In
particular, we have a breaking change in the JSON command output if:

- a key in a JSON object is renamed or removed.
- a value in a JSON object or array changes meaning or changes format.

We do **NOT** have a breaking change if:

- a new key is added to an existing JSON object

### Breaking changes in the gRPC API

To ensure gRPC API backward compatibility the only allowed changes are:

- adding a new service
- adding a new method to a service
- adding a field to an existing message
- adding a value to an enum

In general, **adding** to the gRPC API is allowed, **ANY OTHER** change will be considered a breaking change, some
examples are:

- renaming a service
- renaming a method
- changing a method signature
- renaming a field in a message
- changing a field type in a message
- deleting a field in a message
- etc.

The gRPC API is defined as a gRPC service running in the endpoint `cc.arduino.cli.commands.v1`. When a breaking change
happens a new gRPC endpoint is created from the existing API. The first breaking change will be implemented in the new
service `cc.arduino.cli.commands.v2`.

### Breaking changes in the golang API

The public golang API from the import path `github.com/arduino/arduino-cli` is guaranteed to be stable. Breaking changes
in the API will follow the go-lang guidelines and will be implemented by changing the import path by adding the `/v2`
suffix: `github.com/arduino/arduino-cli/v2`.

## Development process for the next major releases.

The development of the 2.0.0 release will proceed in a separate git branch `2.x.x`, in parallel with the 1.0.0 releases
that will continue on the `master` git branch.

New features and bug fixes should be made on the `master` branch and ported to the `2.x.x` once completed (unless it's a
2.0 specific change, in that case, it's fine to develop directly on the `2.x.x` branch).

Future releases and pre-releases of the `2.x.x` will follow the following versioning policy:

- Beta `2.0.0-beta.X.Y`: The v2 API is still under testing and review. Bug fixes and new features are released with
  increments of Y. Breaking changes are still possible and released with increments of X.
- Release Candidate `2.0.0-rc.X`: The v2 API is ready for release. Release candidates are distributed for user testing.
  Bug-fix releases only are allowed (with increments of X).
- `2.0.0` and up: The same rules for the `1.0.0` applies.

After the 2.0.0 release, the `master` branch will be moved to `2.x.x`, and the 1.0 branch will be tracked by a new
`1.x.x` branch.

The command-line interface for CLI 2.0 will be incompatible with CLI 1.0. Some commands may still be compatible though
depending on the amount of changes.

The gRPC daemon is flexible enough to run both services v1 and v2 at the same time. This capability allows a deprecation
period to allow a soft transition from v1 API to v2 API. We will deprecate the v1 API in the CLI 2.0 series but we will
continue to support it until the next major release CLI 3.0. At that point, we may decide to drop the support for the v1
API entirely but, depending on the balance between user demand and maintenance effort, we may decide to continue to
support it.

The go-lang API import path will be updated, following the go modules guidelines, by adding the `/v2` suffix:
`github.com/arduino/arduino-cli/v2`.

Unlike the gRPC counterpart, we will not guarantee a deprecation policy and a soft transition period for the go-lang API
(but again depending on the balance between user demand and maintenance effort we may decide to deprecate some API).
