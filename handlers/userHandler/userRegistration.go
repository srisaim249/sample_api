package userHandler

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"runtime"
	"sample_api/globals"
	logger "sample_api/log"
	"sample_api/postgres"
	"sample_api/structs"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// UserRegistrationHandler ... registring the user and checking the mandatory fields
func UserRegistrationHandler(responseWriter http.ResponseWriter, request *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			_, filePath, lineNo, _ := runtime.Caller(0)
			logger.InsertMsgIntoLogFile(filePath, strconv.Itoa(lineNo+1), "error", fmt.Sprintf("UserRegistrationHandler: panic error:%+v", err))
		}
	}()
	transactionID := fmt.Sprint(time.Now().UnixMilli())
	requestBody, err := FetchRequestBodyForUserRegistration(request)
	if err != nil {
		ErrorResponseForrequest(responseWriter, request, http.StatusBadRequest, err.Error(), transactionID)
		return
	}
	if err := CheckUserAlreadyExists(requestBody); err != nil { //if user already exist then returning error
		ErrorResponseForrequest(responseWriter, request, http.StatusInternalServerError, err.Error(), transactionID)
		return
	}
	_, filePath, lineNo, _ := runtime.Caller(0)
	logger.InsertMsgIntoLogFile(filePath, strconv.Itoa(lineNo+1), "info", fmt.Sprintf("txID:%v,payload:%+v", transactionID, requestBody))

	if err := InsertUserDataIntoDb(requestBody); err != nil {
		ErrorResponseForrequest(responseWriter, request, http.StatusInternalServerError, err.Error(), transactionID)
		return
	}
	response := make(map[string]interface{})
	response["message"] = globals.VERIFICATION_SENT
	SuccessResponse(transactionID, responseWriter, response)
}

// CheckUserAlreadyExists ...
func CheckUserAlreadyExists(payload structs.USER_REGISTRATION_REQUEST) error {
	rows, cancelFunc, err := fetchUserInfoFromDB(payload.USER_EMAIL, "", false)
	defer cancelFunc()
	if err != nil {
		return err
	}
	userInfo, err := fetchUserDetailsFromRows(rows)
	if err != nil {
		return err
	}
	if val, ok := userInfo["email"]; ok {
		if strings.EqualFold(fmt.Sprint(val), payload.USER_EMAIL) {
			return errors.New(globals.USER_EXISTS)
		}
	}
	return nil
}

// fetchUserDetailsFromRows ...
func fetchUserDetailsFromRows(rows *sql.Rows) (map[string]interface{}, error) {
	userInfo := make(map[string]interface{}, 0)
	defer rows.Close()
	var iID int
	var email, userName, password, firstName, lastname string
	for rows.Next() {
		err := rows.Scan(&iID, &email, &userName, &firstName, &lastname, &password)
		if err != nil {
			return userInfo, err
		}
		userInfo["iID"] = iID
		userInfo["username"] = userName
		userInfo["firstname"] = firstName
		userInfo["lastname"] = lastname
		userInfo["email"] = email
	}
	return userInfo, nil
}

// fetchUserInfoFromDB ...
func fetchUserInfoFromDB(email, password string, isLogin bool) (*sql.Rows, context.CancelFunc, error) {
	args := make([]interface{}, 0)
	args = append(args, email)
	query := globals.CONFIG_OBJ.QUERY_CONFIG.FETCH_USER_BY_EMAIL.QUERY
	queryTimeOut := globals.CONFIG_OBJ.QUERY_CONFIG.FETCH_USER_BY_EMAIL.TIME_OUT_SEC

	if isLogin {
		args = append(args, password)
		query = globals.CONFIG_OBJ.QUERY_CONFIG.LOGIN_UESR_QUERY.QUERY
		queryTimeOut = globals.CONFIG_OBJ.QUERY_CONFIG.LOGIN_UESR_QUERY.TIME_OUT_SEC
	}

	dbName := globals.CONFIG_OBJ.SQL_CONNECTION.SQL_DATABASE
	tableName := globals.CONFIG_OBJ.SQL_CONNECTION.SQL_TABLE
	rows, cancelFunc, err := postgres.FetchDataFromDb(globals.USER_DB_CONN, dbName, tableName, query, args, queryTimeOut)
	// cancelFunc - releases the resources, once operation completes then its will release the memory
	return rows, cancelFunc, err

}

