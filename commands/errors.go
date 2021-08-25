// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package commands

import (
	"fmt"

	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func composeErrorMsg(msg string, cause error) string {
	if cause == nil {
		return msg
	}
	return fmt.Sprintf("%v: %v", msg, cause)
}

// CommandError is an error that may be converted into a gRPC status.
type CommandError interface {
	ToRPCStatus() *status.Status
}

// InvalidInstanceError is returned if the instance used in the command is not valid.
type InvalidInstanceError struct{}

func (e *InvalidInstanceError) Error() string {
	return tr("Invalid instance")
}

func (e *InvalidInstanceError) ToRPCStatus() *status.Status {
	return status.New(codes.InvalidArgument, e.Error())
}

// InvalidFQBNError is returned when the FQBN has syntax errors
type InvalidFQBNError struct {
	Cause error
}

func (e *InvalidFQBNError) Error() string {
	return composeErrorMsg(tr("Invalid FQBN"), e.Cause)
}

func (e *InvalidFQBNError) ToRPCStatus() *status.Status {
	return status.New(codes.InvalidArgument, e.Error())
}

func (e *InvalidFQBNError) Unwrap() error {
	return e.Cause
}

// MissingFQBNError is returned when the FQBN is mandatory and not specified
type MissingFQBNError struct{}

func (e *MissingFQBNError) Error() string {
	return tr("Missing FQBN (Fully Qualified Board Name)")
}

func (e *MissingFQBNError) ToRPCStatus() *status.Status {
	return status.New(codes.InvalidArgument, e.Error())
}

// UnknownFQBNError is returned when the FQBN is not found
type UnknownFQBNError struct {
	Cause error
}

func (e *UnknownFQBNError) Error() string {
	return composeErrorMsg(tr("Unknown FQBN"), e.Cause)
}

func (e *UnknownFQBNError) Unwrap() error {
	return e.Cause
}

func (e *UnknownFQBNError) ToRPCStatus() *status.Status {
	return status.New(codes.NotFound, e.Error())
}

// MissingPortProtocolError is returned when the port protocol is mandatory and not specified
type MissingPortProtocolError struct{}

func (e *MissingPortProtocolError) Error() string {
	return tr("Missing port protocol")
}

func (e *MissingPortProtocolError) ToRPCStatus() *status.Status {
	return status.New(codes.InvalidArgument, e.Error())
}

// MissingProgrammerError is returned when the programmer is mandatory and not specified
type MissingProgrammerError struct{}

func (e *MissingProgrammerError) Error() string {
	return tr("Missing programmer")
}

func (e *MissingProgrammerError) ToRPCStatus() *status.Status {
	return status.New(codes.InvalidArgument, e.Error())
}

// ProgreammerRequiredForUploadError is returned then the upload can be done only using a programmer
type ProgreammerRequiredForUploadError struct{}

func (e *ProgreammerRequiredForUploadError) Error() string {
	return tr("A programmer is required to upload")
}

func (e *ProgreammerRequiredForUploadError) ToRPCStatus() *status.Status {
	st, _ := status.
		New(codes.InvalidArgument, e.Error()).
		WithDetails(&rpc.ProgrammerIsRequiredForUploadError{})
	return st
}

// UnknownProgrammerError is returned when the programmer is not found
type UnknownProgrammerError struct {
	Cause error
}

func (e *UnknownProgrammerError) Error() string {
	return composeErrorMsg(tr("Unknown programmer"), e.Cause)
}

func (e *UnknownProgrammerError) Unwrap() error {
	return e.Cause
}

func (e *UnknownProgrammerError) ToRPCStatus() *status.Status {
	return status.New(codes.NotFound, e.Error())
}

// InvalidPlatformPropertyError is returned when a property in the platform is not valid
type InvalidPlatformPropertyError struct {
	Property string
	Value    string
}

func (e *InvalidPlatformPropertyError) Error() string {
	return tr("Invalid '%[1]s' property: %[2]s", e.Property, e.Value)
}

func (e *InvalidPlatformPropertyError) ToRPCStatus() *status.Status {
	return status.New(codes.FailedPrecondition, e.Error())
}

// MissingPlatformPropertyError is returned when a property in the platform is not found
type MissingPlatformPropertyError struct {
	Property string
}

func (e *MissingPlatformPropertyError) Error() string {
	return tr("Property '%s' is undefined", e.Property)
}

func (e *MissingPlatformPropertyError) ToRPCStatus() *status.Status {
	return status.New(codes.FailedPrecondition, e.Error())
}

// PlatformNotFound is returned when a platform is not found
type PlatformNotFound struct {
	Platform string
}

func (e *PlatformNotFound) Error() string {
	return tr("Platform '%s' is not installed", e.Platform)
}

func (e *PlatformNotFound) ToRPCStatus() *status.Status {
	return status.New(codes.FailedPrecondition, e.Error())
}

// MissingSketchPathError is returned when the sketch path is mandatory and not specified
type MissingSketchPathError struct{}

func (e *MissingSketchPathError) Error() string {
	return tr("Missing sketch path")
}

func (e *MissingSketchPathError) ToRPCStatus() *status.Status {
	return status.New(codes.InvalidArgument, e.Error())
}

// SketchNotFoundError is returned when the sketch is not found
type SketchNotFoundError struct {
	Cause error
}

func (e *SketchNotFoundError) Error() string {
	return composeErrorMsg(tr("Sketch not found"), e.Cause)
}

func (e *SketchNotFoundError) Unwrap() error {
	return e.Cause
}

func (e *SketchNotFoundError) ToRPCStatus() *status.Status {
	return status.New(codes.NotFound, e.Error())
}

// FailedUploadError is returned when the upload fails
type FailedUploadError struct {
	Message string
	Cause   error
}

func (e *FailedUploadError) Error() string {
	return composeErrorMsg(e.Message, e.Cause)
}

func (e *FailedUploadError) Unwrap() error {
	return e.Cause
}

func (e *FailedUploadError) ToRPCStatus() *status.Status {
	return status.New(codes.Internal, e.Error())
}

// CompileFailedError is returned when the compile fails
type CompileFailedError struct {
	Message string
	Cause   error
}

func (e *CompileFailedError) Error() string {
	return composeErrorMsg(e.Message, e.Cause)
}

func (e *CompileFailedError) Unwrap() error {
	return e.Cause
}

func (e *CompileFailedError) ToRPCStatus() *status.Status {
	return status.New(codes.Internal, e.Error())
}

// InvalidArgumentError is returned when an invalid argument is passed to the command
type InvalidArgumentError struct {
	Message string
	Cause   error
}

func (e *InvalidArgumentError) Error() string {
	return composeErrorMsg(e.Message, e.Cause)
}

func (e *InvalidArgumentError) Unwrap() error {
	return e.Cause
}

func (e *InvalidArgumentError) ToRPCStatus() *status.Status {
	return status.New(codes.InvalidArgument, e.Error())
}

// NotFoundError is returned when a resource is not found
type NotFoundError struct {
	Message string
	Cause   error
}

func (e *NotFoundError) Error() string {
	return composeErrorMsg(e.Message, e.Cause)
}

func (e *NotFoundError) Unwrap() error {
	return e.Cause
}

func (e *NotFoundError) ToRPCStatus() *status.Status {
	return status.New(codes.NotFound, e.Error())
}

// PermissionDeniedError is returned when a resource cannot be accessed or modified
type PermissionDeniedError struct {
	Message string
	Cause   error
}

func (e *PermissionDeniedError) Error() string {
	return composeErrorMsg(e.Message, e.Cause)
}

func (e *PermissionDeniedError) Unwrap() error {
	return e.Cause
}

func (e *PermissionDeniedError) ToRPCStatus() *status.Status {
	return status.New(codes.PermissionDenied, e.Error())
}

// UnavailableError is returned when a resource is temporarily not available
type UnavailableError struct {
	Message string
	Cause   error
}

func (e *UnavailableError) Error() string {
	return composeErrorMsg(e.Message, e.Cause)
}

func (e *UnavailableError) Unwrap() error {
	return e.Cause
}

func (e *UnavailableError) ToRPCStatus() *status.Status {
	return status.New(codes.Unavailable, e.Error())
}
