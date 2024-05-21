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

package cmderrors

import (
	"fmt"
	"strings"

	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
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
	// GRPCStatus convertes the error into a *status.Status
	GRPCStatus() *status.Status
}

// InvalidInstanceError is returned if the instance used in the command is not valid.
type InvalidInstanceError struct{}

func (e *InvalidInstanceError) Error() string {
	return i18n.Tr("Invalid instance")
}

// GRPCStatus converts the error into a *status.Status
func (e *InvalidInstanceError) GRPCStatus() *status.Status {
	return status.New(codes.InvalidArgument, e.Error())
}

// InvalidFQBNError is returned when the FQBN has syntax errors
type InvalidFQBNError struct {
	Cause error
}

func (e *InvalidFQBNError) Error() string {
	return composeErrorMsg(i18n.Tr("Invalid FQBN"), e.Cause)
}

// GRPCStatus converts the error into a *status.Status
func (e *InvalidFQBNError) GRPCStatus() *status.Status {
	return status.New(codes.InvalidArgument, e.Error())
}

func (e *InvalidFQBNError) Unwrap() error {
	return e.Cause
}

// InvalidURLError is returned when the URL has syntax errors
type InvalidURLError struct {
	Cause error
}

func (e *InvalidURLError) Error() string {
	return composeErrorMsg(i18n.Tr("Invalid URL"), e.Cause)
}

// GRPCStatus converts the error into a *status.Status
func (e *InvalidURLError) GRPCStatus() *status.Status {
	return status.New(codes.InvalidArgument, e.Error())
}

func (e *InvalidURLError) Unwrap() error {
	return e.Cause
}

// InvalidLibraryError is returned when the library has syntax errors
type InvalidLibraryError struct {
	Cause error
}

func (e *InvalidLibraryError) Error() string {
	return composeErrorMsg(i18n.Tr("Invalid library"), e.Cause)
}

// GRPCStatus converts the error into a *status.Status
func (e *InvalidLibraryError) GRPCStatus() *status.Status {
	return status.New(codes.InvalidArgument, e.Error())
}

func (e *InvalidLibraryError) Unwrap() error {
	return e.Cause
}

// InvalidVersionError is returned when the version has syntax errors
type InvalidVersionError struct {
	Cause error
}

func (e *InvalidVersionError) Error() string {
	return composeErrorMsg(i18n.Tr("Invalid version"), e.Cause)
}

// GRPCStatus converts the error into a *status.Status
func (e *InvalidVersionError) GRPCStatus() *status.Status {
	return status.New(codes.InvalidArgument, e.Error())
}

func (e *InvalidVersionError) Unwrap() error {
	return e.Cause
}

// NoBoardsDetectedError is returned when detecting the FQBN of a board
// does not produce any result.
type NoBoardsDetectedError struct {
	Port *rpc.Port
}

func (e *NoBoardsDetectedError) Error() string {
	return i18n.Tr(
		"Please specify an FQBN. The board on port %[1]s with protocol %[2]s can't be identified",
		e.Port.GetAddress(),
		e.Port.GetProtocol(),
	)
}

// GRPCStatus converts the error into a *status.Status
func (e *NoBoardsDetectedError) GRPCStatus() *status.Status {
	return status.New(codes.InvalidArgument, e.Error())
}

// MultipleBoardsDetectedError is returned when trying to detect
// the FQBN of a board connected to a port fails because that
// are multiple possible boards detected.
type MultipleBoardsDetectedError struct {
	Port *rpc.Port
}

func (e *MultipleBoardsDetectedError) Error() string {
	return i18n.Tr(
		"Please specify an FQBN. Multiple possible boards detected on port %[1]s with protocol %[2]s",
		e.Port.GetAddress(),
		e.Port.GetProtocol(),
	)
}

// GRPCStatus converts the error into a *status.Status
func (e *MultipleBoardsDetectedError) GRPCStatus() *status.Status {
	return status.New(codes.InvalidArgument, e.Error())
}

