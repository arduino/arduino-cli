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

package lib

import (
	"context"
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/cli/output"
	"github.com/arduino/arduino-cli/commands/lib"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/spf13/cobra"
)

func initInstallCommand() *cobra.Command {
	installCommand := &cobra.Command{
		Use:   "install LIBRARY[@VERSION_NUMBER](S)",
		Short: "Installs one or more specified libraries into the system.",
		Long:  "Installs one or more specified libraries into the system.",
		Example: "" +
			"  " + os.Args[0] + " lib install AudioZero       # for the latest version.\n" +
			"  " + os.Args[0] + " lib install AudioZero@1.0.0 # for the specific version.",
		Args: cobra.MinimumNArgs(1),
		Run:  runInstallCommand,
	}
	installCommand.Flags().BoolVar(&installFlags.noDeps, "no-deps", false, "Do not install dependencies.")
	installCommand.Flags().BoolVar(&installFlags.gitURL, "git-url", false, "Enter git url for libraries hosted on repositories")
	installCommand.Flags().BoolVar(&installFlags.zipPath, "zip-path", false, "Enter a path to zip file")
	return installCommand
}

var installFlags struct {
	noDeps  bool
	gitURL  bool
	zipPath bool
}

func runInstallCommand(cmd *cobra.Command, args []string) {
	instance := instance.CreateInstanceIgnorePlatformIndexErrors()
	if installFlags.zipPath {
		ziplibraryInstallReq := &rpc.ZipLibraryInstallReq{
			Instance: instance,
			Path:     args[0],
		}
		err := lib.ZipLibraryInstall(context.Background(), ziplibraryInstallReq, output.TaskProgress())
		if err != nil {
			feedback.Errorf("Error installing Zip Library: %v", err)
			os.Exit(errorcodes.ErrGeneric)
		}
	} else if installFlags.gitURL {
		gitlibraryInstallReq := &rpc.GitLibraryInstallReq{
			Instance: instance,
			Url:      args[0],
		}
		err := lib.GitLibraryInstall(context.Background(), gitlibraryInstallReq, output.TaskProgress())
		if err != nil {
			feedback.Errorf("Error installing Git Library: %v", err)
			os.Exit(errorcodes.ErrGeneric)
		}
	} else {
		libRefs, err := ParseLibraryReferenceArgsAndAdjustCase(instance, args)
		if err != nil {
			feedback.Errorf("Arguments error: %v", err)
			os.Exit(errorcodes.ErrBadArgument)
		}

		toInstall := map[string]*rpc.LibraryDependencyStatus{}
		if installFlags.noDeps {
			for _, libRef := range libRefs {
				toInstall[libRef.Name] = &rpc.LibraryDependencyStatus{
					Name:            libRef.Name,
					VersionRequired: libRef.Version,
				}
			}
		} else {
			for _, libRef := range libRefs {
				depsResp, err := lib.LibraryResolveDependencies(context.Background(), &rpc.LibraryResolveDependenciesReq{
					Instance: instance,
					Name:     libRef.Name,
					Version:  libRef.Version,
				})
				if err != nil {
					feedback.Errorf("Error resolving dependencies for %s: %s", libRef, err)
					os.Exit(errorcodes.ErrGeneric)
				}
				for _, dep := range depsResp.GetDependencies() {
					feedback.Printf("%s depends on %s@%s", libRef, dep.GetName(), dep.GetVersionRequired())
					if existingDep, has := toInstall[dep.GetName()]; has {
						if existingDep.GetVersionRequired() != dep.GetVersionRequired() {
							// TODO: make a better error
							feedback.Errorf("The library %s is required in two different versions: %s and %s",
								dep.GetName(), dep.GetVersionRequired(), existingDep.GetVersionRequired())
							os.Exit(errorcodes.ErrGeneric)
						}
					}
					toInstall[dep.GetName()] = dep
				}
			}
		}

		for _, library := range toInstall {
			libraryInstallReq := &rpc.LibraryInstallReq{
				Instance: instance,
				Name:     library.Name,
				Version:  library.VersionRequired,
			}
			err := lib.LibraryInstall(context.Background(), libraryInstallReq, output.ProgressBar(), output.TaskProgress())
			if err != nil {
				feedback.Errorf("Error installing %s: %v", library, err)
				os.Exit(errorcodes.ErrGeneric)
			}
		}
	}
}
