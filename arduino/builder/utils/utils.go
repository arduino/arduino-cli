package utils

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"unicode"

	"github.com/arduino/arduino-cli/arduino/builder/compilation"
	"github.com/arduino/arduino-cli/arduino/builder/logger"
	"github.com/arduino/arduino-cli/arduino/builder/progress"
	"github.com/arduino/arduino-cli/arduino/globals"
	"github.com/arduino/arduino-cli/executils"
	"github.com/arduino/arduino-cli/i18n"
	f "github.com/arduino/arduino-cli/internal/algorithms"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var tr = i18n.Tr

// ObjFileIsUpToDate fixdoc
func ObjFileIsUpToDate(sourceFile, objectFile, dependencyFile *paths.Path) (bool, error) {
	logrus.Debugf("Checking previous results for %v (result = %v, dep = %v)", sourceFile, objectFile, dependencyFile)
	if objectFile == nil || dependencyFile == nil {
		logrus.Debugf("Not found: nil")
		return false, nil
	}

	sourceFile = sourceFile.Clean()
	sourceFileStat, err := sourceFile.Stat()
	if err != nil {
		return false, errors.WithStack(err)
	}

	objectFile = objectFile.Clean()
	objectFileStat, err := objectFile.Stat()
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Debugf("Not found: %v", objectFile)
			return false, nil
		}
		return false, errors.WithStack(err)
	}

	dependencyFile = dependencyFile.Clean()
	dependencyFileStat, err := dependencyFile.Stat()
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Debugf("Not found: %v", dependencyFile)
			return false, nil
		}
		return false, errors.WithStack(err)
	}

	if sourceFileStat.ModTime().After(objectFileStat.ModTime()) {
		logrus.Debugf("%v newer than %v", sourceFile, objectFile)
		return false, nil
	}
	if sourceFileStat.ModTime().After(dependencyFileStat.ModTime()) {
		logrus.Debugf("%v newer than %v", sourceFile, dependencyFile)
		return false, nil
	}

	rows, err := dependencyFile.ReadFileAsLines()
	if err != nil {
		return false, errors.WithStack(err)
	}

	rows = f.Map(rows, removeEndingBackSlash)
	rows = f.Map(rows, strings.TrimSpace)
	rows = f.Map(rows, unescapeDep)
	rows = f.Filter(rows, f.NotEquals(""))

	if len(rows) == 0 {
		return true, nil
	}

	firstRow := rows[0]
	if !strings.HasSuffix(firstRow, ":") {
		logrus.Debugf("No colon in first line of depfile")
		return false, nil
	}
	objFileInDepFile := firstRow[:len(firstRow)-1]
	if objFileInDepFile != objectFile.String() {
		logrus.Debugf("Depfile is about different file: %v", objFileInDepFile)
		return false, nil
	}

	// The first line of the depfile contains the path to the object file to generate.
	// The second line of the depfile contains the path to the source file.
	// All subsequent lines contain the header files necessary to compile the object file.

	// If we don't do this check it might happen that trying to compile a source file
	// that has the same name but a different path wouldn't recreate the object file.
	if sourceFile.String() != strings.Trim(rows[1], " ") {
		return false, nil
	}

	rows = rows[1:]
	for _, row := range rows {
		depStat, err := os.Stat(row)
		if err != nil && !os.IsNotExist(err) {
			// There is probably a parsing error of the dep file
			// Ignore the error and trigger a full rebuild anyway
			logrus.WithError(err).Debugf("Failed to read: %v", row)
			return false, nil
		}
		if os.IsNotExist(err) {
			logrus.Debugf("Not found: %v", row)
			return false, nil
		}
		if depStat.ModTime().After(objectFileStat.ModTime()) {
			logrus.Debugf("%v newer than %v", row, objectFile)
			return false, nil
		}
	}

	return true, nil
}

func removeEndingBackSlash(s string) string {
	return strings.TrimSuffix(s, "\\")
}