// MissingFQBNError is returned when the FQBN is mandatory and not specified
type MissingFQBNError struct{}

func (e *MissingFQBNError) Error() string {
	return i18n.Tr("Missing FQBN (Fully Qualified Board Name)")
}

// GRPCStatus converts the error into a *status.Status
func (e *MissingFQBNError) GRPCStatus() *status.Status {
	return status.New(codes.InvalidArgument, e.Error())
}

// UnknownFQBNError is returned when the FQBN is not found
type UnknownFQBNError struct {
	Cause error
}

func (e *UnknownFQBNError) Error() string {
	return composeErrorMsg(i18n.Tr("Unknown FQBN"), e.Cause)
}

func (e *UnknownFQBNError) Unwrap() error {
	return e.Cause
}

// GRPCStatus converts the error into a *status.Status
func (e *UnknownFQBNError) GRPCStatus() *status.Status {
	return status.New(codes.NotFound, e.Error())
}

// UnknownProfileError is returned when the profile is not found
type UnknownProfileError struct {
	Profile string
	Cause   error
}

func (e *UnknownProfileError) Error() string {
	return composeErrorMsg(i18n.Tr("Profile '%s' not found", e.Profile), e.Cause)
}

func (e *UnknownProfileError) Unwrap() error {
	return e.Cause
}

// GRPCStatus converts the error into a *status.Status
func (e *UnknownProfileError) GRPCStatus() *status.Status {
	return status.New(codes.NotFound, e.Error())
}

// InvalidProfileError is returned when the profile has errors
type InvalidProfileError struct {
	Cause error
}

func (e *InvalidProfileError) Error() string {
	return composeErrorMsg(i18n.Tr("Invalid profile"), e.Cause)
}

func (e *InvalidProfileError) Unwrap() error {
	return e.Cause
}

// GRPCStatus converts the error into a *status.Status
func (e *InvalidProfileError) GRPCStatus() *status.Status {
	return status.New(codes.FailedPrecondition, e.Error())
}

// MissingPortAddressError is returned when the port protocol is mandatory and not specified
type MissingPortAddressError struct{}

func (e *MissingPortAddressError) Error() string {
	return i18n.Tr("Missing port address")
}

// GRPCStatus converts the error into a *status.Status
func (e *MissingPortAddressError) GRPCStatus() *status.Status {
	return status.New(codes.InvalidArgument, e.Error())
}

// MissingPortProtocolError is returned when the port protocol is mandatory and not specified
type MissingPortProtocolError struct{}

func (e *MissingPortProtocolError) Error() string {
	return i18n.Tr("Missing port protocol")
}

// GRPCStatus converts the error into a *status.Status
func (e *MissingPortProtocolError) GRPCStatus() *status.Status {
	return status.New(codes.InvalidArgument, e.Error())
}

// MissingPortError is returned when the port is mandatory and not specified
type MissingPortError struct{}

func (e *MissingPortError) Error() string {
	return i18n.Tr("Missing port")
}

// GRPCStatus converts the error into a *status.Status
func (e *MissingPortError) GRPCStatus() *status.Status {
	return status.New(codes.InvalidArgument, e.Error())
}

// NoMonitorAvailableForProtocolError is returned when a monitor for the specified port protocol is not available
type NoMonitorAvailableForProtocolError struct {
	Protocol string
}

func (e *NoMonitorAvailableForProtocolError) Error() string {
	return i18n.Tr("No monitor available for the port protocol %s", e.Protocol)
}

// GRPCStatus converts the error into a *status.Status
func (e *NoMonitorAvailableForProtocolError) GRPCStatus() *status.Status {
	return status.New(codes.InvalidArgument, e.Error())
}

// MissingProgrammerError is returned when the programmer is mandatory and not specified
type MissingProgrammerError struct{}

func (e *MissingProgrammerError) Error() string {
	return i18n.Tr("Missing programmer")
}

