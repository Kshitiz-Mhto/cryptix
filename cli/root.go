/*
Copyright © 2025 Kshitiz Mhto <kshitizmhto101@gmail.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cli

import (
	"fmt"
	"os"

	"github.com/Kshitiz-Mhto/cryptix/cli/logger"
	"github.com/Kshitiz-Mhto/cryptix/cli/subcmd"
	"github.com/Kshitiz-Mhto/cryptix/cli/subcmd/keys"
	"github.com/Kshitiz-Mhto/cryptix/cli/subcmd/mail"
	"github.com/spf13/cobra"
)

var version bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cryptix",
	Short: "Simple CLI tool that encrypt and decrypt messages/text for sharing.",
	Long: `Cryptix is a command-line utility designed to encrypt a given message or text using the AES algorithm. 
The AES key is itself encrypted with an RSA public key, ensuring that the encrypted message can only be decrypted 
using the corresponding RSA private key. Additionally, the tool offers an option for sharing the encrypted data securely, 
allowing recipients with the necessary private key to decrypt and access the original message.`,
	Run: func(cmd *cobra.Command, args []string) {
		if version {
			versionCMD.Run(cmd, args)
		} else {
			fmt.Print(logo)
			cmd.Help()
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	logger.InitLogger()
	rootCmd.AddCommand(versionCMD)
	rootCmd.AddCommand(subcmd.EmbadeCmd)
	rootCmd.AddCommand(subcmd.DecodeCmd)
	rootCmd.AddCommand(mail.SendMailCmd)
	rootCmd.AddCommand(keys.GenerateKeyCmd)

	rootCmd.Flags().BoolP("version", "v", false, "Version of CLI")
}
