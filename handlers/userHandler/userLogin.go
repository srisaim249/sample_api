package userHandler

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"sample_api/globals"
	logger "sample_api/log"
	"sample_api/structs"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// UserLoginHandler ...
func UserLoginHandler(responseWriter http.ResponseWriter, request *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			_, filePath, lineNo, _ := runtime.Caller(0)
			logger.InsertMsgIntoLogFile(filePath, strconv.Itoa(lineNo+1), "error", fmt.Sprintf("UserLoginHandler: panic error:%+v", err))
		}
	}()
	transactionID := fmt.Sprint(time.Now().UnixMilli())
	requestBody, err := FetchRequestBodyForUserLogin(request)
	if err != nil {
		ErrorResponseForrequest(responseWriter, request, http.StatusBadRequest, err.Error(), transactionID)
		return
	}
	rows, cancelFunc, err := fetchUserInfoFromDB(requestBody.USER_EMAIL, requestBody.PASSWORD, true)
	defer cancelFunc()
	if err != nil {
		ErrorResponseForrequest(responseWriter, request, http.StatusBadRequest, err.Error(), transactionID)
		return
	}
	userInfo, err := fetchUserDetailsFromRows(rows)
	if err != nil {
		ErrorResponseForrequest(responseWriter, request, http.StatusBadRequest, err.Error(), transactionID)
		return
	}
	val := userInfo["email"]
	if !strings.EqualFold(fmt.Sprint(val), requestBody.USER_EMAIL) {
		err := errors.New(globals.USER_NOT_FOUND)
		ErrorResponseForrequest(responseWriter, request, http.StatusBadRequest, err.Error(), transactionID)
		return
	}

	resultMap := enrichUserInfoResponse(userInfo)
	SuccessResponse(transactionID, responseWriter, resultMap)
}

// FetchRequestBodyForUserLogin ...
func FetchRequestBodyForUserLogin(request *http.Request) (structs.USER_LOGIN_REQUEST, error) {
	requestBodyStruct := structs.USER_LOGIN_REQUEST{}
	requestBodyInfo, err := io.ReadAll(request.Body)
	if err != nil {
		return requestBodyStruct, err
	}
	err = json.Unmarshal(requestBodyInfo, &requestBodyStruct)
	if err != nil {
		return requestBodyStruct, err
	}
	if len(requestBodyStruct.USER_EMAIL) == 0 || !IsvalidEmail(requestBodyStruct.USER_EMAIL) {
		return requestBodyStruct, errors.New("please provide valid email")
	}
	if len(requestBodyStruct.PASSWORD) == 0 {
		return requestBodyStruct, errors.New("please provide password")
	}
	requestBodyStruct.PASSWORD = fmt.Sprintf("%x", md5.Sum([]byte(requestBodyStruct.PASSWORD)))
	return requestBodyStruct, nil

}

// enrichUserInfoResponse ...
func enrichUserInfoResponse(userInfo map[string]interface{}) map[string]interface{} {
	resultMap := make(map[string]interface{})
	resultMap["token"] = uuid.New().String()
	resultMap["user"] = userInfo
	return resultMap
}
