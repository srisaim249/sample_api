package log_connection

import (
	"fmt"
	"log"
	"os"
	"sample_api/file"
	"sample_api/globals"
	"strings"
	"time"

	"github.com/fatih/color"
)

// InsertMsgIntoLogFile ... preparing the custom log format and writing data into file
// sample log filename format -- SMPP_CLIENT_2021-09-14_18.log
func InsertMsgIntoLogFile(filePath, line, level, msg string) {
	appName := globals.CONFIG_OBJ.APP_CONFIG.APP_NAME
	t := time.Now().Format("02-Jan-2006 15:04:05")
	totalMsg := fmt.Sprintf("[%v] [%v] [%v] [%v] \n", t, appName+" ==> "+file.GetFileNameFromFilePath(filePath)+globals.LINE_FIELD+line, strings.ToUpper(level), msg)
	coloredMsg := GetColoredString(level, totalMsg)

	logPath := globals.CONFIG_OBJ.APP_CONFIG.LOG_PATH
	logFileName := globals.CONFIG_OBJ.APP_CONFIG.LOG_FILE_NAME
	logFileName += "_" + time.Now().Format("2006-01-02_15")

	if len(logPath) == 0 || len(logFileName) == 0 {
		log.Fatal("logFile path or logFileName not provided in configuration")
	}
	file, err := os.OpenFile(logPath+"/"+logFileName+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	if err != nil {
		log.Println("Failed to open file ", err)
	}
	WriteDataInfoFile(file, totalMsg, coloredMsg, level)
	defer file.Close()
}

// WriteDataInfoFile ... writing the msg into the file , debug<info<warn<error<critical
// if system log level is info then we are not storing the below info level
func WriteDataInfoFile(file *os.File, totalMsg, coloredMsg, level string) {
	appLevelValue := globals.LOG_HEIRARCHY[level]
	systemLevelValue := globals.LOG_HEIRARCHY[globals.CONFIG_OBJ.APP_CONFIG.LOG_LEVEL]

	if appLevelValue >= systemLevelValue {
		_, err := file.Write([]byte(totalMsg))
		if err != nil {
			log.Println("Error while writing data into file ", err)
		}
		fmt.Print(coloredMsg)
	}
}

// GetColoredString ... applying color to the msg based on level
func GetColoredString(level, msg string) string {
	var coloredMsg string
	switch level {
	case "info":
		coloredMsg = color.GreenString(msg)
	case "warn":
		coloredMsg = color.YellowString(msg)
	case "error":
		coloredMsg = color.RedString(msg)
	case "debug":
		coloredMsg = color.BlueString(msg)
	default:
		coloredMsg = color.CyanString(msg)
	}
	return coloredMsg
}
