/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package subcmd

import (
	"github.com/spf13/cobra"
)

// sendCmd represents the send command
var SendCmd = &cobra.Command{
	Use:     "send",
	Short:   "It basically help to send mail with attachment",
	Example: "stegomail send --image <path/to/image_file>",
	Run:     runSendMailCmd,
}

func runSendMailCmd(cmd *cobra.Command, args []string) {

}
