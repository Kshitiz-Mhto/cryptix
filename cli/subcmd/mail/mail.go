package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Kshitiz-Mhto/cryptix/cli/logger"
	"github.com/Kshitiz-Mhto/cryptix/pkg/env"
	"github.com/Kshitiz-Mhto/cryptix/utility"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v2"
	"google.golang.org/api/option"
)

func sendHtmlEmailWithRetry(to []string, subject string, htmlBody string, maxRetries int, retryInterval time.Duration) error {
	auth := smtp.PlainAuth(
		"cryptrix",
		env.Vars.FromEmail,
		env.Vars.FromEmailPassword,
		env.Vars.FromEmailSMTP,
	)

	headers := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";"

	message := "Subject: " + subject + "\n" + headers + "\n\n" + htmlBody

	var lastError error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		lastError = smtp.SendMail(
			env.Vars.SMTPAddress,
			auth,
			env.Vars.FromEmail,
			to,
			[]byte(message),
		)
		if lastError == nil {
			return nil
		}

		// If failed, wait before retrying
		utility.Warning("Attempt %d failed: %s. Retrying in %v...", attempt, lastError.Error(), retryInterval)
		time.Sleep(retryInterval)

		// Exponentially increase the retry interval for next attempt
		retryInterval = retryInterval * 2
	}

	logger.Logger.WithFields(logrus.Fields{
		"failed": lastError.Error(),
	}).Info("Retry mechanism stats")
	utility.Error("Failed to send email after %d attempts: %s", maxRetries, lastError.Error())
	return lastError
}

func HTMLTemplateMailHandler(addr, subject string, vars map[string]interface{}) bool {
	logger.Logger.Info("Email sending initialization")
	var emailSubject string
	basePathForEmailHtml := "./static/"

	if subject == "" {
		emailSubject = env.Vars.SUBJECT_DESC
	}

	to := strings.Split(addr, ",")

	// Parse the HTML template
	templatePath := filepath.Join(basePathForEmailHtml, env.Vars.HTML_TEMPLATE)
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		utility.Error("failed to parse template: %v", err)
		logger.Logger.Errorf("failed to parse template: %v", err)
		return false
	}

	// Render the template with the map data
	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, vars); err != nil {
		utility.Error("failed to render template: %v", err)
		logger.Logger.Errorf("failed to render template: %v", err)
		return false
	}

	// Define max retries and initial retry interval
	maxRetries := 3
	initialRetryInterval := 2 * time.Second

	// Attempt to send the email with retry logic
	err = sendHtmlEmailWithRetry(to, emailSubject, rendered.String(), maxRetries, initialRetryInterval)
	if err != nil {
		utility.Error("%s", err.Error())
		logger.Logger.Errorf("failed to send mail: %v", err)
		return false
	}

	return true
}

func UploadFileToGoogleDrive(filePath string) (string, error) {
	credentialsPath := env.Vars.OAUTH_CREDENTIALS_PATH
	b, err := os.ReadFile(credentialsPath)
	if err != nil {
		utility.Error("Unable to read credentials file: %s", err)
		logger.Logger.WithFields(logrus.Fields{
			"credentialsPath": credentialsPath,
			"err":             err,
		}).Error("Unable to read credentials file")
		return "", err
	}

	// Parse the credentials and create a config.
	config, err := google.ConfigFromJSON(b, drive.DriveFileScope)
	if err != nil {
		utility.Error("Unable to parse credentials file: %s", err)
		logger.Logger.WithFields(logrus.Fields{
			"credentialsPath": credentialsPath,
			"err":             err,
		}).Error("Unable to parse credentials file")
		return "", err
	}

	utility.Success("OAuth2 credentials loaded successfully")
	logger.Logger.Info("OAuth2 credentials loaded successfully")

	// Obtain an authenticated HTTP client.
	client := getClient(config)

	// Create a new Drive service using the authenticated client.
	service, err := drive.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		utility.Error("Unable to create Drive service: %s", err)
		logger.Logger.WithFields(logrus.Fields{
			"filePath": filePath,
			"err":      err,
		}).Error("Unable to create Drive service")
		return "", err
	}

	// Open the file for upload.
	f, err := os.Open(filePath)
	if err != nil {
		utility.Error("Unable to open file for upload: %s", err)
		logger.Logger.WithFields(logrus.Fields{
			"filePath": filePath,
			"err":      err,
		}).Error("Unable to open file for upload")
		return "", err
	}
	defer f.Close()

	// Create file metadata for Google Drive.
	driveFile := &drive.File{
		Description: filepath.Base(filePath),
	}

	logger.Logger.WithFields(logrus.Fields{
		"filePath": filePath,
	}).Info("Uploading file to Google Drive...")

	// Upload the file.
	uploadedFile, err := service.Files.Insert(driveFile).Media(f).Do()
	if err != nil {
		utility.Error("Unable to open file for upload to GD: %s", err)
		logger.Logger.WithFields(logrus.Fields{
			"filePath": filePath,
			"err":      err,
		}).Error("Unable to upload file to Google Drive")
		return "", err
	}

	logger.Logger.WithFields(logrus.Fields{
		"filePath": uploadedFile.Description,
		"fileId":   uploadedFile.Id,
	}).Info("File uploaded successfully to Google Drive")

	// Set the file's permissions to public (read-only).
	permission := &drive.Permission{
		Type: "anyone",
		Role: "read",
	}
	_, err = service.Permissions.Insert(uploadedFile.Id, permission).Do()
	if err != nil {
		utility.Error("Unable to update file permissions: %s", err)
		logger.Logger.WithFields(logrus.Fields{
			"filePath": filePath,
			"err":      err,
		}).Error("Unable to update file permissions")
		return "", err
	}

	logger.Logger.WithFields(logrus.Fields{
		"fileId": uploadedFile.Id,
	}).Info("File permissions updated to allow public access")
	utility.Success("Successfully uploaded file to google drive")
	logger.Logger.Info("Successfully uploaded file to google drive")

	// Construct and return the shareable download link.
	fileLink := fmt.Sprintf("https://drive.google.com/file/d/%s/view?usp=sharing", uploadedFile.Id)
	return fileLink, nil
}

func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		utility.Error("%s", err)
		logger.Logger.Errorf("%s", err)
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
