/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package mail

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"time"

	"github.com/Kshitiz-Mhto/cryptix/cli/logger"
	"github.com/Kshitiz-Mhto/cryptix/utility"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	sourcePath string
	mail       string
	subject    string
)

// SendMailCmd represents the send command
var SendMailCmd = &cobra.Command{
	Use:     "send",
	Short:   "It basically help to send mail with attachment",
	Example: "stegomail send --source <path/to/file> --mail <email_address> --subject <mail_subject>",
	Run:     runSendMailCmd,
}

func runSendMailCmd(cmd *cobra.Command, args []string) {
	absSourceFilePath, err := filepath.Abs(sourcePath)
	if err != nil {
		utility.Info("Aborting operation: %s", utility.Red("Absolute path retrieval"))
		logger.Logger.WithFields(logrus.Fields{"err": err}).Error("Absolute path retrieval")
		os.Exit(1)
	}

	/*
		Upload file to Google Drive
		downloadLink, err := UploadFileToGoogleDrive(absSourceFilePath)
		if err != nil {
			utility.Error("Failed to upload file to Google Drive: %s", err)
			logger.Logger.WithFields(logrus.Fields{"err": err}).Error("Failed to upload file to Google Drive")
			os.Exit(1)
		}
	*/

	//test
	downloadLink := "https://drive.google.com/file/d/1BSOF_hPOZqycyKYuzBNOBPe83KFKX76M/view?usp=sharing"

	fileData, err := os.ReadFile(absSourceFilePath)
	if err != nil {
		utility.Error("Failed to read encrypted file: %s", err)
		logger.Logger.WithFields(logrus.Fields{
			"file": absSourceFilePath,
			"err":  err,
		}).Error("Failed to read encrypted file")
		os.Exit(1)
	}
	encodedFileData := base64.StdEncoding.EncodeToString(fileData)
	fileName := filepath.Base(absSourceFilePath)

	vars := map[string]interface{}{
		"filename":     fileName,
		"filedata":     encodedFileData,
		"downloadlink": downloadLink,
		"time":         time.Now().Format(time.RFC1123),
	}

	success := HTMLTemplateMailHandler(mail, subject, vars)
	if !success {
		utility.Error("Failed to send mail")
		logger.Logger.Error("Failed to send mail")
		os.Exit(1)
	}

	utility.Success("Email sent successfully!!")
	logger.Logger.Info("Email sent successfully!!")
}

func init() {
	SendMailCmd.Flags().StringVarP(&sourcePath, "source", "s", "", "Specify the source path of file.[*Required]")
	SendMailCmd.Flags().StringVarP(&mail, "mail", "m", "", "Specify mail address. [*Required]")
	SendMailCmd.Flags().StringVarP(&subject, "subject", "S", "", "Specify your mail subject. [Optional]")

	SendMailCmd.MarkFlagsRequiredTogether("source", "mail")
}
