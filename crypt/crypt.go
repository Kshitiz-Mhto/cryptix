package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Kshitiz-Mhto/cryptix/cli/logger"
	"github.com/Kshitiz-Mhto/cryptix/pkg/env"
	"github.com/Kshitiz-Mhto/cryptix/utility"
	"github.com/sirupsen/logrus"
)

type EncryptedData struct {
	EncryptedMessage []byte `json:"encrypted_message"`
	EncryptedAESKey  []byte `json:"encrypted_aes_key"`
}

func HybridEncryption(plaintext []byte, pub *rsa.PublicKey) ([]byte, []byte, error) {
	logger.Logger.Info("Starting hybrid encryption process")
	// Generate a random 32-byte AES key.
	aesKey := make([]byte, 32)
	if _, err := rand.Read(aesKey); err != nil {
		utility.Error("failed to generate AES key: %v", err)
		logger.Logger.WithFields(logrus.Fields{"err": err}).Error("Failed to generate AES key")
		return nil, nil, err
	}

	// Encrypt the plaintext using AES-GCM.
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		utility.Error("failed to create AES cipher: %v", err)
		logger.Logger.WithFields(logrus.Fields{"err": err}).Error("Failed to create AES cipher")
		return nil, nil, err
	}

	// Create a new Galois/Counter Mode (GCM) block cipher mode using the AES block cipher.
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		utility.Error("failed to create AES-GCM mode: %v", err)
		logger.Logger.WithFields(logrus.Fields{"err": err}).Error("Failed to create AES-GCM mode")
		return nil, nil, err
	}

	// Generate a unique nonce (number used once) for AES-GCM encryption.
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		utility.Error("failed to generate nonce: %v", err)
		logger.Logger.WithFields(logrus.Fields{"err": err}).Error("Failed to generate nonce")
		return nil, nil, err
	}
	encryptedMsg := gcm.Seal(nonce, nonce, plaintext, nil)

	// Encrypt the AES key with RSA-OAEP.
	encryptedAESKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pub, aesKey, []byte(""))
	if err != nil {
		utility.Error("failed to encrypt AES key with RSA: %v", err)
		logger.Logger.WithFields(logrus.Fields{"err": err}).Error("Failed to encrypt AES key with RSA")
		return nil, nil, err
	}

	utility.Success("Message is encrypted successfully!")
	logger.Logger.Info("Message is encrypted successfully!")
	return encryptedMsg, encryptedAESKey, nil
}

func EncryptHybridData(encryptedMsg, encryptedAESKey []byte, outputFilePath, outputFileName string) error {
	data := EncryptedData{
		EncryptedMessage: encryptedMsg,
		EncryptedAESKey:  encryptedAESKey,
	}
	absOutputFilePath, err := filepath.Abs(outputFilePath)
	if err != nil {
		utility.Error("failed to represent absolute path: %v", err)
		logger.Logger.WithFields(logrus.Fields{"err": err}).Error("Representing absolute path")
		return err
	}

	if err := os.MkdirAll(outputFilePath, os.ModePerm); err != nil {
		utility.Error("failed to create output directory: %v", err)
		logger.Logger.WithFields(logrus.Fields{"err": err}).Error("Directory creation")
		return err
	}

	err = utility.ValidateFilename(outputFileName)
	if err != nil {
		utility.Error("Extension validation: %v", err)
		logger.Logger.WithFields(logrus.Fields{"err": err}).Error("Extension validation")
		return err
	}

	fullPath := filepath.Join(absOutputFilePath, outputFileName+env.Vars.JSON_FORMAT)
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		utility.Error("failed to marshal encrypted data: %v", err)
		logger.Logger.WithFields(logrus.Fields{"err": err}).Error("Marshal encryption")
		return err
	}

	if err := os.WriteFile(fullPath, jsonData, 0644); err != nil {
		utility.Error("failed to write encrypted data to file: %v", err)
		logger.Logger.WithFields(logrus.Fields{"err": err}).Error("Writing data to file")
		return err
	}

	utility.Success("Encrypted data file created successfully!!")
	logger.Logger.Infof("Encrypted data successfully saved to: %s", fullPath)
	return nil
}

