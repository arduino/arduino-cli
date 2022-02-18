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
	"time"

	"github.com/arduino/arduino-cli/arduino/discovery"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

func main() {
	discoveries := []*discovery.PluggableDiscovery{}
	discEvent := make(chan *discovery.Event)
	for _, discCmd := range os.Args[1:] {
		disc, err := discovery.New("", discCmd)
		if err != nil {
			log.Fatal("Error initializing discovery:", err)
		}

		if err := disc.Run(); err != nil {
			log.Fatal("Error starting discovery:", err)
		}
		if err := disc.Start(); err != nil {
			log.Fatal("Error starting discovery:", err)
		}
		eventChan, err := disc.StartSync(10)
		if err != nil {
			log.Fatal("Error starting discovery:", err)
		}
		go func() {
			for msg := range eventChan {
				discEvent <- msg
			}
		}()
		discoveries = append(discoveries, disc)
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
		for _, disc := range discoveries {
			for i, port := range disc.ListCachedPorts() {
				rows = append(rows, fmt.Sprintf(" [%04d] Address: %s", i, port.AddressLabel))
				rows = append(rows, fmt.Sprintf("        Protocol: %s", port.ProtocolLabel))
				keys := port.Properties.Keys()
				sort.Strings(keys)
				for _, k := range keys {
					rows = append(rows, fmt.Sprintf("                  %s=%s", k, port.Properties.Get(k)))
				}
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

		case <-discEvent:
			updateList()
		}

		ui.Render(l)
	}

	for _, disc := range discoveries {
		disc.Quit()
		fmt.Println("Discovery QUITed")
		for disc.State() == discovery.Alive {
			time.Sleep(time.Millisecond)
		}
		fmt.Println("Discovery correctly terminated")
	}

}
