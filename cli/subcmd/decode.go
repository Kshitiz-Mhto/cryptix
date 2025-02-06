/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package subcmd

import (
	"github.com/spf13/cobra"
)

// decodeCmd represents the decode command
var DecodeCmd = &cobra.Command{
	Use:     "decode",
	Aliases: []string{"decrypt", "de"},
	Short:   "It helps to decrypt the encoded message inside the image file.",
	Example: "stegomail decode --image <path/to/image_file>",
	Run:     runDecodeSecretsCmd,
}

func runDecodeSecretsCmd(cmd *cobra.Command, args []string) {

}
