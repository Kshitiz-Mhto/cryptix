/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package subcmd

import (
	"os"
	"path/filepath"

	"github.com/Kshitiz-Mhto/cryptix/cli/logger"
	"github.com/Kshitiz-Mhto/cryptix/crypt"
	"github.com/Kshitiz-Mhto/cryptix/pkg/env"
	"github.com/Kshitiz-Mhto/cryptix/utility"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	privateKeyFilePath   string
	sourcePath           string
	outputMsgFileName    string
	outputPath           string
	DecryptedMsgFilePath string
)

// DecodeCmd represents the decode command.
var DecodeCmd = &cobra.Command{
	Use:     "decode",
	Aliases: []string{"decrypt", "de"},
	Short:   "Decrypt the encoded message from json file.",
	Example: "cryptix decode --source <path/to/source_file> --name <file_name> --output <path/to/storge_dir> --prikey <path/to/private_key>",
	Run:     runDecodeSecretsCmd,
}

func runDecodeSecretsCmd(cmd *cobra.Command, args []string) {
	privateKeyFilePath, _ = cmd.Flags().GetString("prikey")
	sourcePath, _ = cmd.Flags().GetString("source")
	outputMsgFileName, _ = cmd.Flags().GetString("name")
	outputPath, _ = cmd.Flags().GetString("output")

	// Load private key.
	privKey, err := crypt.LoadPrivateKey(privateKeyFilePath)
	if err != nil {
		utility.Error("%s", err)
		utility.Info("Aborting operation: %s", utility.Red("Private key file loading"))
		os.Exit(1)
	}

	// Decrypt the hybrid-encrypted data.
	plaintext, err := crypt.HybridDecryption(sourcePath, privKey)
	if err != nil {
		utility.Info("Aborting operation: %s", utility.Red("Decryption failed"))
		os.Exit(1)
	}

	absOutputPath, err := filepath.Abs(outputPath)
	if err != nil {
		utility.Info("Aborting operation: %s", utility.Red("Absolute path retrieval"))
		logger.Logger.WithFields(logrus.Fields{"err": err}).Error("Absolute path retrieval")
		os.Exit(1)
	}

	if err := os.MkdirAll(absOutputPath, os.ModePerm); err != nil {
		utility.Info("Aborting operation: %s", utility.Red("Directory creation"))
		logger.Logger.WithFields(logrus.Fields{"err": err}).Error("Directory creation")
		os.Exit(1)
	}

	DecryptedMsgFilePath = filepath.Join(absOutputPath, outputMsgFileName+env.Vars.TXT_FORMAT)
	if err = os.WriteFile(DecryptedMsgFilePath, plaintext, 0644); err != nil {
		utility.Error("Failed to write output message: %v", err)
		logger.Logger.WithFields(logrus.Fields{"err": err}).Error("Write failure")
		os.Exit(1)
	}

}

func init() {
	DecodeCmd.Flags().StringVarP(&privateKeyFilePath, "prikey", "k", "", "Specify private key file path. [*Required]")
	DecodeCmd.Flags().StringVarP(&sourcePath, "source", "s", "", "Specify the source path file path containing encrypted data. [*Required]")
	DecodeCmd.Flags().StringVarP(&outputMsgFileName, "name", "n", "", "Specify the filename for storing decrypted message with not extension. [*Required]")
	DecodeCmd.Flags().StringVarP(&outputPath, "output", "o", ".", "Specify the path where you want to store file that stored decrypted message. Optional[]")

	DecodeCmd.MarkFlagsRequiredTogether("prikey", "source", "name")
}
