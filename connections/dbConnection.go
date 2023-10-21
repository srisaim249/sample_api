package connections

import (
	"fmt"
	"runtime"
	"sample_api/globals"
	logger "sample_api/log"
	postgres "sample_api/postgres"
	"strconv"
)

// PrepareUserDbConnection ...
func PrepareUserDbConnection() error {
	connection, err := postgres.GetSession(globals.CONFIG_OBJ.SQL_CONNECTION)
	if err != nil {
		_, filePath, lineNo, _ := runtime.Caller(0)
		logger.InsertMsgIntoLogFile(filePath, strconv.Itoa(lineNo+1), "error", fmt.Sprintf("error occured while connecting to the database,err:%+v", err))
		return err
	}
	globals.USER_DB_CONN = connection
	return nil
}