// HybridDecryption decrypts the AES key with RSA-OAEP and then decrypts the message using AES-GCM.
func HybridDecryption(jsonFilePath string, privKey *rsa.PrivateKey) ([]byte, error) {
	logger.Logger.Info("Starting hybrid decryption process")

	jsonFilePath = filepath.Clean(jsonFilePath)
	data, err := os.ReadFile(jsonFilePath)
	if err != nil {
		utility.Error("Failed to read encrypted file: %s", err)
		logger.Logger.WithFields(logrus.Fields{
			"file": jsonFilePath,
			"err":  err,
		}).Error("Failed to read encrypted file")
		return nil, fmt.Errorf("failed to read encrypted file: %w", err)
	}

	var encryptedData EncryptedData
	if err := json.Unmarshal(data, &encryptedData); err != nil {
		utility.Error("Failed to parse encrypted JSON: %s", err)
		logger.Logger.WithFields(logrus.Fields{
			"file": jsonFilePath,
			"err":  err,
		}).Error("Failed to parse encrypted JSON")
		return nil, fmt.Errorf("failed to parse encrypted JSON: %w", err)
	}

	logger.Logger.Info("Successfully loaded encrypted data")

	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privKey, encryptedData.EncryptedAESKey, nil)
	if err != nil {
		utility.Error("RSA decryption failed: %s", err)
		logger.Logger.WithFields(logrus.Fields{
			"file": jsonFilePath,
			"err":  err,
		}).Error("RSA decryption failed")
		return nil, fmt.Errorf("RSA decryption failed: %w", err)
	}

	if len(aesKey) != 32 {
		utility.Error("Invalid AES key length: expected 32 bytes, got %d", len(aesKey))
		logger.Logger.WithFields(logrus.Fields{
			"expected": 32,
			"got":      len(aesKey),
		}).Error("Invalid AES key length")
		return nil, fmt.Errorf("invalid AES key length: expected 32 bytes, got %d", len(aesKey))
	}

	utility.Success("AES key successfully decrypted")
	logger.Logger.Info("AES key successfully decrypted")

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		utility.Error("AES initialization failed: %s", err)
		logger.Logger.WithFields(logrus.Fields{
			"err": err,
		}).Error("AES initialization failed")
		return nil, fmt.Errorf("AES initialization failed: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		utility.Error("GCM mode creation failed: %s", err)
		logger.Logger.WithFields(logrus.Fields{
			"err": err,
		}).Error("GCM mode creation failed")
		return nil, fmt.Errorf("GCM mode creation failed: %w", err)
	}

	// Ensure encrypted message contains the nonce
	nonceSize := gcm.NonceSize()
	if len(encryptedData.EncryptedMessage) < nonceSize {
		utility.Error("Malformed ciphertext: length %d is less than nonce size %d", len(encryptedData.EncryptedMessage), nonceSize)
		logger.Logger.WithFields(logrus.Fields{
			"length":     len(encryptedData.EncryptedMessage),
			"nonce_size": nonceSize,
		}).Error("Malformed ciphertext")
		return nil, fmt.Errorf("malformed ciphertext: length %d is less than nonce size %d", len(encryptedData.EncryptedMessage), nonceSize)
	}

	// Extract nonce and ciphertext
	nonce, ciphertext := encryptedData.EncryptedMessage[:nonceSize], encryptedData.EncryptedMessage[nonceSize:]

	// Decrypt message using AES-GCM
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		utility.Error("AES decryption failed: %s", err)
		logger.Logger.WithFields(logrus.Fields{
			"err": err,
		}).Error("AES decryption failed")
		return nil, fmt.Errorf("AES decryption failed: %w", err)
	}

	utility.Success("Decryption completed successfully!!")
	logger.Logger.Info("Decryption completed successfully!!")

	return plaintext, nil
}

func LoadPublicKey(path string) (*rsa.PublicKey, error) {
	absPathOfKey, err := filepath.Abs(path)
	if err != nil {
		utility.Error("Failed to get absolute path")
		logger.Logger.WithFields(logrus.Fields{
			"path": absPathOfKey,
			"err":  err,
		}).Error("Failed to get absolute path")
		return nil, err
	}

	pubKeyBytes, err := os.ReadFile(absPathOfKey)
	if err != nil {
		utility.Error("Failed to read public key file")
		logger.Logger.WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to read public key file")
		return nil, err
	}

	// Decode PEM block.
	block, _ := pem.Decode(pubKeyBytes)
	if block == nil {
		utility.Error("Invalid public key format")
		logger.Logger.WithFields(logrus.Fields{
			"err": err,
		}).Error("Invalid public key format")
		return nil, err
	}

	var pubKey *rsa.PublicKey

	switch block.Type {
	case "PUBLIC KEY":
		// PKIX format.
		pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			utility.Error("failed to parse public key (PKIX)")
			logger.Logger.WithFields(logrus.Fields{
				"err": err,
			}).Error("failed to parse public key (PKIX)")
			return nil, err
		}
		var ok bool
		pubKey, ok = pubInterface.(*rsa.PublicKey)
		if !ok {
			utility.Error("not an RSA public key")
			logger.Logger.Error("not an RSA public key")
			return nil, errors.New("not an RSA public key")
		}

	case "RSA PUBLIC KEY":
		// PKCS#1 format.
		pubKey, err = x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			utility.Error("failed to parse RSA public key (PKCS#1)")
			logger.Logger.WithFields(logrus.Fields{
				"err": err,
			}).Error("failed to parse RSA public key (PKCS#1)")
			return nil, err
		}

	default:
		utility.Error("unsupported public key type: %s", block.Type)
		logger.Logger.WithFields(logrus.Fields{
			"type": block.Type,
		}).Error("unsupported public key type")
		return nil, fmt.Errorf("unsupported public key type: %s", block.Type)
	}

	utility.Success("Public key file loaded successfully!")
	logger.Logger.Info("Public key file loaded successfully!")
	return pubKey, nil
}

func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	absPathOfKey, err := filepath.Abs(path)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"path": absPathOfKey,
			"err":  err,
		}).Error("Failed to get absolute path")
		return nil, err
	}

	privateKeyBytes, err := os.ReadFile(absPathOfKey)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to read private key file")
		return nil, err
	}

	// Decode PEM block.
	block, _ := pem.Decode(privateKeyBytes)
	if block == nil {
		logger.Logger.WithFields(logrus.Fields{
			"err": err,
		}).Error("Invalid private key format")
		return nil, err
	}

	// Try PKCS#1 first.
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"err": err,
		}).Error("failed to parse private key PKCS1")
		// Fall back to PKCS#8.
		key, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			logger.Logger.WithFields(logrus.Fields{
				"err": err,
			}).Error("failed to parse private key PKCS1")
			return nil, err2
		}
		rsaPriv, ok := key.(*rsa.PrivateKey)
		if !ok {
			utility.Error("not an RSA private key")
			logger.Logger.Error("not an RSA private key")
			return nil, errors.New("not an RSA private key")
		}

		utility.Success("Private key file loaded successfully!")
		logger.Logger.Info("Private key file loaded successfully!")
		return rsaPriv, nil
	}

	utility.Success("Private key file loaded successfully!")
	logger.Logger.Info("Private key file loaded successfully!")

	return priv, nil
}
