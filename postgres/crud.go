package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"runtime"
	logger "sample_api/log"
	"strconv"
	"time"
)

var defaultQueryTimeoutSec = 10

// ModifyDataInDB ... for insert/update/delete
func ModifyDataInDB(db *sql.DB, dbName, tableName, query string, values []interface{}, queryTimeOutSec int) (int64, int64, context.CancelFunc, error) {
	if queryTimeOutSec == 0 {
		queryTimeOutSec = defaultQueryTimeoutSec
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Duration(queryTimeOutSec)*time.Second)
	var err error
	err1 := CheckDBSession(db)
	if err1 != nil {
		_, filePath, lineNo, _ := runtime.Caller(0)
		logger.InsertMsgIntoLogFile(filePath, strconv.Itoa(lineNo+1), "error", fmt.Sprintf("lost db connection to database:%s, err:%s", dbName, err1.Error()))
		return -1, -1, cancelfunc, err1
	}
	if len(query) == 0 {
		_, filePath, lineNo, _ := runtime.Caller(0)
		logger.InsertMsgIntoLogFile(filePath, strconv.Itoa(lineNo+1), "error", fmt.Sprintf("Invalid query received:%s", query))
		return 0, 0, cancelfunc, errors.New("invalid query received")
	}
	tx, err := db.Begin()
	if err != nil {
		_, filePath, lineNo, _ := runtime.Caller(0)
		logger.InsertMsgIntoLogFile(filePath, strconv.Itoa(lineNo+1), "error", fmt.Sprintf("Error while Modifying data in  table:%s,err:%s", tableName, err.Error()))
		return 0, 0, cancelfunc, err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(query)
	// log.Printf("SQL:TABLE:%s,Query:%s,Args:%+v\n", tableName, query, values)
	if err != nil {
		_, filePath, lineNo, _ := runtime.Caller(0)
		logger.InsertMsgIntoLogFile(filePath, strconv.Itoa(lineNo+1), "error", fmt.Sprintf("Error while Modifying data in table:%s,err:%s", tableName, err.Error()))
		return 0, 0, cancelfunc, err
	}
	defer stmt.Close()
	result, err := stmt.ExecContext(ctx, values...)
	if err != nil {
		_, filePath, lineNo, _ := runtime.Caller(0)
		logger.InsertMsgIntoLogFile(filePath, strconv.Itoa(lineNo+1), "error", fmt.Sprintf("Error while executing data into  table:%s,err:%s", tableName, err.Error()))
		return 0, 0, cancelfunc, err
	}
	err = tx.Commit()
	if err != nil {
		_, filePath, lineNo, _ := runtime.Caller(0)
		logger.InsertMsgIntoLogFile(filePath, strconv.Itoa(lineNo+1), "error", fmt.Sprintf("Error while commi data into  table:%s,err:%s", tableName, err.Error()))
		return 0, 0, cancelfunc, err
	}
	lastInsertID, _ := result.LastInsertId()
	rowCnt, _ := result.RowsAffected()
	_, filePath, lineNo, _ := runtime.Caller(0)
	logger.InsertMsgIntoLogFile(filePath, strconv.Itoa(lineNo+1), "info", fmt.Sprintf("DBName:%s,TableName:%s,LastInsertedID:%d, Rows Affcted Count :%d\n", dbName, tableName, lastInsertID, rowCnt))
	return rowCnt, lastInsertID, cancelfunc, nil

}

// FetchDataFromDb ... fetching the data from sql
func FetchDataFromDb(db *sql.DB, dbName, tableName, query string, args []interface{}, queryTimeOutSec int) (*sql.Rows, context.CancelFunc, error) {
	if queryTimeOutSec == 0 {
		queryTimeOutSec = defaultQueryTimeoutSec
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Duration(queryTimeOutSec)*time.Second)
	var err error
	err1 := CheckDBSession(db)
	if err1 != nil {
		_, filePath, lineNo, _ := runtime.Caller(0)
		logger.InsertMsgIntoLogFile(filePath, strconv.Itoa(lineNo+1), "error", fmt.Sprintf("lost db connection to database:%s, err:%s", dbName, err1.Error()))
		return nil, cancelfunc, err1
	}
	// log.Printf("Table:%s,Query:%s,Args:%+v\n", tableName, query, args)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		_, filePath, lineNo, _ := runtime.Caller(0)
		logger.InsertMsgIntoLogFile(filePath, strconv.Itoa(lineNo+1), "error", fmt.Sprintf("Error while after reading rows:%s", err.Error()))
		return nil, cancelfunc, err
	}
	return rows, cancelfunc, nil
}
