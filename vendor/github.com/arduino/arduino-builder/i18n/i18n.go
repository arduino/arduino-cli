/*
 * This file is part of Arduino Builder.
 *
 * Arduino Builder is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2015 Arduino LLC (http://www.arduino.cc/)
 */

package i18n

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var PLACEHOLDER = regexp.MustCompile("{(\\d)}")

type Logger interface {
	Fprintln(w io.Writer, level string, format string, a ...interface{})
	UnformattedFprintln(w io.Writer, s string)
	UnformattedWrite(w io.Writer, data []byte)
	Println(level string, format string, a ...interface{})
	Name() string
	Flush() string
}

type NoopLogger struct{}

func (s NoopLogger) Fprintln(w io.Writer, level string, format string, a ...interface{}) {}

func (s NoopLogger) UnformattedFprintln(w io.Writer, str string) {}

func (s NoopLogger) UnformattedWrite(w io.Writer, data []byte) {}

func (s NoopLogger) Println(level string, format string, a ...interface{}) {}

func (s NoopLogger) Flush() string {
	return ""
}

func (s NoopLogger) Name() string {
	return "noop"
}

type AccumulatorLogger struct {
	Buffer *[]string
}

func (s AccumulatorLogger) Fprintln(w io.Writer, level string, format string, a ...interface{}) {
	*s.Buffer = append(*s.Buffer, Format(format, a...))
}

func (s AccumulatorLogger) UnformattedFprintln(w io.Writer, str string) {
	*s.Buffer = append(*s.Buffer, str)
}

func (s AccumulatorLogger) UnformattedWrite(w io.Writer, data []byte) {
	*s.Buffer = append(*s.Buffer, string(data))
}

func (s AccumulatorLogger) Println(level string, format string, a ...interface{}) {
	s.Fprintln(nil, level, format, a...)
}

func (s AccumulatorLogger) Flush() string {
	str := strings.Join(*s.Buffer, "\n")
	*s.Buffer = (*s.Buffer)[0:0]
	return str
}

func (s AccumulatorLogger) Name() string {
	return "accumulator"
}

type HumanTagsLogger struct{}

func (s HumanTagsLogger) Fprintln(w io.Writer, level string, format string, a ...interface{}) {
	format = "[" + level + "] " + format
	fprintln(w, Format(format, a...))
}

func (s HumanTagsLogger) Println(level string, format string, a ...interface{}) {
	s.Fprintln(os.Stdout, level, format, a...)
}

func (s HumanTagsLogger) UnformattedFprintln(w io.Writer, str string) {
	fprintln(w, str)
}

func (s HumanTagsLogger) UnformattedWrite(w io.Writer, data []byte) {
	write(w, data)
}

func (s HumanTagsLogger) Flush() string {
	return ""
}

func (s HumanTagsLogger) Name() string {
	return "humantags"
}

type HumanLogger struct{}

func (s HumanLogger) Fprintln(w io.Writer, level string, format string, a ...interface{}) {
	fprintln(w, Format(format, a...))
}

func (s HumanLogger) Println(level string, format string, a ...interface{}) {
	s.Fprintln(os.Stdout, level, format, a...)
}

func (s HumanLogger) UnformattedFprintln(w io.Writer, str string) {
	fprintln(w, str)
}

func (s HumanLogger) UnformattedWrite(w io.Writer, data []byte) {
	write(w, data)
}

func (s HumanLogger) Flush() string {
	return ""
}

func (s HumanLogger) Name() string {
	return "human"
}

type MachineLogger struct{}

func (s MachineLogger) Fprintln(w io.Writer, level string, format string, a ...interface{}) {
	printMachineFormattedLogLine(w, level, format, a)
}

func (s MachineLogger) Println(level string, format string, a ...interface{}) {
	printMachineFormattedLogLine(os.Stdout, level, format, a)
}

func (s MachineLogger) UnformattedFprintln(w io.Writer, str string) {
	fprintln(w, str)
}

func (s MachineLogger) Flush() string {
	return ""
}

func (s MachineLogger) Name() string {
	return "machine"
}

func (s MachineLogger) UnformattedWrite(w io.Writer, data []byte) {
	write(w, data)
}

func printMachineFormattedLogLine(w io.Writer, level string, format string, a []interface{}) {
	a = append([]interface{}(nil), a...)
	for idx, value := range a {
		typeof := reflect.Indirect(reflect.ValueOf(value)).Kind()
		if typeof == reflect.String {
			a[idx] = url.QueryEscape(value.(string))
		}
	}
	fprintf(w, "===%s ||| %s ||| %s\n", level, format, a)
}

var lock sync.Mutex

func fprintln(w io.Writer, s string) {
	lock.Lock()
	defer lock.Unlock()
	fmt.Fprintln(w, s)
}

func write(w io.Writer, data []byte) {
	lock.Lock()
	defer lock.Unlock()
	w.Write(data)
}

func fprintf(w io.Writer, format string, a ...interface{}) {
	lock.Lock()
	defer lock.Unlock()
	fmt.Fprintf(w, format, a...)
}

func FromJavaToGoSyntax(s string) string {
	submatches := PLACEHOLDER.FindAllStringSubmatch(s, -1)
	for _, submatch := range submatches {
		idx, err := strconv.Atoi(submatch[1])
		if err != nil {
			panic(err)
		}
		idx = idx + 1
		s = strings.Replace(s, submatch[0], "%["+strconv.Itoa(idx)+"]v", -1)
	}

	s = strings.Replace(s, "''", "'", -1)

	return s
}

func Format(format string, a ...interface{}) string {
	format = FromJavaToGoSyntax(format)
	message := fmt.Sprintf(format, a...)
	return message
}