// GRPCStatus converts the error into a *status.Status
func (e *MissingProgrammerError) GRPCStatus() *status.Status {
	s, _ := status.New(codes.InvalidArgument, e.Error()).WithDetails(&rpc.MissingProgrammerError{})
	return s
}

// ProgrammerRequiredForUploadError is returned when the upload can be done only using a programmer
type ProgrammerRequiredForUploadError struct{}

func (e *ProgrammerRequiredForUploadError) Error() string {
	return i18n.Tr("A programmer is required to upload")
}

// GRPCStatus converts the error into a *status.Status
func (e *ProgrammerRequiredForUploadError) GRPCStatus() *status.Status {
	st, _ := status.
		New(codes.InvalidArgument, e.Error()).
		WithDetails(&rpc.ProgrammerIsRequiredForUploadError{})
	return st
}

// InitFailedError is returned when the instance initialization fails
type InitFailedError struct {
	Code   codes.Code
	Cause  error
	Reason rpc.FailedInstanceInitReason
}

func (ife *InitFailedError) Error() string {
	return ife.Cause.Error()
}

// GRPCStatus converts the error into a *status.Status
func (ife *InitFailedError) GRPCStatus() *status.Status {
	st, _ := status.
		New(ife.Code, ife.Cause.Error()).
		WithDetails(&rpc.FailedInstanceInitError{
			Reason:  ife.Reason,
			Message: ife.Cause.Error(),
		})
	return st
}

// ProgrammerNotFoundError is returned when the programmer is not found
type ProgrammerNotFoundError struct {
	Programmer string
	Cause      error
}

func (e *ProgrammerNotFoundError) Error() string {
	return composeErrorMsg(i18n.Tr("Programmer '%s' not found", e.Programmer), e.Cause)
}

func (e *ProgrammerNotFoundError) Unwrap() error {
	return e.Cause
}

// GRPCStatus converts the error into a *status.Status
func (e *ProgrammerNotFoundError) GRPCStatus() *status.Status {
	return status.New(codes.NotFound, e.Error())
}

// MonitorNotFoundError is returned when the pluggable monitor is not found
type MonitorNotFoundError struct {
	Monitor string
	Cause   error
}

func (e *MonitorNotFoundError) Error() string {
	return composeErrorMsg(i18n.Tr("Monitor '%s' not found", e.Monitor), e.Cause)
}

func (e *MonitorNotFoundError) Unwrap() error {
	return e.Cause
}

// GRPCStatus converts the error into a *status.Status
func (e *MonitorNotFoundError) GRPCStatus() *status.Status {
	return status.New(codes.NotFound, e.Error())
}

// InvalidPlatformPropertyError is returned when a property in the platform is not valid
type InvalidPlatformPropertyError struct {
	Property string
	Value    string
}

func (e *InvalidPlatformPropertyError) Error() string {
	return i18n.Tr("Invalid '%[1]s' property: %[2]s", e.Property, e.Value)
}

// GRPCStatus converts the error into a *status.Status
func (e *InvalidPlatformPropertyError) GRPCStatus() *status.Status {
	return status.New(codes.FailedPrecondition, e.Error())
}

// MissingPlatformPropertyError is returned when a property in the platform is not found
type MissingPlatformPropertyError struct {
	Property string
}

func (e *MissingPlatformPropertyError) Error() string {
	return i18n.Tr("Property '%s' is undefined", e.Property)
}

// GRPCStatus converts the error into a *status.Status
func (e *MissingPlatformPropertyError) GRPCStatus() *status.Status {
	return status.New(codes.FailedPrecondition, e.Error())
}

// PlatformNotFoundError is returned when a platform is not found
type PlatformNotFoundError struct {
	Platform string
	Cause    error
}

func (e *PlatformNotFoundError) Error() string {
	return composeErrorMsg(i18n.Tr("Platform '%s' not found", e.Platform), e.Cause)
}

