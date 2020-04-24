package main

import (
	"fmt"
	"os"

	"github.com/arduino/arduino-cli/i18n/cmd/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
