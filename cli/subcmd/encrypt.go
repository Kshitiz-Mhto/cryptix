/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package subcmd

import (
	"os"
	"path/filepath"

	"github.com/Kshitiz-Mhto/cryptix/cli/logger"
	"github.com/Kshitiz-Mhto/cryptix/crypt"
	"github.com/Kshitiz-Mhto/cryptix/utility"
	"github.com/spf13/cobra"
)

var (
	msg            string
	pubkeyPath     string
	outputFileName string
	outputFilePath string
)

// EmbadeCmd represents the encode command
var EmbadeCmd = &cobra.Command{
	Use:     "encrypt",
	Aliases: []string{"encode", "en"},
	Short:   "It helps to endcode the message inside image files using DCT",
	Example: "cryptix encode --message <message_content> --output <path/to/> --name <filename> --pubkey <path/to/public_key>",
	Run:     runEncodingSecretsCmd,
}

func runEncodingSecretsCmd(cmd *cobra.Command, args []string) {
	msg, _ = cmd.Flags().GetString("message")
	pubkeyPath, _ = cmd.Flags().GetString("pubkey")
	outputFilePath, _ = cmd.Flags().GetString("output")
	outputFileName, _ = cmd.Flags().GetString("name")

	if msg == "" {
		utility.Error("Message to be encrypted  is empty.")
		logger.Logger.Fatal("Message to be encrypted is empty")
	}

	pubKey, err := crypt.LoadPublicKey(pubkeyPath)
	if err != nil {
		utility.Info("Aborting operation process: %s", utility.Red("PubKey file loading"))
		os.Exit(1)
	}

	encryptedMsg, encryptedAESKey, err := crypt.HybridEncryption([]byte(msg), pubKey)
	if err != nil {
		utility.Info("Aborting operation process : %s", utility.Red("Encryption generation"))
		os.Exit(1)
	}

	err = crypt.EncryptHybridData(encryptedMsg, encryptedAESKey, outputFilePath, outputFileName)
	if err != nil {
		utility.Info("Aborting operation process : %s", utility.Red("Message encryption"))
		os.Exit(1)
	}

	utility.Success("Encryption successful! Encrypted data saved at: %s", filepath.Join(outputFilePath, outputFileName))
}

func init() {
	EmbadeCmd.Flags().StringVarP(&msg, "message", "m", "", "Specify your message that will be encoded. [*Required]")
	EmbadeCmd.Flags().StringVarP(&outputFilePath, "output", "o", ".", "Specify the directory where file will be located. [Default path: current directory]")
	EmbadeCmd.Flags().StringVarP(&pubkeyPath, "pubkey", "k", "", "Specify your public key file path. [*Required]")
	EmbadeCmd.Flags().StringVarP(&outputFileName, "name", "n", "", "Specify your output file name(dont include extension). [*Required]")

	EmbadeCmd.MarkFlagsRequiredTogether("message", "pubkey", "name")
}