// GRPCStatus converts the error into a *status.Status
func (e *PlatformNotFoundError) GRPCStatus() *status.Status {
	return status.New(codes.FailedPrecondition, e.Error())
}

func (e *PlatformNotFoundError) Unwrap() error {
	return e.Cause
}

// PlatformLoadingError is returned when a platform has fatal errors that prevents loading
type PlatformLoadingError struct {
	Cause error
}

func (e *PlatformLoadingError) Error() string {
	return composeErrorMsg(i18n.Tr("Error loading hardware platform"), e.Cause)
}

// GRPCStatus converts the error into a *status.Status
func (e *PlatformLoadingError) GRPCStatus() *status.Status {
	s, _ := status.New(codes.FailedPrecondition, e.Error()).
		WithDetails(&rpc.PlatformLoadingError{})
	return s
}

func (e *PlatformLoadingError) Unwrap() error {
	return e.Cause
}

// LibraryNotFoundError is returned when a platform is not found
type LibraryNotFoundError struct {
	Library string
	Cause   error
}

func (e *LibraryNotFoundError) Error() string {
	return composeErrorMsg(i18n.Tr("Library '%s' not found", e.Library), e.Cause)
}

// GRPCStatus converts the error into a *status.Status
func (e *LibraryNotFoundError) GRPCStatus() *status.Status {
	return status.New(codes.FailedPrecondition, e.Error())
}

func (e *LibraryNotFoundError) Unwrap() error {
	return e.Cause
}

// LibraryDependenciesResolutionFailedError is returned when an inconsistency is found in library dependencies
// or a solution cannot be found.
type LibraryDependenciesResolutionFailedError struct {
	Cause error
}

func (e *LibraryDependenciesResolutionFailedError) Error() string {
	return composeErrorMsg(i18n.Tr("No valid dependencies solution found"), e.Cause)
}

// GRPCStatus converts the error into a *status.Status
func (e *LibraryDependenciesResolutionFailedError) GRPCStatus() *status.Status {
	return status.New(codes.FailedPrecondition, e.Error())
}

func (e *LibraryDependenciesResolutionFailedError) Unwrap() error {
	return e.Cause
}

// PlatformAlreadyAtTheLatestVersionError is returned when a platform is up to date
type PlatformAlreadyAtTheLatestVersionError struct {
	Platform string
}

func (e *PlatformAlreadyAtTheLatestVersionError) Error() string {
	return i18n.Tr("Platform '%s' is already at the latest version", e.Platform)
}

// GRPCStatus converts the error into a *status.Status
func (e *PlatformAlreadyAtTheLatestVersionError) GRPCStatus() *status.Status {
	st, _ := status.
		New(codes.AlreadyExists, e.Error()).
		WithDetails(&rpc.AlreadyAtLatestVersionError{})
	return st
}

// MissingSketchPathError is returned when the sketch path is mandatory and not specified
type MissingSketchPathError struct{}

func (e *MissingSketchPathError) Error() string {
	return i18n.Tr("Missing sketch path")
}

// GRPCStatus converts the error into a *status.Status
func (e *MissingSketchPathError) GRPCStatus() *status.Status {
	return status.New(codes.InvalidArgument, e.Error())
}

// CantCreateSketchError is returned when the sketch cannot be created
type CantCreateSketchError struct {
	Cause error
}

func (e *CantCreateSketchError) Error() string {
	return composeErrorMsg(i18n.Tr("Can't create sketch"), e.Cause)
}

func (e *CantCreateSketchError) Unwrap() error {
	return e.Cause
}

// CantUpdateSketchError is returned when the sketch cannot be updated
type CantUpdateSketchError struct {
	Cause error
}

func (e *CantUpdateSketchError) Error() string {
	return composeErrorMsg(i18n.Tr("Can't update sketch"), e.Cause)
}

func (e *CantUpdateSketchError) Unwrap() error {
	return e.Cause
}