func unescapeDep(s string) string {
	s = strings.ReplaceAll(s, "\\ ", " ")
	s = strings.ReplaceAll(s, "\\\t", "\t")
	s = strings.ReplaceAll(s, "\\#", "#")
	s = strings.ReplaceAll(s, "$$", "$")
	s = strings.ReplaceAll(s, "\\\\", "\\")
	return s
}

// NormalizeUTF8 byte slice
// TODO: use it more often troughout all the project (maybe on logger interface?)
func NormalizeUTF8(buf []byte) []byte {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.Bytes(t, buf)
	return result
}

var sourceControlFolders = map[string]bool{"CVS": true, "RCS": true, ".git": true, ".github": true, ".svn": true, ".hg": true, ".bzr": true, ".vscode": true, ".settings": true, ".pioenvs": true, ".piolibdeps": true}

// filterOutSCCS is a ReadDirFilter that excludes known VSC or project files
func filterOutSCCS(file *paths.Path) bool {
	return !sourceControlFolders[file.Base()]
}

// filterReadableFiles is a ReadDirFilter that accepts only readable files
func filterReadableFiles(file *paths.Path) bool {
	// See if the file is readable by opening it
	f, err := file.Open()
	if err != nil {
		return false
	}
	f.Close()
	return true
}

// filterOutHiddenFiles is a ReadDirFilter that exclude files with a "." prefix in their name
var filterOutHiddenFiles = paths.FilterOutPrefixes(".")

// FindFilesInFolder fixdoc
func FindFilesInFolder(dir *paths.Path, recurse bool, extensions ...string) (paths.PathList, error) {
	fileFilter := paths.AndFilter(
		filterOutHiddenFiles,
		filterOutSCCS,
		paths.FilterOutDirectories(),
		filterReadableFiles,
	)
	if len(extensions) > 0 {
		fileFilter = paths.AndFilter(
			paths.FilterSuffixes(extensions...),
			fileFilter,
		)
	}
	if recurse {
		dirFilter := paths.AndFilter(
			filterOutHiddenFiles,
			filterOutSCCS,
		)
		return dir.ReadDirRecursiveFiltered(dirFilter, fileFilter)
	}
	return dir.ReadDir(fileFilter)
}

// nolint
const (
	Ignore        = 0 // Redirect to null
	Show          = 1 // Show on stdout/stderr as normal
	ShowIfVerbose = 2 // Show if verbose is set, Ignore otherwise
	Capture       = 3 // Capture into buffer
)

func printableArgument(arg string) string {
	if strings.ContainsAny(arg, "\"\\ \t") {
		arg = strings.ReplaceAll(arg, "\\", "\\\\")
		arg = strings.ReplaceAll(arg, "\"", "\\\"")
		return "\"" + arg + "\""
	}
	return arg
}

// Convert a command and argument slice back to a printable string.
// This adds basic escaping which is sufficient for debug output, but
// probably not for shell interpretation. This essentially reverses
// ParseCommandLine.
func printableCommand(parts []string) string {
	return strings.Join(f.Map(parts, printableArgument), " ")
}

// ExecCommand fixdoc
func ExecCommand(
	verbose bool,
	stdoutWriter, stderrWriter io.Writer,
	command *executils.Process, stdout int, stderr int,
) ([]byte, []byte, []byte, error) {
	verboseInfoBuf := &bytes.Buffer{}
	if verbose {
		verboseInfoBuf.WriteString(printableCommand(command.GetArgs()))
	}

	stdoutBuffer := &bytes.Buffer{}
	if stdout == Capture {
		command.RedirectStdoutTo(stdoutBuffer)
	} else if stdout == Show || (stdout == ShowIfVerbose && verbose) {
		if stdoutWriter != nil {
			command.RedirectStdoutTo(stdoutWriter)
		} else {
			command.RedirectStdoutTo(os.Stdout)
		}
	}

	stderrBuffer := &bytes.Buffer{}
	if stderr == Capture {
		command.RedirectStderrTo(stderrBuffer)
	} else if stderr == Show || (stderr == ShowIfVerbose && verbose) {
		if stderrWriter != nil {
			command.RedirectStderrTo(stderrWriter)
		} else {
			command.RedirectStderrTo(os.Stderr)
		}
	}

	err := command.Start()
	if err != nil {
		return verboseInfoBuf.Bytes(), nil, nil, errors.WithStack(err)
	}

	err = command.Wait()
	return verboseInfoBuf.Bytes(), stdoutBuffer.Bytes(), stderrBuffer.Bytes(), errors.WithStack(err)
}

