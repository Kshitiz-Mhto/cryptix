package utility

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gookit/color"
)

func GetBuildDate() string {
	filePath := "bin/cryptix"
	info, err := os.Stat(filePath)
	if err != nil {
		return ""
	}
	return info.ModTime().Format(time.RFC3339)
}

func ValidateFilename(filename string) error {
	ext := filepath.Ext(filename)

	if ext != "" {
		return errors.New("invalid filename: extensions are not allowed")
	}

	// Additional check to prevent filenames like "file."
	if strings.HasSuffix(filename, ".") {
		return errors.New("invalid filename: cannot end with a dot")
	}

	return nil
}

// Error is the function to handler all error in the Cli
func Error(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", color.Red.Sprintf("Error"), fmt.Sprintf(msg, args...))
}

// Info is the function to handler all info messages in the Cli
func Info(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", color.Blue.Sprintf("Info"), fmt.Sprintf(msg, args...))
}

// Warning is the function to handler all warnings in the Cli
func Warning(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", color.Yellow.Sprintf("Warning"), fmt.Sprintf(msg, args...))
}

func Success(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", color.Green.Sprintf("sucess"), fmt.Sprintf(msg, args...))
}

// YellowConfirm is the function to handler all delete confirm
func YellowConfirm(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s: %s", color.Warn.Sprintf("Warning"), fmt.Sprintf(msg, args...))
}