// CantOpenSketchError is returned when the sketch is not found or cannot be opened
type CantOpenSketchError struct {
	Cause error
}

func (e *CantOpenSketchError) Error() string {
	return composeErrorMsg(i18n.Tr("Can't open sketch"), e.Cause)
}

func (e *CantOpenSketchError) Unwrap() error {
	return e.Cause
}

// GRPCStatus converts the error into a *status.Status
func (e *CantOpenSketchError) GRPCStatus() *status.Status {
	return status.New(codes.NotFound, e.Error())
}

// FailedInstallError is returned if an install operation fails
type FailedInstallError struct {
	Message string
	Cause   error
}

func (e *FailedInstallError) Error() string {
	return composeErrorMsg(e.Message, e.Cause)
}

func (e *FailedInstallError) Unwrap() error {
	return e.Cause
}

// GRPCStatus converts the error into a *status.Status
func (e *FailedInstallError) GRPCStatus() *status.Status {
	return status.New(codes.Internal, e.Error())
}

// FailedLibraryInstallError is returned if a library install operation fails
type FailedLibraryInstallError struct {
	Cause error
}

func (e *FailedLibraryInstallError) Error() string {
	return composeErrorMsg(i18n.Tr("Library install failed"), e.Cause)
}

func (e *FailedLibraryInstallError) Unwrap() error {
	return e.Cause
}

// GRPCStatus converts the error into a *status.Status
func (e *FailedLibraryInstallError) GRPCStatus() *status.Status {
	return status.New(codes.Internal, e.Error())
}

// FailedUninstallError is returned if an uninstall operation fails
type FailedUninstallError struct {
	Message string
	Cause   error
}

func (e *FailedUninstallError) Error() string {
	return composeErrorMsg(e.Message, e.Cause)
}

func (e *FailedUninstallError) Unwrap() error {
	return e.Cause
}

// GRPCStatus converts the error into a *status.Status
func (e *FailedUninstallError) GRPCStatus() *status.Status {
	return status.New(codes.Internal, e.Error())
}

// FailedDownloadError is returned when a network download fails
type FailedDownloadError struct {
	Message string
	Cause   error
}

func (e *FailedDownloadError) Error() string {
	return composeErrorMsg(e.Message, e.Cause)
}

func (e *FailedDownloadError) Unwrap() error {
	return e.Cause
}

