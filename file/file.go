package file

import (
	"errors"
	"path/filepath"
	"strings"
)

// GetFileNameFromFilePath ... taking filename from path
// login_otp/utils/file/filePath.go >> filePath.go
func GetFileNameFromFilePath(path string) string {
	arr := strings.Split(path, "/")
	return arr[len(arr)-1]
}

// GetFilePathFromArgs ... taking the input from the user
// returning error if configuration file is missing
func GetFilePathFromArgs(input []string) (string, error) {
	var filePathName string
	if len(input) == 1 {
		return filePathName, errors.New("please provide the configuration file path")
	}
	filePathName = input[1]
	if filepath.Ext(filePathName) != ".json" {
		return filePathName, errors.New("please provide the file with .json extension")
	}
	return filePathName, nil
}
