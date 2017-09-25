/*
 * This file is part of arduino-cli.
 *
 * arduino-cli is free software; you can redistribute it and/or modify
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
 * Copyright 2017 ARDUINO AG (http://www.arduino.cc/)
 */

package cmd

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/bcmi-labs/arduino-cli/auth"
	"github.com/bcmi-labs/arduino-cli/create_client_helpers"
	"github.com/bgentry/go-netrc/netrc"
	homedir "github.com/mitchellh/go-homedir"

	"github.com/bcmi-labs/arduino-modules/sketches"

	"github.com/bcmi-labs/arduino-cli/cmd/formatter"
	"github.com/bcmi-labs/arduino-cli/common"

	"github.com/spf13/cobra"
)

var arduinoSketchCmd = &cobra.Command{
	Use:     "sketch",
	Short:   `Arduino CLI Sketch Commands`,
	Long:    `Arduino CLI Sketch Commands`,
	Example: `arduino sketch sync`,
	RunE:    executeSketchCommand,
}

var arduinoSketchSyncCmd = &cobra.Command{
	Use:     "sync",
	Short:   `Arduino CLI Sketch Commands`,
	Long:    `Arduino CLI Sketch Commands`,
	Example: `arduino sketch sync`,
	RunE:    executeSketchSyncCommand,
}

func executeSketchCommand(cmd *cobra.Command, args []string) error {
	return errors.New("No subcommand specified")
}

func executeSketchSyncCommand(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return errors.New("No arguments are accepted")
	}

	sketchbook, err := common.GetDefaultArduinoHomeFolder()
	if err != nil {
		formatter.PrintErrorMessage("Cannot get sketchbook folder")
		return nil
	}

	username, bearerToken, err := login()
	if err != nil {
		if GlobalFlags.Verbose == 0 {
			formatter.PrintErrorMessage("Cannot login automatically: try arduino login the run again this command")
		} else {
			formatter.PrintError(err)
		}
	}

	sketchMap := sketches.Find(sketchbook, "libraries") //exclude libraries folder

	client := createClient.New(nil)
	tok := "Bearer " + bearerToken
	resp, err := client.SearchSketches(context.Background(), createClient.SearchSketchesPath(), nil, &username, &tok)
	if err != nil {
		formatter.PrintErrorMessage("Cannot get create sketches, sync failed")
		return nil
	}
	defer resp.Body.Close()

	onlineSketches, err := client.DecodeArduinoCreateSketches(resp)
	if err != nil {
		formatter.PrintErrorMessage("Cannot unmarshal response from create, sync failed")
		return nil
	}

	onlineSketchesMap := make(map[string]*createClient.ArduinoCreateSketch, len(onlineSketches.Sketches))
	for _, item := range onlineSketches.Sketches {
		onlineSketchesMap[*item.Name] = item
	}

	for _, item := range sketchMap {

		itemOnline, hasConflict := onlineSketchesMap[item.Name]
		if hasConflict {
			//solve conflicts
			priority := arduinoSketchSyncFlags.Priority
			if priority == "ask" {
				if !formatter.IsCurrentFormat("text") {
					formatter.PrintErrorMessage("ask mode for this command is only supported using text format")
					return nil
				}
				firstAsk := true
				for priority != "remote" &&
					priority != "local" &&
					priority != "skip-conflict" {
					if !firstAsk {
						formatter.Print("Invalid option: " + priority)
					}
					formatter.Print("What should I retain if I have a conflict between local and remote sketches? [remote | local | skip-conflict]")
					fmt.Scanln(&priority)
					firstAsk = false
				}
			}
			switch priority {
			case "remote":
				formatter.Print("pushing edits of sketch: " + item.Name)
				err := editSketch(*item, sketchbook, bearerToken)
				if err != nil {
					formatter.PrintError(err)
				}
				break
			case "local":
				formatter.Print("pulling " + item.Name)
				err := pullSketch(itemOnline, sketchbook, bearerToken)
				if err != nil {
					formatter.PrintError(err)
				}
				break
			case "skip-conflict":
				formatter.Print("skipping " + item.Name)
				break
			default:
				priority = "skip-conflict"
				if GlobalFlags.Verbose > 0 {
					formatter.Print("Priority not recognized, using skip-conflict")
				}
				formatter.Print("skipping " + item.Name)
			}

		} else { //only local, push
			formatter.Print("pushing " + item.Name)
			pushSketch(*item, sketchbook, bearerToken)
		}
	}
	for _, item := range onlineSketches.Sketches {
		_, hasConflict := onlineSketchesMap[*item.Name]
		if hasConflict {
			continue
		}
		//only online, pull
		formatter.Print("pulling " + *item.Name)
		err := pullSketch(item, sketchbook, bearerToken)
		if err != nil {
			formatter.PrintError(err)
		}
	}
	formatter.PrintResult("OK") // Issue # : Provide output struct to print the result in a prettier way.
	return nil
}