// GRPCStatus converts the error into a *status.Status
func (e *FailedDownloadError) GRPCStatus() *status.Status {
	return status.New(codes.Internal, e.Error())
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

// GRPCStatus converts the error into a *status.Status
func (e *FailedUploadError) GRPCStatus() *status.Status {
	return status.New(codes.Internal, e.Error())
}

// FailedDebugError is returned when the debug fails
type FailedDebugError struct {
	Message string
	Cause   error
}

func (e *FailedDebugError) Error() string {
	return composeErrorMsg(e.Message, e.Cause)
}

func (e *FailedDebugError) Unwrap() error {
	return e.Cause
}

// GRPCStatus converts the error into a *status.Status
func (e *FailedDebugError) GRPCStatus() *status.Status {
	return status.New(codes.Internal, e.Error())
}

// FailedMonitorError is returned when opening the monitor port of a board fails
type FailedMonitorError struct {
	Cause error
}

func (e *FailedMonitorError) Error() string {
	return composeErrorMsg(i18n.Tr("Port monitor error"), e.Cause)
}

func (e *FailedMonitorError) Unwrap() error {
	return e.Cause
}

// GRPCStatus converts the error into a *status.Status
func (e *FailedMonitorError) GRPCStatus() *status.Status {
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

// GRPCStatus converts the error into a *status.Status
func (e *CompileFailedError) GRPCStatus() *status.Status {
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

// GRPCStatus converts the error into a *status.Status
func (e *InvalidArgumentError) GRPCStatus() *status.Status {
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

// GRPCStatus converts the error into a *status.Status
func (e *NotFoundError) GRPCStatus() *status.Status {
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

// GRPCStatus converts the error into a *status.Status
func (e *PermissionDeniedError) GRPCStatus() *status.Status {
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

// GRPCStatus converts the error into a *status.Status
func (e *UnavailableError) GRPCStatus() *status.Status {
	return status.New(codes.Unavailable, e.Error())
}

// TempDirCreationFailedError is returned if a temp dir could not be created
type TempDirCreationFailedError struct {
	Cause error
}

func (e *TempDirCreationFailedError) Error() string {
	return composeErrorMsg(i18n.Tr("Cannot create temp dir"), e.Cause)
}

func (e *TempDirCreationFailedError) Unwrap() error {
	return e.Cause
}

// GRPCStatus converts the error into a *status.Status
func (e *TempDirCreationFailedError) GRPCStatus() *status.Status {
	return status.New(codes.Unavailable, e.Error())
}

// TempFileCreationFailedError is returned if a temp file could not be created
type TempFileCreationFailedError struct {
	Cause error
}

func (e *TempFileCreationFailedError) Error() string {
	return composeErrorMsg(i18n.Tr("Cannot create temp file"), e.Cause)
}

func (e *TempFileCreationFailedError) Unwrap() error {
	return e.Cause
}

// GRPCStatus converts the error into a *status.Status
func (e *TempFileCreationFailedError) GRPCStatus() *status.Status {
	return status.New(codes.Unavailable, e.Error())
}

// SignatureVerificationFailedError is returned if a signature verification fails
type SignatureVerificationFailedError struct {
	File  string
	Cause error
}

func (e *SignatureVerificationFailedError) Error() string {
	return composeErrorMsg(i18n.Tr("'%s' has an invalid signature", e.File), e.Cause)
}

func (e *SignatureVerificationFailedError) Unwrap() error {
	return e.Cause
}

// GRPCStatus converts the error into a *status.Status
func (e *SignatureVerificationFailedError) GRPCStatus() *status.Status {
	return status.New(codes.Unavailable, e.Error())
}

// MultiplePlatformsError is returned when trying to detect
// the Platform the user is trying to interact with and
// multiple results are found.
type MultiplePlatformsError struct {
	Platforms    []string
	UserPlatform string
}

func (e *MultiplePlatformsError) Error() string {
	return i18n.Tr("Found %d platforms matching \"%s\": %s",
		len(e.Platforms), e.UserPlatform, strings.Join(e.Platforms, ", "))
}

// GRPCStatus converts the error into a *status.Status
func (e *MultiplePlatformsError) GRPCStatus() *status.Status {
	return status.New(codes.InvalidArgument, e.Error())
}

// MultipleLibraryInstallDetected is returned when the user request an
// operation on a library but multiple libraries with the same name
// (in library.properties) are detected.
type MultipleLibraryInstallDetected struct {
	LibName string
	LibsDir paths.PathList
	Message string
}

func (e *MultipleLibraryInstallDetected) Error() string {
	res := i18n.Tr("The library %s has multiple installations:", e.LibName) + "\n"
	for _, lib := range e.LibsDir {
		res += fmt.Sprintf("- %s\n", lib)
	}
	res += e.Message
	return res
}

// GRPCStatus converts the error into a *status.Status
func (e *MultipleLibraryInstallDetected) GRPCStatus() *status.Status {
	return status.New(codes.InvalidArgument, e.Error())
}

// InstanceNeedsReinitialization
type InstanceNeedsReinitialization struct {
}

func (e *InstanceNeedsReinitialization) Error() string {
	return i18n.Tr("The instance is no longer valid and needs to be reinitialized")
}

// GRPCStatus converts the error into a *status.Status
func (e *InstanceNeedsReinitialization) GRPCStatus() *status.Status {
	st, _ := status.
		New(codes.InvalidArgument, e.Error()).
		WithDetails(&rpc.InstanceNeedsReinitializationError{})
	return st
}
