//
// Copyright 2014-2017 Cristian Maglie. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package serial // import "go.bug.st/serial.v1"

//sys regEnumValue(key syscall.Handle, index uint32, name *uint16, nameLen *uint32, reserved *uint32, class *uint16, value *uint16, valueLen *uint32) (regerrno error) = advapi32.RegEnumValueW

//sys getCommState(handle syscall.Handle, dcb *dcb) (err error) = GetCommState

//sys setCommState(handle syscall.Handle, dcb *dcb) (err error) = SetCommState

//sys setCommTimeouts(handle syscall.Handle, timeouts *commTimeouts) (err error) = SetCommTimeouts

//sys escapeCommFunction(handle syscall.Handle, function uint32) (res bool) = EscapeCommFunction

//sys getCommModemStatus(handle syscall.Handle, bits *uint32) (res bool) = GetCommModemStatus

//sys createEvent(eventAttributes *uint32, manualReset bool, initialState bool, name *uint16) (handle syscall.Handle, err error) = CreateEventW

//sys resetEvent(handle syscall.Handle) (err error) = ResetEvent

//sys getOverlappedResult(handle syscall.Handle, overlapEvent *syscall.Overlapped, n *uint32, wait bool) (err error) = GetOverlappedResult

//sys purgeComm(handle syscall.Handle, flags uint32) (err error) = PurgeComm