func pushSketch(sketch sketches.Sketch, sketchbook string, bearerToken string) error {
	client := createClient.New(nil)

	resp, err := client.CreateSketches(context.Background(), createClient.CreateSketchesPath(), createClient.ConvertFrom(sketch), "Bearer "+bearerToken)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = client.DecodeArduinoCreateSketch(resp)
	if err != nil {
		errorMsg, err := client.DecodeErrorResponse(resp)
		if err != nil {
			return err
		}
		return errorMsg
	}

	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}
	return nil
}

func editSketch(sketch sketches.Sketch, sketchbook string, bearerToken string) error {
	client := createClient.New(nil)

	resp, err := client.EditSketches(context.Background(), createClient.EditSketchesPath(sketch.ID), createClient.ConvertFrom(sketch), "Bearer "+bearerToken)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = client.DecodeArduinoCreateSketch(resp)
	if err != nil {
		errorMsg, err := client.DecodeErrorResponse(resp)
		if err != nil {
			return err
		}
		return errorMsg
	}

	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}
	return nil
}

func pullSketch(sketch *createClient.ArduinoCreateSketch, sketchbook string, bearerToken string) error {
	client := createClient.New(nil)
	bearer := "Bearer " + bearerToken

	resp, err := client.ShowSketches(context.Background(), createClient.ShowSketchesPath(fmt.Sprint(sketch.ID)), &bearer)
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

	sketchFolder, err := ioutil.TempDir(sketchbook, fmt.Sprintf("%s-temp", *sketch.Name))
	if err != nil {
		return err
	}
	defer os.RemoveAll(sketchFolder)

	destFolder := filepath.Join(sketchbook, *sketch.Name)

	for _, file := range append(r.Files, sketch.Ino) {
		path := findPathOf(*sketch.Name, *file.Path)

		err := os.MkdirAll(filepath.Dir(path), 0755)
		if err != nil {
			return err
		}

		resp, err := client.ShowFiles(context.Background(), createClient.ShowFilesPath("sketch", sketch.ID.String(), path))
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

		path = filepath.Join(sketchFolder, path)
		decodedData, err := base64.StdEncoding.DecodeString(*filewithData.Data)
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(path, decodedData, 0666)
		if err != nil {
			return errors.New("Copy of a file of the downloaded sketch failed, sync failed")
		}
	}

	err = os.RemoveAll(destFolder)
	if err != nil {
		return err
	}

	err = os.Rename(sketchFolder, destFolder)
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
			return filepath.Join(list[i+1 : len(list)]...)
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

	netRCFile := filepath.Join(home, ".netrc")
	file, err := os.OpenFile(netRCFile, os.O_CREATE|os.O_RDONLY, 0666)
	if err != nil {
		return "", "", err
	}
	NetRC, err := netrc.Parse(file)
	if err != nil {
		return "", "", err
	}

	arduinoMachine := NetRC.FindMachine("arduino.cc")
	if arduinoMachine.Name != "arduino.cc" {
		return "", "", errors.New("Credentials not present, try login with arduino login first")
	}

	newToken, err := authConf.Refresh(arduinoMachine.Password)
	if err != nil {
		return "", "", err
	}

	var token string
	if newToken.TTL != 0 { //we haven't recently requested a valid token, which is in .netrc under "account", so we have to update it
		arduinoMachine.UpdatePassword(newToken.Refresh)
		arduinoMachine.UpdateAccount(newToken.Access)
		token = newToken.Access
	} else {
		token = arduinoMachine.Account
	}

	content, err := NetRC.MarshalText()
	if err == nil { //serialize new info
		err := ioutil.WriteFile(netRCFile, content, 0666)
		if err != nil && GlobalFlags.Verbose > 0 {
			formatter.Print(err.Error())
		}
	} else if GlobalFlags.Verbose > 0 {
		formatter.Print(err.Error())
	}
	return arduinoMachine.Login, token, nil
}
