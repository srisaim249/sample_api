package config

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sample_api/globals"
	logger "sample_api/log"
	"strconv"

	"github.com/go-playground/validator"
)

// ReadConfigFromFile ... reading data from json file and storing it into the Object
func ReadConfigFromFile(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error while reading json file:%s", err)
	}
	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'CONFIG_OBJECT' which we defined above
	err = json.Unmarshal(content, &globals.CONFIG_OBJ)
	if err != nil {
		return fmt.Errorf("error while fetching data from object:%s", err.Error())
	}
	validate := validator.New()
	err = validate.Struct(globals.CONFIG_OBJ)
	if err != nil {
		return fmt.Errorf("error while validating structure data:%s", err.Error())
	}
	_, filePath, lineNo, _ := runtime.Caller(0)
	logger.InsertMsgIntoLogFile(filePath, strconv.Itoa(lineNo+1), "info", "config loaded sucessfully")

	return nil
}
