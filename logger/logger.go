package logger

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

var (
	infoStyle  = color.New(color.FgGreen, color.Bold).SprintfFunc()
	errorStyle = color.New(color.FgRed, color.Bold).SprintfFunc()
	debugStyle = color.New(color.FgBlue).SprintfFunc()
	warnStyle  = color.New(color.FgYellow).SprintFunc()
	textStyle  = color.RGB(175, 175, 175).SprintFunc()
)

func LogMessage(level string, message string, prefix ...string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	level = strings.ToUpper(level)

	var coloredLevel string
	switch level {
	case "INFO":
		coloredLevel = infoStyle("[INFO]")
	case "ERROR":
		coloredLevel = errorStyle("[ERROR]")
	case "DEBUG":
		coloredLevel = debugStyle("[DEBUG]")
	case "WARN":
		coloredLevel = warnStyle("[WARN]")
	default:
		coloredLevel = fmt.Sprintf("[%s]", level)
	}

	prefixString := strings.Join(prefix, "][")

	fmt.Printf("[%s]%s %s %s\n", prefixString, coloredLevel, timestamp, textStyle(message))
}