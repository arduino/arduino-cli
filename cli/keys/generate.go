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

package keys

import (
	"context"
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/commands/keys"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	keyName       []string // Name of the custom keys to generate. Can be used multiple times for multiple security keys.
	algorithmType string   // Algorithm type to use
	keysKeychain  string   // Path of the dir where to save the custom keys
)

func initGenerateCommand() *cobra.Command {
	generateCommand := &cobra.Command{
		Use:   "generate",
		Short: tr("Generate the security keys."),
		Long:  tr("Generate the security keys required for secure boot"),
		Example: "" +
			"  " + os.Args[0] + " keys generate -t ecdsa-p256 --key-name ecdsa-p256-signing-key.pem --key-name ecdsa-p256-encrypt-key.pem --keys-keychain /home/user/Arduino/MyKeys\n" +
			"  " + os.Args[0] + " keys generate --key-name ecdsa-p256-signing-key.pem",
		Args: cobra.NoArgs,
		Run:  runGenerateCommand,
	}

	generateCommand.Flags().StringVarP(&algorithmType, "type", "t", "ecdsa-p256", tr("Algorithm type to use"))
	generateCommand.Flags().StringArrayVar(&keyName, "key-name", []string{}, tr("Name of the custom key to generate. Can be used multiple times for multiple security keys."))
	generateCommand.MarkFlagRequired("key-name")
	generateCommand.RegisterFlagCompletionFunc("key-name", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		defaultKeyNames := []string{"ecdsa-p256-signing-key.pem", "ecdsa-p256-encrypt-key.pem"}
		return defaultKeyNames, cobra.ShellCompDirectiveDefault
	})
	generateCommand.Flags().StringVar(&keysKeychain, "keys-keychain", "", tr("The path of the dir where to save the custom keys"))

	return generateCommand
}

func runGenerateCommand(command *cobra.Command, args []string) {

	logrus.Info("Executing `arduino-cli keys generate`")

	resp, err := keys.Generate(context.Background(), &rpc.KeysGenerateRequest{
		AlgorithmType: algorithmType,
		KeyName:       keyName,
		KeysKeychain:  keysKeychain,
	})
	if err != nil {
		feedback.Errorf(tr("Error during generate: %v"), err)
		os.Exit(errorcodes.ErrGeneric)
	}
	feedback.PrintResult(result{resp})
}

// output from this command requires special formatting, let's create a dedicated
// feedback.Result implementation
type result struct {
	Keychain *rpc.KeysGenerateResponse `json:"keys_keychain"`
}

func (dr result) Data() interface{} {
	return dr.Keychain
}

func (dr result) String() string {
	return (tr("Keys created in: %s", dr.Keychain.KeysKeychain))
}
