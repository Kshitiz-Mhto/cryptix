package keys

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"

	"github.com/Kshitiz-Mhto/cryptix/cli/logger"
	"github.com/Kshitiz-Mhto/cryptix/utility"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var path string

var GenerateKeyCmd = &cobra.Command{
	Use:     "gen",
	Aliases: []string{"generate-keys", "gen-key"},
	Short:   "Generates RSA key pair for encryption and decryption.",
	Run:     runKeyGenerationCmd,
}

func runKeyGenerationCmd(cmd *cobra.Command, args []string) {
	path, _ = cmd.Flags().GetString("path")
	GenerateRSAKeys(path)
}

func GenerateRSAKeys(path string) {

	absolutePath, err := filepath.Abs(path)
	if err != nil {
		utility.Error("failed to get absolute path: %v", err)
		logger.Logger.WithFields(logrus.Fields{
			"path": absolutePath,
			"err":  err,
		}).Fatal("failed to get absolute path")
	}

	// Ensure the directory exists
	if err := os.MkdirAll(absolutePath, 0700); err != nil {
		utility.Error("failed to create directory %s: %v", path, err)
		logger.Logger.WithFields(logrus.Fields{
			"path": absolutePath,
			"err":  err,
		}).Fatal("failed to create directory")
	}

	// Generate 2048-bit RSA key pair
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		utility.Error("failed to generate RSA key: %v", err)
		logger.Logger.WithFields(logrus.Fields{
			"err": err,
		}).Fatal("failed to generate RSA key")
	}
	pubKey := &privKey.PublicKey

	// Save private key
	privKeyBytes := x509.MarshalPKCS1PrivateKey(privKey)
	privKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privKeyBytes,
	}
	privPath := filepath.Join(absolutePath, "private.pem")
	privFile, err := os.Create(privPath)
	if err != nil {
		utility.Error("failed to create private key file: %v", err)
		logger.Logger.WithFields(logrus.Fields{
			"err": err,
		}).Fatal("failed to create private key file")
	}
	defer privFile.Close()

	if err := pem.Encode(privFile, privKeyPEM); err != nil {
		utility.Error("failed to write private key: %v", err)
		logger.Logger.WithFields(logrus.Fields{
			"err": err,
		}).Fatal("failed to write private key")
	}

	// Save public key
	pubKeyBytes := x509.MarshalPKCS1PublicKey(pubKey)
	pubKeyPEM := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubKeyBytes,
	}
	pubPath := filepath.Join(absolutePath, "public.pem")
	pubFile, err := os.Create(pubPath)
	if err != nil {
		utility.Error("failed to create public key file: %v", err)
		logger.Logger.WithFields(logrus.Fields{
			"err": err,
		}).Fatal("failed to create public key file")
	}
	defer pubFile.Close()

	if err := pem.Encode(pubFile, pubKeyPEM); err != nil {
		utility.Error("failed to write public key: %v", err)
		logger.Logger.WithFields(logrus.Fields{
			"err": err,
		}).Fatal("failed to write public key")
	}

	utility.Success("RSA key pair generated successfully! at path: %s", absolutePath)
	logger.Logger.WithFields(logrus.Fields{
		"path": absolutePath,
	}).Info("RSA key pair generated successfully!")
}

func init() {
	GenerateKeyCmd.Flags().StringVarP(&path, "path", "o", ".", "Path where RSA keys-pairs will be created. [Default path: current directory]")
}
