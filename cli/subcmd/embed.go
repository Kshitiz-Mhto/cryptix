/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package subcmd

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Kshitiz-Mhto/stegomail/cli/logger"
	"github.com/Kshitiz-Mhto/stegomail/utility"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	imageFile          string
	msg                string
	encryptedImagePath string
	pubkeyPath         string
)

// EmbadeCmd represents the encode command
var EmbadeCmd = &cobra.Command{
	Use:     "embed",
	Aliases: []string{"encrypt", "encode", "eb"},
	Short:   "It helps to endcode the message inside image files using DCT",
	Example: "stegomail encode --image <path/to/image> --message <message_content> --output <path/to/> --pubkey <path/to/public_key>",
	Run:     runEncodingSecretsCmd,
}

func runEncodingSecretsCmd(cmd *cobra.Command, args []string) {
	imageFile, _ = cmd.Flags().GetString("image")
	msg, _ = cmd.Flags().GetString("message")
	pubkeyPath, _ = cmd.Flags().GetString("pubkey")

	if msg == "" {
		utility.Error("Message to be encrypted  is empty.")
		logger.Logger.Fatal("Message to be encrypted is empty")
	}

	absPathOfImage, err := filepath.Abs(imageFile)
	if err != nil {
		utility.Error("failed to get absolute path: %v", err)
		logger.Logger.WithFields(logrus.Fields{
			"path": absPathOfImage,
			"err":  err,
		}).Fatal("failed to get absolute path")
	}

	absPathOfPubkey, err := filepath.Abs(pubkeyPath)
	if err != nil {
		utility.Error("failed to get absolute path: %v", err)
		logger.Logger.WithFields(logrus.Fields{
			"path": absPathOfPubkey,
			"err":  err,
		}).Fatal("failed to get absolute path")
	}

	pubKey, err := LoadPublicKey(pubkeyPath)

	if err != nil {
		utility.Info("Aborting operation process: %s", utility.Red("PubKey file loading"))
		os.Exit(1)
	}

	encryptedMsg, encryptedAESKey, err := HybridEncryption([]byte(msg), pubKey)
	if err != nil {
		utility.Info("Aborting operation process : %s", utility.Red("Msg Encryption"))
		os.Exit(1)
	}

	utility.Info("%s ------ %s", encryptedMsg, encryptedAESKey)

}

func init() {
	EmbadeCmd.Flags().StringVarP(&imageFile, "image", "i", "", "Specify the path of image that will be embaded with message. [*Required]")
	EmbadeCmd.Flags().StringVarP(&msg, "message", "m", "", "Specify your message that will be encoded and embedded. [*Required]")
	EmbadeCmd.Flags().StringVarP(&encryptedImagePath, "output", "o", ".", "Specify the directory where embeded image will reside. [Default path: current directory]")
	EmbadeCmd.Flags().StringVarP(&pubkeyPath, "pubkey", "k", "", "Specify your public key file path. [*Required]")

	EmbadeCmd.MarkFlagsRequiredTogether("image", "message", "pubkey")
}

func HybridEncryption(message []byte, pubKey *rsa.PublicKey) ([]byte, []byte, error) {
	// Generate random AES-256 key
	aesKey := make([]byte, 32)
	if _, err := rand.Read(aesKey); err != nil {
		utility.Error("failed to generate AES key: %v", err)
		logger.Logger.WithFields(logrus.Fields{"err": err}).Error("Failed to generate AES key")
		return nil, nil, err
	}

	// Create AES cipher
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		utility.Error("failed to create AES cipher: %v", err)
		logger.Logger.WithFields(logrus.Fields{"err": err}).Error("Failed to create AES cipher")
		return nil, nil, err
	}

	// Create AES-GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		utility.Error("failed to create AES-GCM mode: %v", err)
		logger.Logger.WithFields(logrus.Fields{"err": err}).Error("Failed to create AES-GCM mode")
		return nil, nil, err
	}

	// Generate a nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		utility.Error("failed to generate nonce: %v", err)
		logger.Logger.WithFields(logrus.Fields{"err": err}).Error("Failed to generate nonce")
		return nil, nil, err
	}

	// Encrypt message using AES-GCM
	encryptedMsg := gcm.Seal(nonce, nonce, message, nil)

	// Encrypt AES key with RSA-OAEP
	encryptedAESKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, aesKey, nil)
	if err != nil {
		utility.Error("failed to encrypt AES key with RSA: %v", err)
		logger.Logger.WithFields(logrus.Fields{"err": err}).Error("Failed to encrypt AES key with RSA")
		return nil, nil, err
	}

	utility.Success("Message is encrypted successfully!")
	logger.Logger.Info("Message is encrypted successfully!")
	return encryptedMsg, encryptedAESKey, nil
}

func LoadPublicKey(path string) (*rsa.PublicKey, error) {

	absPathOfPubkey, err := filepath.Abs(path)
	if err != nil {
		utility.Error("failed to get absolute path: %v", err)
		logger.Logger.WithFields(logrus.Fields{
			"path": absPathOfPubkey,
			"err":  err,
		}).Error("failed to get absolute path")
		return nil, err
	}

	// Read the file
	pubKeyBytes, err := os.ReadFile(absPathOfPubkey)
	if err != nil {
		utility.Error("failed to read public key file: %v", err)
		logger.Logger.WithFields(logrus.Fields{
			"err": err,
		}).Error("failed to read public key file")
		return nil, err
	}

	// Decode PEM block
	block, _ := pem.Decode(pubKeyBytes)
	if block == nil || block.Type != "RSA PUBLIC KEY" {
		utility.Error("invalid public key format")
		logger.Logger.Error("invalid public key format")
		return nil, fmt.Errorf("invalid public key format")
	}

	// Parse the public key
	pubKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		utility.Error("failed to parse public key: %v", err)
		logger.Logger.WithFields(logrus.Fields{
			"err": err,
		}).Error("failed to parse public key")
		return nil, err
	}

	utility.Success("PubKey file is loaded succesfully!!")
	logger.Logger.Info("PubKey file is loaded succesfully!!")

	return pubKey, nil
}
