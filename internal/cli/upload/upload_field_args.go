// This file is part of arduino-cli.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package upload

import (
	"fmt"

	"github.com/arduino/arduino-cli/internal/cli/arguments"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/spf13/cobra"
)

type argumentsUploadFields map[string]string

func (a *argumentsUploadFields) AddToCommand(cmd *cobra.Command) {
	arguments.AddKeyValuePFlag(cmd, (*map[string]string)(a), "upload-field", "F", nil, i18n.Tr("Set a value for a field required to upload."))
}

func (a *argumentsUploadFields) Get() map[string]string {
	return map[string]string(*a)
}

func (a *argumentsUploadFields) ResolveUserFields(userFields []*rpc.UserField) map[string]string {
	fields := map[string]string{}
	if len(userFields) > 0 {
		if len(*a) > 0 {
			// If the user has specified some fields via cmd-line, we don't ask for them
			for _, field := range userFields {
				if value, ok := (*a)[field.GetName()]; ok {
					fields[field.GetName()] = value
				} else {
					feedback.Fatal(i18n.Tr("Missing required upload field: %s", field.GetName()), feedback.ErrBadArgument)
				}
			}
		} else {
			// Otherwise prompt the user for them
			feedback.Print(i18n.Tr("Uploading to the specified board requires the following information:"))
			if f, err := arguments.AskForUserFields(userFields); err != nil {
				msg := fmt.Sprintf("%s: %s", i18n.Tr("Error getting user input"), err)
				feedback.Fatal(msg, feedback.ErrGeneric)
			} else {
				fields = f
			}
		}
	}
	return fields
}
