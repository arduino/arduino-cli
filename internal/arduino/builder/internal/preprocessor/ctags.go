// This file is part of arduino-cli.
//
// Copyright 2023 ARDUINO SA (http://www.arduino.cc/)
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

package preprocessor

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/arduino/arduino-cli/internal/arduino/builder/cpp"
	"github.com/arduino/arduino-cli/internal/arduino/builder/internal/preprocessor/internal/ctags"
	"github.com/arduino/arduino-cli/internal/arduino/builder/internal/runner"
	"github.com/arduino/arduino-cli/internal/arduino/sketch"
	"github.com/arduino/arduino-cli/internal/i18n"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
)

// DebugPreprocessor when set to true the CTags preprocessor will output debugging info to stdout
// this is useful for unit-testing to provide more infos
var DebugPreprocessor bool

// PreprocessSketchWithCtags performs preprocessing of the arduino sketch using CTags.
func PreprocessSketchWithCtags(
	ctx context.Context,
	sketch *sketch.Sketch, buildPath *paths.Path, includes paths.PathList,
	lineOffset int, buildProperties *properties.Map,
	onlyUpdateCompilationDatabase, verbose bool,
) (*runner.Result, error) {
	// Create a temporary working directory
	tmpDir, err := paths.MkTempDir("", "")
	if err != nil {
		return nil, err
	}
	defer tmpDir.RemoveAll()
	ctagsTarget := tmpDir.Join("sketch_merged.cpp")

	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}

	// Run GCC preprocessor
	sourceFile := buildPath.Join("sketch", sketch.MainFile.Base()+".cpp")
	result, err := GCC(ctx, sourceFile, ctagsTarget, includes, buildProperties)
	stdout.Write(result.Stdout)
	stderr.Write(result.Stderr)
	if err != nil {
		if !onlyUpdateCompilationDatabase {
			return &runner.Result{Args: result.Args, Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}, err
		}

		// Do not bail out if we are generating the compile commands database
		fmt.Fprintf(stderr, "%s: %s",
			i18n.Tr("An error occurred adding prototypes"),
			i18n.Tr("the compilation database may be incomplete or inaccurate"))
		if err := sourceFile.CopyTo(ctagsTarget); err != nil {
			return &runner.Result{Args: result.Args, Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}, err
		}
	}

	if src, err := ctagsTarget.ReadFile(); err == nil {
		filteredSource := filterSketchSource(sketch, bytes.NewReader(src), false, stderr)
		if err := ctagsTarget.WriteFile([]byte(filteredSource)); err != nil {
			return &runner.Result{Args: result.Args, Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}, err
		}
	} else {
		return &runner.Result{Args: result.Args, Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}, err
	}

	// Run CTags on gcc-preprocessed source
	ctagsOutput, ctagsStdErr, ctagsCmdLine, err := RunCTags(ctx, ctagsTarget, buildProperties)
	if verbose {
		stdout.Write([]byte(ctagsCmdLine + "\n"))
		stderr.Write(ctagsStdErr)
	}
	if err != nil {
		return &runner.Result{Args: result.Args, Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}, err
	}

	// Parse CTags output
	parser := &ctags.Parser{}
	prototypes, firstFunctionLine := parser.Parse(ctagsOutput, sketch.MainFile)
	if firstFunctionLine == -1 {
		firstFunctionLine = 0
	}

	// Add prototypes to the original sketch source
	var source string
	if sourceData, err := sourceFile.ReadFile(); err == nil {
		source = string(sourceData)
	} else {
		return &runner.Result{Args: result.Args, Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}, err
	}
	source = strings.ReplaceAll(source, "\r\n", "\n")
	source = strings.ReplaceAll(source, "\r", "\n")
	sourceRows := strings.Split(source, "\n")
	if isFirstFunctionOutsideOfSource(firstFunctionLine, sourceRows) {
		return &runner.Result{Args: result.Args, Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}, nil
	}

	insertionLine := firstFunctionLine + lineOffset - 1
	firstFunctionChar := len(strings.Join(sourceRows[:insertionLine], "\n")) + 1
	prototypeSection := composePrototypeSection(firstFunctionLine, prototypes)
	preprocessedSource := source[:firstFunctionChar] + prototypeSection + source[firstFunctionChar:]

	if DebugPreprocessor {
		fmt.Println("#PREPROCESSED SOURCE")
		prototypesRows := strings.Split(prototypeSection, "\n")
		prototypesRows = prototypesRows[:len(prototypesRows)-1]
		for i := 0; i < len(sourceRows)+len(prototypesRows); i++ {
			if i < insertionLine {
				fmt.Printf("   |%s\n", sourceRows[i])
			} else if i < insertionLine+len(prototypesRows) {
				fmt.Printf("PRO|%s\n", prototypesRows[i-insertionLine])
			} else {
				fmt.Printf("   |%s\n", sourceRows[i-len(prototypesRows)])
			}
		}
		fmt.Println("#END OF PREPROCESSED SOURCE")
	}

	// Write back arduino-preprocess output to the sourceFile
	err = sourceFile.WriteFile([]byte(preprocessedSource))
	return &runner.Result{Args: result.Args, Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}, err
}

