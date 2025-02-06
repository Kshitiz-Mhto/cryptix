package utility

import "github.com/gookit/color"

// Green is the function to convert str to green in console
func Green(value string) string {
	newColor := color.FgGreen.Render
	return newColor(value)
}

// Yellow is the function to convert str to yellow in console
func Yellow(value string) string {
	newColor := color.New(color.FgYellow).Render
	return newColor(value)
}

// Red is the function to convert str to red in console
func Red(value string) string {
	newColor := color.New(color.FgRed).Render
	return newColor(value)
}
