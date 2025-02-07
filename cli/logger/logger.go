package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Logger = logrus.New()

func InitLogger() {
	file, err := os.OpenFile("cryptix.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		Logger.SetOutput(file)
	} else {
		Logger.Info("Failed to log to file, using default stderr")
	}

	Logger.SetFormatter(&logrus.JSONFormatter{})
	Logger.SetLevel(logrus.DebugLevel)
}
