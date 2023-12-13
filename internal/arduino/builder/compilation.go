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

package builder

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
	"sync"

	"github.com/arduino/arduino-cli/internal/arduino/builder/internal/utils"
	"github.com/arduino/arduino-cli/internal/arduino/globals"
	"github.com/arduino/go-paths-helper"
)

func (b *Builder) compileFiles(
	sourceDir *paths.Path,
	buildPath *paths.Path,
	recurse bool,
	includes []string,
) (paths.PathList, error) {
	validExtensions := []string{}
	for ext := range globals.SourceFilesValidExtensions {
		validExtensions = append(validExtensions, ext)
	}

	sources, err := utils.FindFilesInFolder(sourceDir, recurse, validExtensions...)
	if err != nil {
		return nil, err
	}

	b.Progress.AddSubSteps(len(sources))
	defer b.Progress.RemoveSubSteps()

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
		if !b.buildProperties.ContainsKey(recipe) {
			recipe = fmt.Sprintf("recipe%s.o.pattern", globals.SourceFilesValidExtensions[source.Ext()])
		}
		objectFile, err := b.compileFileWithRecipe(sourceDir, source, buildPath, includes, recipe)
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
	if b.jobs == 0 {
		b.jobs = runtime.NumCPU()
	}
	for i := 0; i < b.jobs; i++ {
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

		b.Progress.CompleteStep()
	}
	close(queue)
	wg.Wait()
	if len(errorsList) > 0 {
		// output the first error
		return nil, errorsList[0]
	}
	objectFiles.Sort()
	return objectFiles, nil
}

// CompileFilesRecursive fixdoc
func (b *Builder) compileFileWithRecipe(
	sourcePath *paths.Path,
	source *paths.Path,
	buildPath *paths.Path,
	includes []string,
	recipe string,
) (*paths.Path, error) {
	properties := b.buildProperties.Clone()
	properties.Set("compiler.warning_flags", properties.Get("compiler.warning_flags."+b.logger.WarningsLevel()))
	properties.Set("includes", strings.Join(includes, " "))
	properties.SetPath("source_file", source)
	relativeSource, err := sourcePath.RelTo(source)
	if err != nil {
		return nil, err
	}
	depsFile := buildPath.Join(relativeSource.String() + ".d")
	objectFile := buildPath.Join(relativeSource.String() + ".o")

	properties.SetPath("object_file", objectFile)
	err = objectFile.Parent().MkdirAll()
	if err != nil {
		return nil, err
	}

	objIsUpToDate, err := utils.ObjFileIsUpToDate(source, objectFile, depsFile)
	if err != nil {
		return nil, err
	}

	command, err := b.prepareCommandForRecipe(properties, recipe, false)
	if err != nil {
		return nil, err
	}
	if b.compilationDatabase != nil {
		b.compilationDatabase.Add(source, command)
	}
	if !objIsUpToDate && !b.onlyUpdateCompilationDatabase {
		commandStdout, commandStderr := &bytes.Buffer{}, &bytes.Buffer{}
		command.RedirectStdoutTo(commandStdout)
		command.RedirectStderrTo(commandStderr)

		if b.logger.Verbose() {
			b.logger.Info(utils.PrintableCommand(command.GetArgs()))
		}
		// Since this compile could be multithreaded, we first capture the command output
		if err := command.Start(); err != nil {
			return nil, err
		}
		err := command.Wait()
		// and transfer all at once at the end...
		if b.logger.Verbose() {
			b.logger.WriteStdout(commandStdout.Bytes())
		}
		b.logger.WriteStderr(commandStderr.Bytes())

		// Parse the output of the compiler to gather errors and warnings...
		if b.compilerOutputParser != nil {
			b.compilerOutputParser(command.GetArgs(), commandStdout.Bytes())
			b.compilerOutputParser(command.GetArgs(), commandStderr.Bytes())
		}

		// ...and then return the error
		if err != nil {
			return nil, err
		}
	} else if b.logger.Verbose() {
		if objIsUpToDate {
			b.logger.Info(tr("Using previously compiled file: %[1]s", objectFile))
		} else {
			b.logger.Info(tr("Skipping compile of: %[1]s", objectFile))
		}
	}

	return objectFile, nil
}
