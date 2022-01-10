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

package i18n

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var PLACEHOLDER = regexp.MustCompile(`{(\d)}`)

type Logger interface {
	Fprintln(w io.Writer, level string, format string, a ...interface{})
	Println(level string, format string, a ...interface{})
	Name() string
	Flush() string
}

type LoggerToCustomStreams struct {
	Stdout io.Writer
	Stderr io.Writer
	mux    sync.Mutex
}

func (s *LoggerToCustomStreams) Fprintln(w io.Writer, level string, format string, a ...interface{}) {
	s.mux.Lock()
	defer s.mux.Unlock()
	target := s.Stdout
	if w == os.Stderr {
		target = s.Stderr
	}
	fmt.Fprintln(target, Format(format, a...))
}

func (s *LoggerToCustomStreams) Println(level string, format string, a ...interface{}) {
	s.Fprintln(nil, level, format, a...)
}

func (s *LoggerToCustomStreams) Flush() string {
	return ""
}

func (s *LoggerToCustomStreams) Name() string {
	return "LoggerToCustomStreams"
}

type NoopLogger struct{}

func (s NoopLogger) Fprintln(w io.Writer, level string, format string, a ...interface{}) {}

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

func (s MachineLogger) Flush() string {
	return ""
}

func (s MachineLogger) Name() string {
	return "machine"
}

func printMachineFormattedLogLine(w io.Writer, level string, format string, a []interface{}) {
	a = append([]interface{}(nil), a...)
	for idx, value := range a {
		if str, ok := value.(string); ok {
			a[idx] = url.QueryEscape(str)
		} else if stringer, ok := value.(fmt.Stringer); ok {
			a[idx] = url.QueryEscape(stringer.String())
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
