package postgres

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"runtime"
	logger "sample_api/log"
	"sample_api/structs"
	"strconv"
	"time"

	"github.com/go-playground/validator"
	_ "github.com/lib/pq"
)

// GetSession ... To get the pgsql connection with retry
func GetSession(sqlStruct structs.SQL_CONNECTION_DETAILS) (*sql.DB, error) {
	validate := validator.New()
	err := validate.Struct(sqlStruct)
	if err != nil {
		return nil, fmt.Errorf("error while validating structure data:%s", err.Error())
	}
	delayTime := sqlStruct.SQL_DELAY_TIME
	connString, err := PrepareConnectionURL(sqlStruct)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open(sqlStruct.SQL_DRIVER_NAME, connString)
	counter := 0
	//retrying for the sql connection
	for (err != nil || db == nil) && counter < sqlStruct.SQL_RETRY_COUNT {
		_, filePath, lineNo, _ := runtime.Caller(0)
		logger.InsertMsgIntoLogFile(filePath, strconv.Itoa(lineNo+1), "error", fmt.Sprintf("Error detected!:%s in SQL, Attemping to re-connect...%d times", err, counter))
		time.Sleep(time.Duration(delayTime) * time.Second)
		db, err = sql.Open(sqlStruct.SQL_DRIVER_NAME, connString)
		if err == nil && db != nil && CheckDBSession(db) != nil {
			break
		}
		counter += 1
	}
	err = CheckDBSession(db)
	if err != nil { //checkfor connection is open or not
		return nil, err
	}
	_, filePath, lineNo, _ := runtime.Caller(0)
	logger.InsertMsgIntoLogFile(filePath, strconv.Itoa(lineNo+1), "info", fmt.Sprintf("pgsql connected successfuly,host:%s,port:%s,userName:%s,databse:%s", sqlStruct.SQL_HOST, sqlStruct.SQL_PORT, sqlStruct.SQL_USERNAME, sqlStruct.SQL_DATABASE))

	return db, nil
}

// CheckDBSession ... checking whether the connection stablished or not
// it should return success response if it is connected
func CheckDBSession(db *sql.DB) error {
	ctx, cancelfunc := context.WithTimeout(context.Background(), time.Second*3)

	/*
		Canceling this context releases resources associated with it, so code should
		call cancel as soon as the operations running in this Context complete:
	*/
	defer cancelfunc()
	return db.PingContext(ctx)
}

// PrepareConnectionURL ... using server details preparing the sql url
func PrepareConnectionURL(sqlStruct structs.SQL_CONNECTION_DETAILS) (string, error) {

	host := sqlStruct.SQL_HOST
	user := sqlStruct.SQL_USERNAME
	pwd := url.QueryEscape(sqlStruct.SQL_PASSWORD) //converts the special characters to encoded values
	port := sqlStruct.SQL_PORT
	sslMode := sqlStruct.SQL_SSL_MODE
	dbName := sqlStruct.SQL_DATABASE
	encodedUrl := fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=%s", sqlStruct.SQL_DRIVER_NAME, user, string(pwd), host, port, dbName, sslMode)

	if sqlStruct.TLS_ENABLED { //based on tls secure adding the certs to connection
		// sslrootcert=root.crt sslkey=client.key sslcert=client.crt
		rootCert := sqlStruct.CA_CERT_PATH
		clientKey := sqlStruct.CLIENT_KEY_PATH
		clientCert := sqlStruct.CLIENT_CERT_PATH
		encodedUrl = fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=%s&sslrootcert=%s&sslkey=%s&sslcert=%s", sqlStruct.SQL_DRIVER_NAME, user, string(pwd), host, port, dbName, sslMode, rootCert, clientKey, clientCert)
	}
	return encodedUrl, nil
}

// GetSqlTLSDetails ... preparing the tls details of sql
func GetSqlTLSDetails(sqlStruct structs.SQL_CONNECTION_DETAILS) (*tls.Config, error) {
	tlsConfig := new(tls.Config)
	certBytes, err := os.ReadFile(sqlStruct.CA_CERT_PATH)
	if err != nil {
		_, filePath, lineNo, _ := runtime.Caller(0)
		logger.InsertMsgIntoLogFile(filePath, strconv.Itoa(lineNo+1), "error", fmt.Sprintf("Failed to read sql ca file:%s", err))
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(certBytes); !ok {
		_, filePath, lineNo, _ := runtime.Caller(0)
		logger.InsertMsgIntoLogFile(filePath, strconv.Itoa(lineNo+1), "error", fmt.Sprintf("Failed to parse sql ca file:%s", err))
		return nil, err
	}
	cert, err := tls.LoadX509KeyPair(sqlStruct.CLIENT_CERT_PATH, sqlStruct.CLIENT_KEY_PATH)
	if err != nil {
		_, filePath, lineNo, _ := runtime.Caller(0)
		logger.InsertMsgIntoLogFile(filePath, strconv.Itoa(lineNo+1), "error", fmt.Sprintf("LoadX509KeyPair is failed,error:%s", err.Error()))
		return tlsConfig, err
	}
	tlsConfig.Certificates = append(tlsConfig.Certificates, cert)
	tlsConfig.InsecureSkipVerify = true
	tlsConfig.RootCAs = caCertPool
	return tlsConfig, nil
}
