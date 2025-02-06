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
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"os"
	"path/filepath"

	"github.com/Kshitiz-Mhto/stegomail/cli/logger"
	"github.com/Kshitiz-Mhto/stegomail/pkg/env"
	"github.com/Kshitiz-Mhto/stegomail/utility"
	"github.com/disintegration/imaging"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	imageFile          string
	msg                string
	encryptedImagePath string
	pubkeyPath         string
	outputFileName     string
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
	encryptedImagePath, _ = cmd.Flags().GetString("output")
	outputFileName, _ = cmd.Flags().GetString("name")

	if msg == "" {
		utility.Error("Message to be encrypted  is empty.")
		logger.Logger.Fatal("Message to be encrypted is empty")
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

	err = EmbedHybridData(imageFile, encryptedMsg, encryptedAESKey, encryptedImagePath, outputFileName)
	if err != nil {
		utility.Info("Aborting operation process : %s", utility.Red("Embedding msg in image"))
		os.Exit(1)
	}
}

func init() {
	EmbadeCmd.Flags().StringVarP(&imageFile, "image", "i", "", "Specify the path of image that will be embaded with message. [*Required]")
	EmbadeCmd.Flags().StringVarP(&msg, "message", "m", "", "Specify your message that will be encoded and embedded. [*Required]")
	EmbadeCmd.Flags().StringVarP(&encryptedImagePath, "output", "o", ".", "Specify the directory where embeded image will reside. [Default path: current directory]")
	EmbadeCmd.Flags().StringVarP(&pubkeyPath, "pubkey", "k", "", "Specify your public key file path. [*Required]")
	EmbadeCmd.Flags().StringVarP(&outputFileName, "name", "n", "", "Specify your output image file name. [*Required]")

	EmbadeCmd.MarkFlagsRequiredTogether("image", "message", "pubkey", "name")
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

func EmbedHybridData(imgPath string, encryptedMsg, encryptedAESKey []byte, outputPath, imgName string) error {
	absPathOfImage, err := filepath.Abs(imageFile)
	if err != nil {
		utility.Error("failed to get absolute path: %v", err)
		logger.Logger.WithFields(logrus.Fields{
			"path": absPathOfImage,
			"err":  err,
		}).Error("failed to get absolute path")
		return err
	}

	absOutputPath, err := filepath.Abs(outputPath)
	if err != nil {
		utility.Error("failed to get absolute path: %v", err)
		logger.Logger.WithFields(logrus.Fields{
			"path": absOutputPath,
			"err":  err,
		}).Error("failed to get absolute path")
		return err
	}

	if err := os.MkdirAll(absOutputPath, 0700); err != nil {
		utility.Error("failed to create directory %s: %v", absOutputPath, err)
		logger.Logger.WithFields(logrus.Fields{
			"path": absOutputPath,
			"err":  err,
		}).Fatal("failed to create directory")
	}

	// Load the image
	src, err := imaging.Open(absPathOfImage)
	if err != nil {
		utility.Error("failed to open image: %v", err)
		logger.Logger.WithFields(logrus.Fields{"err": err, "path": absPathOfImage}).Error("Failed to open image")
		return err
	}
	dst := image.NewNRGBA(src.Bounds())
	draw.Draw(dst, dst.Bounds(), src, image.Point{}, draw.Src)

	bounds := dst.Bounds()
	payload := append(encryptedMsg, encryptedAESKey...)

	// Convert payload to bitstream
	bitStream, err := BytesToBits(payload)
	if err != nil {
		utility.Error("%s", err)
		logger.Logger.WithFields(logrus.Fields{"err": err}).Error("input data cannot be nil")
		return err
	}
	if len(bitStream) > bounds.Max.X*bounds.Max.Y {
		utility.Error("payload is too large to embed in the image")
		logger.Logger.Error("Payload is too large to embed in the given image")
		return errors.New("")
	}

	// Embed bits into DCT coefficients (Pseudo-code)
	bitIndex := 0
	for y := 0; y < bounds.Max.Y; y += 8 {
		for x := 0; x < bounds.Max.X; x += 8 {
			block, err := Get8x8Block(dst, x, y)
			if err != nil {
				utility.Error("%s", err)
				logger.Logger.WithFields(logrus.Fields{"err": err}).Error("Error during geting image 8 by 8 blcok bit")
				return err
			}
			dctBlock, err := ApplyDCT(block)
			if err != nil {
				utility.Error("%s", err)
				logger.Logger.WithFields(logrus.Fields{"err": err}).Error("Error during applying DCT")
				return err
			}

			// Embed bits into DCT block
			bitIndex, err = EmbedBitsInDCT(dctBlock, bitStream, bitIndex)
			if err != nil {
				utility.Error("%s", err)
				logger.Logger.WithFields(logrus.Fields{"err": err}).Error("Error during embedding bits into DCT block")
				return err
			}

			block, err = ApplyInverseDCT(dctBlock)
			if err != nil {
				utility.Error("%s", err)
				logger.Logger.WithFields(logrus.Fields{"err": err}).Error("Applying Inverse SCT")
				return err
			}
			Set8x8Block(dst, x, y, block)

			// Stop if all bits are embedded
			if bitIndex >= len(bitStream) {
				break
			}
		}
		if bitIndex >= len(bitStream) {
			break
		}
	}

	outputImgPath := filepath.Join(absOutputPath, imgName+env.Vars.JPG_FORMAT)

	// Save the modified image
	err = imaging.Save(dst, outputImgPath)
	if err != nil {
		utility.Error("failed to save embedded image: %v", err)
		logger.Logger.WithFields(logrus.Fields{"err": err, "outputFilePath": outputImgPath}).Error("Failed to save embedded image")
		return err
	}

	utility.Success("Successfully embedded data into the image")
	logger.Logger.Info("Successfully embedded data into the image")
	return nil
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

// Convert byte slice to bit stream (MSB first)
func BytesToBits(data []byte) ([]int, error) {
	if data == nil {
		return nil, errors.New("input data cannot be nil")
	}
	bits := make([]int, 0, len(data)*8)
	for _, b := range data {
		for i := 7; i >= 0; i-- {
			bits = append(bits, int((b>>i)&1))
		}
	}
	return bits, nil
}

func Get8x8Block(img image.Image, x, y int) ([][]color.Color, error) {
	if img == nil {
		return nil, errors.New("image cannot be nil")
	}
	bounds := img.Bounds()
	block := make([][]color.Color, 8)
	for j := 0; j < 8; j++ {
		row := make([]color.Color, 8)
		for i := 0; i < 8; i++ {
			px, py := x+i, y+j
			if px >= bounds.Max.X || py >= bounds.Max.Y {
				row[i] = color.Gray{0} // Pad out-of-bounds pixels with black
			} else {
				row[i] = img.At(px, py)
			}
		}
		block[j] = row
	}
	return block, nil
}

// Apply 2D DCT to 8x8 block (naive implementation)
func ApplyDCT(block [][]color.Color) ([][]float64, error) {
	if len(block) != 8 || len(block[0]) != 8 {
		return nil, errors.New("input block must be 8x8")
	}
	dctBlock := make([][]float64, 8)
	for v := 0; v < 8; v++ {
		dctBlock[v] = make([]float64, 8)
		for u := 0; u < 8; u++ {
			sum := 0.0
			for y := 0; y < 8; y++ {
				for x := 0; x < 8; x++ {
					gray := color.GrayModel.Convert(block[y][x]).(color.Gray)
					pixel := float64(gray.Y) - 128 // Center around 0

					cu, cv := 1.0, 1.0
					if u == 0 {
						cu = 1 / math.Sqrt2
					}
					if v == 0 {
						cv = 1 / math.Sqrt2
					}

					cosX := math.Cos((float64(2*x+1) * float64(u) * math.Pi) / 16)
					cosY := math.Cos((float64(2*y+1) * float64(v) * math.Pi) / 16)
					sum += cu * cv * pixel * cosX * cosY
				}
			}
			dctBlock[v][u] = sum / 4
		}
	}
	return dctBlock, nil
}

// Embed bits in DCT coefficients (skip DC coefficient)
func EmbedBitsInDCT(dctBlock [][]float64, bitStream []int, bitIndex int) (int, error) {
	if len(dctBlock) != 8 || len(dctBlock[0]) != 8 {
		return bitIndex, errors.New("DCT block must be 8x8")
	}
	if bitStream == nil {
		return bitIndex, errors.New("bit stream cannot be nil")
	}

	zigzag := []struct{ x, y int }{
		{1, 0}, {0, 1}, {0, 2}, {1, 1}, {2, 0}, {3, 0}, {2, 1}, {1, 2},
		// Complete zigzag pattern up to 63 coefficients
	}

	for _, pos := range zigzag {
		if bitIndex >= len(bitStream) {
			return bitIndex, nil
		}
		coeff := math.Trunc(dctBlock[pos.y][pos.x])
		dctBlock[pos.y][pos.x] = coeff + math.Copysign(float64(bitStream[bitIndex]), coeff)
		bitIndex++
	}
	return bitIndex, nil
}

// Apply inverse DCT
func ApplyInverseDCT(dctBlock [][]float64) ([][]color.Color, error) {
	if len(dctBlock) != 8 || len(dctBlock[0]) != 8 {
		return nil, errors.New("DCT block must be 8x8")
	}

	block := make([][]color.Color, 8)
	for y := 0; y < 8; y++ {
		block[y] = make([]color.Color, 8)
		for x := 0; x < 8; x++ {
			sum := 0.0
			for v := 0; v < 8; v++ {
				for u := 0; u < 8; u++ {
					cu, cv := 1.0, 1.0
					if u == 0 {
						cu = 1 / math.Sqrt2
					}
					if v == 0 {
						cv = 1 / math.Sqrt2
					}
					cosX := math.Cos((float64(2*x+1) * float64(u) * math.Pi) / 16)
					cosY := math.Cos((float64(2*y+1) * float64(v) * math.Pi) / 16)
					sum += cu * cv * dctBlock[v][u] * cosX * cosY
				}
			}
			val := math.Round(sum/4) + 128
			val = math.Max(0, math.Min(255, val))
			block[y][x] = color.Gray{uint8(val)}
		}
	}
	return block, nil
}

// Set modified 8x8 block back into image (NRGBA format)
func Set8x8Block(img draw.Image, x, y int, block [][]color.Color) error {
	if img == nil {
		return errors.New("image cannot be nil")
	}
	if len(block) != 8 || len(block[0]) != 8 {
		return errors.New("block must be 8x8")
	}

	bounds := img.Bounds()
	for j := 0; j < 8; j++ {
		for i := 0; i < 8; i++ {
			px := x + i
			py := y + j
			if px >= bounds.Max.X || py >= bounds.Max.Y {
				continue
			}
			img.Set(px, py, block[j][i])
		}
	}
	return nil
}