// DirContentIsOlderThan DirContentIsOlderThan returns true if the content of the given directory is
// older than target file. If extensions are given, only the files with these
// extensions are tested.
func DirContentIsOlderThan(dir *paths.Path, target *paths.Path, extensions ...string) (bool, error) {
	targetStat, err := target.Stat()
	if err != nil {
		return false, err
	}
	targetModTime := targetStat.ModTime()

	files, err := FindFilesInFolder(dir, true, extensions...)
	if err != nil {
		return false, err
	}
	for _, file := range files {
		file, err := file.Stat()
		if err != nil {
			return false, err
		}
		if file.ModTime().After(targetModTime) {
			return false, nil
		}
	}
	return true, nil
}

// PrepareCommandForRecipe fixdoc
func PrepareCommandForRecipe(buildProperties *properties.Map, recipe string, removeUnsetProperties bool) (*executils.Process, error) {
	pattern := buildProperties.Get(recipe)
	if pattern == "" {
		return nil, errors.Errorf(tr("%[1]s pattern is missing"), recipe)
	}

	commandLine := buildProperties.ExpandPropsInString(pattern)
	if removeUnsetProperties {
		commandLine = properties.DeleteUnexpandedPropsFromString(commandLine)
	}

	parts, err := properties.SplitQuotedString(commandLine, `"'`, false)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// if the overall commandline is too long for the platform
	// try reducing the length by making the filenames relative
	// and changing working directory to build.path
	var relativePath string
	if len(commandLine) > 30000 {
		relativePath = buildProperties.Get("build.path")
		for i, arg := range parts {
			if _, err := os.Stat(arg); os.IsNotExist(err) {
				continue
			}
			rel, err := filepath.Rel(relativePath, arg)
			if err == nil && !strings.Contains(rel, "..") && len(rel) < len(arg) {
				parts[i] = rel
			}
		}
	}

	command, err := executils.NewProcess(nil, parts...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if relativePath != "" {
		command.SetDir(relativePath)
	}

	return command, nil
}

// CompileFiles fixdoc
func CompileFiles(
	sourceDir, buildPath *paths.Path,
	buildProperties *properties.Map,
	includes []string,
	onlyUpdateCompilationDatabase bool,
	compilationDatabase *compilation.Database,
	jobs int,
	builderLogger *logger.BuilderLogger,
	progress *progress.Struct, progressCB rpc.TaskProgressCB,
) (paths.PathList, error) {
	return compileFiles(
		onlyUpdateCompilationDatabase,
		compilationDatabase,
		jobs,
		sourceDir,
		false,
		buildPath, buildProperties, includes,
		builderLogger,
		progress, progressCB,
	)
}

// CompileFilesRecursive fixdoc
func CompileFilesRecursive(
	sourceDir, buildPath *paths.Path,
	buildProperties *properties.Map,
	includes []string,
	onlyUpdateCompilationDatabase bool,
	compilationDatabase *compilation.Database,
	jobs int,
	builderLogger *logger.BuilderLogger,
	progress *progress.Struct, progressCB rpc.TaskProgressCB,
) (paths.PathList, error) {
	return compileFiles(
		onlyUpdateCompilationDatabase,
		compilationDatabase,
		jobs,
		sourceDir,
		true,
		buildPath, buildProperties, includes,
		builderLogger,
		progress, progressCB,
	)
}

func compileFiles(
	onlyUpdateCompilationDatabase bool,
	compilationDatabase *compilation.Database,
	jobs int,
	sourceDir *paths.Path,
	recurse bool,
	buildPath *paths.Path,
	buildProperties *properties.Map,
	includes []string,
	builderLogger *logger.BuilderLogger,
	progress *progress.Struct,
	progressCB rpc.TaskProgressCB,
) (paths.PathList, error) {
	validExtensions := []string{}
	for ext := range globals.SourceFilesValidExtensions {
		validExtensions = append(validExtensions, ext)
	}

	sources, err := FindFilesInFolder(sourceDir, recurse, validExtensions...)
	if err != nil {
		return nil, err
	}

	progress.AddSubSteps(len(sources))
	defer progress.RemoveSubSteps()

	objectFiles := paths.NewPathList()
	var objectFilesMux sync.Mutex
	if len(sources) == 0 {
		return objectFiles, nil
	}
	var errorsList []error
	var errorsMux sync.Mutex

	queue := make(chan *paths.Path)
	job := func(source *paths.Path) {
		recipe := fmt.Sprintf("recipe%s.o.pattern", source.Ext())
		if !buildProperties.ContainsKey(recipe) {
			recipe = fmt.Sprintf("recipe%s.o.pattern", globals.SourceFilesValidExtensions[source.Ext()])
		}
		objectFile, verboseInfo, verboseStdout, stderr, err := compileFileWithRecipe(
			compilationDatabase,
			onlyUpdateCompilationDatabase,
			sourceDir, source, buildPath, buildProperties, includes, recipe,
			builderLogger,
		)
		if builderLogger.Verbose() {
			builderLogger.WriteStdout(verboseStdout)
			builderLogger.Info(string(verboseInfo))
		}
		builderLogger.WriteStderr(stderr)
		if err != nil {
			errorsMux.Lock()
			errorsList = append(errorsList, err)
			errorsMux.Unlock()
		} else {
			objectFilesMux.Lock()
			objectFiles.Add(objectFile)
			objectFilesMux.Unlock()
		}
	}

	// Spawn jobs runners
	var wg sync.WaitGroup
	if jobs == 0 {
		jobs = runtime.NumCPU()
	}
	for i := 0; i < jobs; i++ {
		wg.Add(1)
		go func() {
			for source := range queue {
				job(source)
			}
			wg.Done()
		}()
	}

	// Feed jobs until error or done
	for _, source := range sources {
		errorsMux.Lock()
		gotError := len(errorsList) > 0
		errorsMux.Unlock()
		if gotError {
			break
		}
		queue <- source

		progress.CompleteStep()
		// PushProgress
		if progressCB != nil {
			progressCB(&rpc.TaskProgress{
				Percent:   progress.Progress,
				Completed: progress.Progress >= 100.0,
			})
		}
	}
	close(queue)
	wg.Wait()
	if len(errorsList) > 0 {
		// output the first error
		return nil, errors.WithStack(errorsList[0])
	}
	objectFiles.Sort()
	return objectFiles, nil
}

func compileFileWithRecipe(
	compilationDatabase *compilation.Database,
	onlyUpdateCompilationDatabase bool,
	sourcePath *paths.Path,
	source *paths.Path,
	buildPath *paths.Path,
	buildProperties *properties.Map,
	includes []string,
	recipe string,
	builderLogger *logger.BuilderLogger,
) (*paths.Path, []byte, []byte, []byte, error) {
	verboseStdout, verboseInfo, errOut := &bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{}

	properties := buildProperties.Clone()
	properties.Set("compiler.warning_flags", properties.Get("compiler.warning_flags."+builderLogger.WarningsLevel()))
	properties.Set("includes", strings.Join(includes, " "))
	properties.SetPath("source_file", source)
	relativeSource, err := sourcePath.RelTo(source)
	if err != nil {
		return nil, nil, nil, nil, errors.WithStack(err)
	}
	depsFile := buildPath.Join(relativeSource.String() + ".d")
	objectFile := buildPath.Join(relativeSource.String() + ".o")

	properties.SetPath("object_file", objectFile)
	err = objectFile.Parent().MkdirAll()
	if err != nil {
		return nil, nil, nil, nil, errors.WithStack(err)
	}

	objIsUpToDate, err := ObjFileIsUpToDate(source, objectFile, depsFile)
	if err != nil {
		return nil, nil, nil, nil, errors.WithStack(err)
	}

	command, err := PrepareCommandForRecipe(properties, recipe, false)
	if err != nil {
		return nil, nil, nil, nil, errors.WithStack(err)
	}
	if compilationDatabase != nil {
		compilationDatabase.Add(source, command)
	}
	if !objIsUpToDate && !onlyUpdateCompilationDatabase {
		// Since this compile could be multithreaded, we first capture the command output
		info, stdout, stderr, err := ExecCommand(
			builderLogger.Verbose(),
			builderLogger.Stdout(),
			builderLogger.Stderr(),
			command,
			Capture,
			Capture,
		)
		// and transfer all at once at the end...
		if builderLogger.Verbose() {
			verboseInfo.Write(info)
			verboseStdout.Write(stdout)
		}
		errOut.Write(stderr)

		// ...and then return the error
		if err != nil {
			return nil, verboseInfo.Bytes(), verboseStdout.Bytes(), errOut.Bytes(), errors.WithStack(err)
		}
	} else if builderLogger.Verbose() {
		if objIsUpToDate {
			verboseInfo.WriteString(tr("Using previously compiled file: %[1]s", objectFile))
		} else {
			verboseInfo.WriteString(tr("Skipping compile of: %[1]s", objectFile))
		}
	}

	return objectFile, verboseInfo.Bytes(), verboseStdout.Bytes(), errOut.Bytes(), nil
}

// ArchiveCompiledFiles fixdoc
func ArchiveCompiledFiles(
	buildPath *paths.Path, archiveFile *paths.Path, objectFilesToArchive paths.PathList, buildProperties *properties.Map,
	onlyUpdateCompilationDatabase, verbose bool,
	stdoutWriter, stderrWriter io.Writer,
) (*paths.Path, []byte, error) {
	verboseInfobuf := &bytes.Buffer{}
	archiveFilePath := buildPath.JoinPath(archiveFile)

	if onlyUpdateCompilationDatabase {
		if verbose {
			verboseInfobuf.WriteString(tr("Skipping archive creation of: %[1]s", archiveFilePath))
		}
		return archiveFilePath, verboseInfobuf.Bytes(), nil
	}

	if archiveFileStat, err := archiveFilePath.Stat(); err == nil {
		rebuildArchive := false
		for _, objectFile := range objectFilesToArchive {
			objectFileStat, err := objectFile.Stat()
			if err != nil || objectFileStat.ModTime().After(archiveFileStat.ModTime()) {
				// need to rebuild the archive
				rebuildArchive = true
				break
			}
		}

		// something changed, rebuild the core archive
		if rebuildArchive {
			if err := archiveFilePath.Remove(); err != nil {
				return nil, nil, errors.WithStack(err)
			}
		} else {
			if verbose {
				verboseInfobuf.WriteString(tr("Using previously compiled file: %[1]s", archiveFilePath))
			}
			return archiveFilePath, verboseInfobuf.Bytes(), nil
		}
	}

	for _, objectFile := range objectFilesToArchive {
		properties := buildProperties.Clone()
		properties.Set("archive_file", archiveFilePath.Base())
		properties.SetPath("archive_file_path", archiveFilePath)
		properties.SetPath("object_file", objectFile)

		command, err := PrepareCommandForRecipe(properties, "recipe.ar.pattern", false)
		if err != nil {
			return nil, verboseInfobuf.Bytes(), errors.WithStack(err)
		}

		verboseInfo, _, _, err := ExecCommand(verbose, stdoutWriter, stderrWriter, command, ShowIfVerbose /* stdout */, Show /* stderr */)
		if verbose {
			verboseInfobuf.WriteString(string(verboseInfo))
		}
		if err != nil {
			return nil, verboseInfobuf.Bytes(), errors.WithStack(err)
		}
	}

	return archiveFilePath, verboseInfobuf.Bytes(), nil
}
