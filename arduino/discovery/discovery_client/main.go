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

// discovery_client is a command line UI client to test pluggable discoveries.
package main

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/arduino/arduino-cli/arduino/discovery"
	"github.com/arduino/arduino-cli/arduino/discovery/discoverymanager"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.ErrorLevel)
	dm := discoverymanager.New()
	for _, discCmd := range os.Args[1:] {
		disc := discovery.New(discCmd, discCmd)
		dm.Add(disc)
	}
	dm.Start()

	activePorts := map[string]*discovery.Port{}
	watcher, err := dm.Watch()
	if err != nil {
		log.Fatalf("failed to start discoveries: %v", err)
	}
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	l := widgets.NewList()
	l.Title = "List"
	l.TextStyle = ui.NewStyle(ui.ColorYellow)
	l.WrapText = false
	w, h := ui.TerminalDimensions()
	l.SetRect(0, 0, w, h)

	updateList := func() {
		rows := []string{}
		rows = append(rows, "Available ports list:")

		ids := sort.StringSlice{}
		for id := range activePorts {
			ids = append(ids, id)
		}
		ids.Sort()
		for _, id := range ids {
			port := activePorts[id]
			rows = append(rows, fmt.Sprintf("> Address: %s", port.AddressLabel))
			rows = append(rows, fmt.Sprintf("  Protocol: %s", port.ProtocolLabel))
			keys := port.Properties.Keys()
			sort.Strings(keys)
			for _, k := range keys {
				rows = append(rows, fmt.Sprintf("                  %s=%s", k, port.Properties.Get(k)))
			}
		}
		l.Rows = rows
	}
	updateList()
	ui.Render(l)

	previousKey := ""
	uiEvents := ui.PollEvents()
out:
	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				l.SetRect(0, 0, payload.Width, payload.Height)
				ui.Clear()
			case "q", "<C-c>":
				break out
			case "j", "<Down>":
				l.ScrollDown()
			case "k", "<Up>":
				l.ScrollUp()
			case "<C-d>":
				l.ScrollHalfPageDown()
			case "<C-u>":
				l.ScrollHalfPageUp()
			case "<C-f>":
				l.ScrollPageDown()
			case "<C-b>":
				l.ScrollPageUp()
			case "g":
				if previousKey == "g" {
					l.ScrollTop()
				}
			case "<Home>":
				l.ScrollTop()
			case "G", "<End>":
				l.ScrollBottom()
			}

			if previousKey == "g" {
				previousKey = ""
			} else {
				previousKey = e.ID
			}

		case ev := <-watcher.Feed():
			if ev.Type == "add" {
				activePorts[ev.Port.Address+"|"+ev.Port.Protocol] = ev.Port
			}
			if ev.Type == "remove" {
				delete(activePorts, ev.Port.Address+"|"+ev.Port.Protocol)
			}
			updateList()
		}

		ui.Render(l)
	}
}