func composePrototypeSection(line int, prototypes []*ctags.Prototype) string {
	if len(prototypes) == 0 {
		return ""
	}

	str := joinPrototypes(prototypes)
	str += "\n#line "
	str += strconv.Itoa(line)
	str += " " + cpp.QuoteString(prototypes[0].File)
	str += "\n"

	return str
}

func joinPrototypes(prototypes []*ctags.Prototype) string {
	prototypesSlice := []string{}
	for _, proto := range prototypes {
		if signatureContainsaDefaultArg(proto) {
			continue
		}
		prototypesSlice = append(prototypesSlice, "#line "+strconv.Itoa(proto.Line)+" "+cpp.QuoteString(proto.File))
		prototypeParts := []string{}
		if proto.Modifiers != "" {
			prototypeParts = append(prototypeParts, proto.Modifiers)
		}
		prototypeParts = append(prototypeParts, proto.Prototype)
		prototypesSlice = append(prototypesSlice, strings.Join(prototypeParts, " "))
	}
	return strings.Join(prototypesSlice, "\n")
}

func signatureContainsaDefaultArg(proto *ctags.Prototype) bool {
	return strings.Contains(proto.Prototype, "=")
}

func isFirstFunctionOutsideOfSource(firstFunctionLine int, sourceRows []string) bool {
	return firstFunctionLine > len(sourceRows)-1
}

// RunCTags performs a run of ctags on the given source file. Returns the ctags output and the stderr contents.
func RunCTags(ctx context.Context, sourceFile *paths.Path, buildProperties *properties.Map) ([]byte, []byte, string, error) {
	ctagsBuildProperties := properties.NewMap()
	ctagsBuildProperties.Set("tools.ctags.path", "{runtime.tools.ctags.path}")
	ctagsBuildProperties.Set("tools.ctags.cmd.path", "{path}/ctags")
	ctagsBuildProperties.Set("tools.ctags.pattern", `"{cmd.path}" -u --language-force=c++ -f - --c++-kinds=svpf --fields=KSTtzns --line-directives "{source_file}"`)
	ctagsBuildProperties.Merge(buildProperties)
	ctagsBuildProperties.Merge(ctagsBuildProperties.SubTree("tools").SubTree("ctags"))
	ctagsBuildProperties.SetPath("source_file", sourceFile)

	pattern := ctagsBuildProperties.Get("pattern")
	if pattern == "" {
		return nil, nil, "", errors.New(i18n.Tr("%s pattern is missing", "ctags"))
	}

	commandLine := ctagsBuildProperties.ExpandPropsInString(pattern)
	parts, err := properties.SplitQuotedString(commandLine, `"'`, false)
	if err != nil {
		return nil, nil, "", err
	}
	proc, err := paths.NewProcess(nil, parts...)
	if err != nil {
		return nil, nil, "", err
	}
	stdout, stderr, err := proc.RunAndCaptureOutput(ctx)

	args := fmt.Sprintln(strings.Join(parts, " "))
	return stdout, stderr, args, err
}

func filterSketchSource(sketch *sketch.Sketch, source io.Reader, removeLineMarkers bool, stderr *bytes.Buffer) string {
	fileNames := paths.NewPathList()
	fileNames.Add(sketch.MainFile)
	fileNames.AddAll(sketch.OtherSketchFiles)

	inSketch := false
	filtered := ""

	scanner := bufio.NewScanner(source)

	var buf []byte
	const maxTokenSize = 1024 * 1024 // 1 MB
	scanner.Buffer(buf, maxTokenSize)

	for scanner.Scan() {
		line := scanner.Text()
		if filename := cpp.ParseLineMarker(line); filename != nil {
			inSketch = fileNames.Contains(filename)
			if inSketch && removeLineMarkers {
				continue
			}
		}

		if inSketch {
			filtered += line + "\n"
		}
	}
	if errors.Is(scanner.Err(), bufio.ErrTooLong) {
		fmt.Fprintf(stderr, "%s: %s",
			i18n.Tr("An error occurred adding prototypes"),
			i18n.Tr("line too long\n"))
	}

	return filtered
}
