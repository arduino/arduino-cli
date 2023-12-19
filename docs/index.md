Arduino CLI is an all-in-one solution that provides Boards/Library Managers, sketch builder, board detection, uploader,
and many other tools needed to use any Arduino compatible board and platform from command line or machine interfaces.

In addition to being a standalone tool, Arduino CLI is the heart of all official Arduino development software (Arduino
IDE, Arduino Web Editor). Parts of this documentation apply to those tools as well.

## Installation

You have several options to install the latest version of the Arduino CLI on your system, see the [installation] page.

## Getting started

Follow the [Getting started guide] to see how to use the most common CLI commands available.

## Using the gRPC interface

The [client_example] folder contains a sample program that shows how to use the gRPC interface of the CLI. Available
services and messages are detailed in the [gRPC reference] pages.

## Versioning and backward compatibility policy

This software is currently under active development: anything can change at any time, API and UI must be considered
unstable until we release version 1.0.0. For more information see our [versioning and backward compatibility] policy.

[installation]: installation.md
[getting started guide]: getting-started.md
[client_example]: https://github.com/arduino/arduino-cli/blob/master/rpc/internal/client_example
[grpc reference]: rpc/commands.md
[versioning and backward compatibility]: versioning.md
