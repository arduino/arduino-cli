/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package sketch

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/arduino/arduino-cli/auth"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/arduino-cli/common/formatter/output"
	"github.com/arduino/arduino-cli/create_client_helpers"
	"github.com/arduino/go-paths-helper"
	"github.com/bcmi-labs/arduino-modules/sketches"
	"github.com/bgentry/go-netrc/netrc"
	"github.com/briandowns/spinner"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	priorityPullRemote = "pull-remote"
	priorityPushLocal  = "push-local"
	prioritySkip       = "skip"
)

func initSyncCommand() *cobra.Command {
	syncCommand := &cobra.Command{
		Use:     "sync",
		Short:   "Arduino CLI Sketch Commands.",
		Long:    "Arduino CLI Sketch Commands.",
		Example: "  " + commands.AppName + " sketch sync",
		Args:    cobra.NoArgs,
		Run:     runSyncCommand,
	}
	usage := "The decision made by default on conflicting sketches. Can be push-local, pull-remote, skip, ask-once, ask-always."
	syncCommand.Flags().StringVar(&syncFlags.priority, "conflict-policy", prioritySkip, usage)
	return syncCommand
}

var syncFlags struct {
	priority string // The decisive resource when we have conflicts. Can be local, remote, skip-conflict.
}

func runSyncCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino sketch sync`")

	sketchbook := commands.Config.SketchbookDir
	isTextMode := formatter.IsCurrentFormat("text")

	logrus.Info("Setting priority")
	priority := syncFlags.priority

	if priority == "ask-once" {
		if !isTextMode {
			formatter.PrintErrorMessage("Ask mode for this command is only supported using text format.")
			os.Exit(commands.ErrBadCall)
		}
		firstAsk := true
		for priority != priorityPullRemote &&
			priority != priorityPushLocal &&
			priority != prioritySkip {
			if !firstAsk {
				formatter.Print("Invalid option: " + priority)
			}
			formatter.Print("What should I do when I detect a conflict? [pull-remote | push-local | skip]")
			fmt.Scanln(&priority)
			firstAsk = false
		}
	}

	logrus.Infof("Priority set to %s", priority)

	logrus.Info("Preparing")

	var loader *spinner.Spinner

	if isTextMode && !commands.GlobalFlags.Debug {
		loader = spinner.New(spinner.CharSets[27], 100*time.Millisecond)
		loader.Prefix = "Syncing Sketches... "

		loader.Start()
	}

	stopSpinner := func() {
		if isTextMode && !commands.GlobalFlags.Debug {
			loader.Stop()
		}
	}

	startSpinner := func() {
		if isTextMode && !commands.GlobalFlags.Debug {
			loader.Start()
		}
	}

	logrus.Info("Logging in")
	username, bearerToken, err := login()
	if err != nil {
		stopSpinner()
		formatter.PrintError(err, "Cannot login")
		os.Exit(commands.ErrNetwork)
	}

	logrus.Info("Finding local sketches")
	sketchMap := sketches.Find(sketchbook.String(), "libraries") // Exclude libraries dirs.

	logrus.Info("Finding online sketches")
	client := createclient.New(nil)
	tok := "Bearer " + bearerToken
	resp, err := client.SearchSketches(context.Background(), createclient.SearchSketchesPath(), nil, &username, &tok)
	if err != nil {
		stopSpinner()
		formatter.PrintError(err, "Cannot get create sketches, sync failed.")
		os.Exit(commands.ErrNetwork)
	}
	defer resp.Body.Close()

	onlineSketches, err := client.DecodeArduinoCreateSketches(resp)
	if err != nil {
		stopSpinner()
		formatter.PrintError(err, "Cannot unmarshal response from create, sync failed.")
		os.Exit(commands.ErrGeneric)
	}

	onlineSketchesMap := make(map[string]*createclient.ArduinoCreateSketch, len(onlineSketches.Sketches))
	for _, item := range onlineSketches.Sketches {
		onlineSketchesMap[*item.Name] = item
	}

	maxLength := len(sketchMap) + len(onlineSketchesMap)

	logrus.Info("Syncing sketches")
	// Create output result struct with empty arrays.
	result := output.SketchSyncResult{
		PushedSketches:  make([]string, 0, maxLength),
		PulledSketches:  make([]string, 0, maxLength),
		SkippedSketches: make([]string, 0, maxLength),
		Errors:          make([]output.SketchSyncError, 0, maxLength),
	}

	for _, item := range sketchMap {
		itemOnline, hasConflict := onlineSketchesMap[item.Name]
		if hasConflict {
			logrus.Warnf("Conflict found for sketch `%s`", item.Name)
			item.ID = itemOnline.ID.String()
			// Resolve conflicts.
			if priority == "ask-always" {
				stopSpinner()

				logrus.Warn("Asking user what to do")
				if !isTextMode {
					logrus.WithField("format", commands.GlobalFlags.Format).Error("ask mode for this command is only supported using text format")
					formatter.PrintErrorMessage("ask mode for this command is only supported using text format.")
					os.Exit(commands.ErrBadCall)
				}

				firstAsk := true
				for priority != priorityPullRemote &&
					priority != priorityPushLocal &&
					priority != prioritySkip {
					if !firstAsk {
						formatter.Print("Invalid option: " + priority)
					}
					formatter.Print(fmt.Sprintf("Conflict detected for `%s` sketch, what should I do? [pull-remote | push-local | skip]", item.Name))
					fmt.Scanln(&priority)
					firstAsk = false
				}
				logrus.Warnf("Decision has been taken: %s", priority)

				startSpinner()
			}
			switch priority {
			case priorityPushLocal:
				logrus.Infof("Pushing local sketch `%s` as edit", item.Name)
				err := editSketch(*item, sketchbook, bearerToken)
				if err != nil {
					logrus.WithError(err).Warnf("Cannot push `%s`", item.Name)
					result.Errors = append(result.Errors, output.SketchSyncError{
						Sketch: item.Name,
						Error:  err,
					})
				} else {
					logrus.Infof("`%s` pushed", item.Name)
					result.PushedSketches = append(result.PushedSketches, item.Name)
				}
				break
			case priorityPullRemote:
				logrus.Infof("Pulling remote sketch `%s`", item.Name)
				err := pullSketch(itemOnline, sketchbook, bearerToken)
				if err != nil {
					logrus.WithError(err).Warnf("Cannot pull `%s`", item.Name)
					result.Errors = append(result.Errors, output.SketchSyncError{
						Sketch: item.Name,
						Error:  err,
					})
				} else {
					logrus.Infof("`%s` pulled", item.Name)
					result.PulledSketches = append(result.PulledSketches, item.Name)
				}
				break
			case prioritySkip:
				logrus.Warnf("Skipping `%s`", item.Name)
				result.SkippedSketches = append(result.SkippedSketches, item.Name)
				break
			default:
				logrus.Warnf("Skipping by default `%s`", item.Name)
				priority = prioritySkip
				result.SkippedSketches = append(result.SkippedSketches, item.Name)
			}
		} else { // Only local, push.
			logrus.Infof("No conflict, pushing `%s` as new sketch", item.Name)
			err := pushSketch(*item, sketchbook, bearerToken)
			if err != nil {
				logrus.WithError(err).Warnf("Cannot push `%s`", item.Name)
				result.Errors = append(result.Errors, output.SketchSyncError{
					Sketch: item.Name,
					Error:  err,
				})
			} else {
				logrus.Infof("`%s` pushed", item.Name)
				result.PushedSketches = append(result.PushedSketches, item.Name)
			}
		}
	}
	for _, item := range onlineSketches.Sketches {
		if sketchMap[*item.Name] != nil {
			continue
		}
		// Only online, pull.
		logrus.Infof("Pulling only online sketch `%s`", *item.Name)
		err := pullSketch(item, sketchbook, bearerToken)
		if err != nil {
			logrus.WithError(err).Warnf("Cannot pull `%s`", *item.Name)
			result.Errors = append(result.Errors, output.SketchSyncError{
				Sketch: *item.Name,
				Error:  err,
			})
		} else {
			logrus.Infof("`%s` pulled", *item.Name)
			result.PulledSketches = append(result.PulledSketches, *item.Name)
		}
	}

	stopSpinner()
	formatter.Print(result)
	logrus.Info("Done")
}

func pushSketch(sketch sketches.Sketch, sketchbook *paths.Path, bearerToken string) error {
	client := createclient.New(nil)

	resp, err := client.CreateSketches(context.Background(), createclient.CreateSketchesPath(), createclient.ConvertFrom(sketch), "Bearer "+bearerToken)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		errorMsg, err := client.DecodeErrorResponse(resp)
		if err != nil {
			return errors.New(resp.Status)
		}
		return errorMsg
	}
	_, err = client.DecodeArduinoCreateSketch(resp)
	if err != nil {
		return err
	}

	return nil
}

func editSketch(sketch sketches.Sketch, sketchbook *paths.Path, bearerToken string) error {
	client := createclient.New(nil)
	resp, err := client.EditSketches(
		context.Background(),
		createclient.EditSketchesPath(sketch.ID),
		createclient.ConvertFrom(sketch),
		"Bearer "+bearerToken)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		errorMsg, err := client.DecodeErrorResponse(resp)
		if err != nil {
			return errors.New(resp.Status)
		}
		return errorMsg
	}
	_, err = client.DecodeArduinoCreateSketch(resp)
	if err != nil {
		return err
	}

	return nil
}

func pullSketch(sketch *createclient.ArduinoCreateSketch, sketchbook *paths.Path, bearerToken string) error {
	client := createclient.New(nil)
	bearer := "Bearer " + bearerToken

	resp, err := client.ShowSketches(context.Background(), createclient.ShowSketchesPath(fmt.Sprint(sketch.ID)), &bearer)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		errorMsg, err := client.DecodeErrorResponse(resp)
		if err != nil {
			return errors.New(resp.Status)
		}
		return errorMsg
	}

	r, err := client.DecodeArduinoCreateSketch(resp)
	if err != nil {
		return err
	}

	sketchDir, err := sketchbook.MkTempDir(fmt.Sprintf("%s-temp", *sketch.Name))
	if err != nil {
		return err
	}
	defer sketchDir.RemoveAll()

	destDir := sketchbook.Join(*sketch.Name)

	for _, file := range append(r.Files, sketch.Ino) {
		path := findPathOf(*sketch.Name, *file.Path)

		err = os.MkdirAll(filepath.Dir(path), 0755)
		if err != nil {
			return err
		}

		resp, err = client.ShowFiles(context.Background(), createclient.ShowFilesPath("sketch", sketch.ID.String(), path))
		if err != nil {
			return err
		}
		filewithData, err := client.DecodeArduinoCreateFile(resp)
		if err != nil {
			if resp.StatusCode != 200 {
				errResp, err := client.DecodeErrorResponse(resp)
				if err != nil {
					return errors.New(resp.Status)
				}
				return errResp
			}
		}

		decodedData, err := base64.StdEncoding.DecodeString(*filewithData.Data)
		if err != nil {
			return err
		}

		destFile := sketchDir.Join(path)
		err = destFile.WriteFile(decodedData)
		if err != nil {
			return errors.New("copy of a file of the downloaded sketch failed, sync failed")
		}
	}

	if err := destDir.RemoveAll(); err != nil {
		return err
	}

	err = sketchDir.Rename(destDir)
	if err != nil {
		return err
	}
	return nil
}

func findPathOf(sketchName string, path string) string {
	list := strings.Split(path, "/")

	for i := len(list) - 1; i > -1; i-- {
		//fmt.Println(list[i], "==", sketchName, "?", list[i] == sketchName)
		if list[i] == sketchName {
			return filepath.Join(list[i+1:]...)
		}
	}
	return ""
}

func login() (string, string, error) {
	authConf := auth.New()

	home, err := homedir.Dir()
	if err != nil {
		return "", "", err
	}

	logrus.Info("Reading ~/.netrc file")
	netRCFile := filepath.Join(home, ".netrc")
	file, err := os.OpenFile(netRCFile, os.O_CREATE|os.O_RDONLY, 0666)
	if err != nil {
		logrus.WithError(err).Error("Cannot read ~/.netrc file")
		return "", "", err
	}
	NetRC, err := netrc.Parse(file)
	if err != nil {
		logrus.WithError(err).Error("Cannot parse ~/.netrc file")
		return "", "", err
	}

	logrus.Info("Searching for user credentials into the ~/.netrc file")
	arduinoMachine := NetRC.FindMachine("arduino.cc")
	if arduinoMachine == nil || arduinoMachine.Name != "arduino.cc" {
		logrus.WithError(err).Error("Credentials not found")
		return "", "", errors.New("credentials not found, try typing `arduino login` to login")
	}

	logrus.Info("Refreshing user session")
	newToken, err := authConf.Refresh(arduinoMachine.Password)
	if err != nil {
		logrus.WithError(err).Error("Session expired, try typing `arduino login` to login again")
		return "", "", err
	}

	var token string
	if newToken.TTL != 0 { // We haven't recently requested a valid token, which is in .netrc under "account", so we have to update it.
		arduinoMachine.UpdatePassword(newToken.Refresh)
		arduinoMachine.UpdateAccount(newToken.Access)
		token = newToken.Access
	} else {
		token = arduinoMachine.Account
	}

	content, err := NetRC.MarshalText()
	if err == nil { //serialize new info
		err = ioutil.WriteFile(netRCFile, content, 0666)
		if err != nil {
			logrus.WithError(err).Error("Cannot write new ~/.netrc file")
		}
	} else {
		logrus.WithError(err).Error("Cannot serialize ~/.netrc file")
	}
	logrus.Info("Login successful")
	return arduinoMachine.Login, token, nil
}
