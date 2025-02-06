/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package subcmd

import (
	"github.com/Kshitiz-Mhto/stegomail/cli/logger"
	"github.com/spf13/cobra"
)

// encodeCmd represents the encode command
var EncodeCmd = &cobra.Command{
	Use:     "encode",
	Aliases: []string{"encrypt", "en"},
	Short:   "It helps to endcode the message inside image files using DCT",
	Example: "stegomail encode --image <path/to/image> --message <message_content> --output <path/to/encrypted_image_file>",
	Run:     runEncodingSecretsCmd,
}

func runEncodingSecretsCmd(cmd *cobra.Command, args []string) {
	logger.Logger.Info("run encdoing")
}