// InsertUserDataIntoDb ...
func InsertUserDataIntoDb(payload structs.USER_REGISTRATION_REQUEST) error {
	args := make([]interface{}, 0)

	args = append(args, payload.USER_EMAIL)
	args = append(args, payload.USER_NAME)
	args = append(args, payload.FIRST_NAME)
	args = append(args, payload.LAST_NAME)
	args = append(args, payload.PASSWORD)

	query := globals.CONFIG_OBJ.QUERY_CONFIG.CREATE_USER.QUERY
	queryTimeOut := globals.CONFIG_OBJ.QUERY_CONFIG.CREATE_USER.TIME_OUT_SEC

	dbName := globals.CONFIG_OBJ.SQL_CONNECTION.SQL_DATABASE
	tableName := globals.CONFIG_OBJ.SQL_CONNECTION.SQL_TABLE
	_, _, cancelFunc, err := postgres.ModifyDataInDB(globals.USER_DB_CONN, dbName, tableName, query, args, queryTimeOut)
	defer cancelFunc() // releases the resources, once operation completes then its will release the memory
	if err != nil {
		return err
	}
	_, filePath, lineNo, _ := runtime.Caller(0)
	logger.InsertMsgIntoLogFile(filePath, strconv.Itoa(lineNo+1), "info", ("user registered successfully"))
	return nil

}

// FetchRequestBody ...
func FetchRequestBodyForUserRegistration(request *http.Request) (structs.USER_REGISTRATION_REQUEST, error) {
	requestBodyStruct := structs.USER_REGISTRATION_REQUEST{}
	requestBodyInfo, err := io.ReadAll(request.Body)
	if err != nil {
		return requestBodyStruct, err
	}
	err = json.Unmarshal(requestBodyInfo, &requestBodyStruct)
	if err != nil {
		return requestBodyStruct, err
	}
	errForUserRegistration := ValidateUserRegistration(&requestBodyStruct)
	if errForUserRegistration != nil {
		return requestBodyStruct, errForUserRegistration
	}
	return requestBodyStruct, nil
}

// ValidateUserRegistration ... validating the request body given by user
func ValidateUserRegistration(userReqBody *structs.USER_REGISTRATION_REQUEST) error {

	userReqBody.USER_NAME = strings.TrimSpace(userReqBody.USER_NAME)
	userReqBody.FIRST_NAME = strings.TrimSpace(userReqBody.FIRST_NAME)
	userReqBody.LAST_NAME = strings.TrimSpace(userReqBody.LAST_NAME)

	if len(userReqBody.FIRST_NAME) == 0 {
		return fmt.Errorf("please provide your firstname")
	}
	if len(userReqBody.LAST_NAME) == 0 {
		return fmt.Errorf("please provide your lastname")
	}
	if len(userReqBody.USER_NAME) == 0 {
		return fmt.Errorf("please provide your username")
	}
	if !verifyPassword(userReqBody.PASSWORD) {
		return fmt.Errorf("please provide password with a digit,1-lowercase,1-uppercase,1-special character and length should be greater than 7 characters")
	}
	userReqBody.PASSWORD = fmt.Sprintf("%x", md5.Sum([]byte(userReqBody.PASSWORD)))
	if len(userReqBody.USER_EMAIL) == 0 || !IsvalidEmail(userReqBody.USER_EMAIL) {
		return fmt.Errorf("please provide valid email")
	}
	return nil
}

// IsvalidEmail ... checking the user email is valid or not
func IsvalidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

/*
	verifyPassword ... checking  the password with one special character, \

one lower case, one upper case character, one digits
*/
func verifyPassword(pwd string) bool {
	var hasNumber, hasUpperCase, hasLowercase, hasSpecial bool
	specialChars := "[$&+,:;=?@#|'<>.-^*()%!]"
	for _, char := range pwd {
		switch {
		case unicode.IsNumber(char) && !hasNumber:
			hasNumber = true
		case unicode.IsUpper(char) && !hasUpperCase:
			hasUpperCase = true
		case unicode.IsLower(char) && !hasLowercase:
			hasLowercase = true
		case strings.Contains(specialChars, string(char)) && !hasSpecial:
			hasSpecial = true
		}
	}
	return hasNumber && hasUpperCase && hasLowercase && hasSpecial && len(pwd) > 7
}

// ErrorResponseForrequest ...
func ErrorResponseForrequest(responseWriter http.ResponseWriter, request *http.Request, errorCode int, errorDesc string, requestID string) {
	_, filePath, lineNo, _ := runtime.Caller(0)
	logger.InsertMsgIntoLogFile(filePath, strconv.Itoa(lineNo+1), "error", fmt.Sprintf("received error message:%+v", errorDesc))
	ErrorStructure := structs.ERROR_RESPONSE_STRCUT{
		TransactionID:    requestID,
		ErrorDescription: errorDesc,
	}

	responseJSON, _ := json.Marshal(ErrorStructure)
	responseWriter.Header().Set(globals.CONTENT_TYPE_FIELD, globals.CONTENT_TYPE)
	responseWriter.WriteHeader(errorCode)
	responseWriter.Write(responseJSON)
}

// SuccessResponse ...
func SuccessResponse(transID string, responseWriter http.ResponseWriter, response map[string]interface{}) {

	response[globals.TRX_ID_FIELD] = transID
	bytesInfo, _ := json.Marshal(response)

	responseWriter.Header().Set(globals.CONTENT_TYPE_FIELD, globals.CONTENT_TYPE)
	responseWriter.WriteHeader(http.StatusOK)
	responseWriter.Write([]byte(bytesInfo))

}
