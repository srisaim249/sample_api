package structs

type CONFIG_OBJECT struct {
	SQL_CONNECTION SQL_CONNECTION_DETAILS `validate:"required"  json:"SQL_CONNECTION"`
	APP_CONFIG     APP_CONFIG             `validate:"required" json:"APP_CONFIG"`
	QUERY_CONFIG   QUERY_CONFIG           `validate:"required" json:"QUERY_CONFIG"`
}

// QUERY_CONFIG ... list of pgsql queries
type QUERY_CONFIG struct {
	LOGIN_UESR_QUERY    QUERY_STRUCT `validate:"required" json:"LOGIN_UESR_QUERY"`
	FETCH_USER_BY_EMAIL QUERY_STRUCT `validate:"required" json:"FETCH_USER_BY_EMAIL"`
	CREATE_USER         QUERY_STRUCT `validate:"required" json:"CREATE_USER"`
}

// QUERY_STRUCT ...
type QUERY_STRUCT struct {
	QUERY        string `json:"QUERY"`
	TIME_OUT_SEC int    `json:"TIME_OUT_SEC"`
}

// SQL_CONNECTION_DETAILS ...
type SQL_CONNECTION_DETAILS struct {
	SQL_HOST        string `validate:"required" json:"SQL_HOST"`
	SQL_USERNAME    string `validate:"required" json:"SQL_USERNAME"`
	SQL_PASSWORD    string `validate:"required" json:"SQL_PASSWORD"`
	SQL_PORT        string `validate:"required" json:"SQL_PORT"`
	SQL_DATABASE    string `validate:"required" json:"SQL_DATABASE"`
	SQL_TABLE       string `validate:"required" json:"SQL_TABLE"`
	SQL_DRIVER_NAME string `validate:"required" json:"SQL_DRIVER_NAME"`

	TLS_ENABLED      bool   `json:"TLS_ENABLED,omitempty"`
	CA_CERT_PATH     string `json:"ca_file"`
	CLIENT_CERT_PATH string `json:"cert_file"`
	CLIENT_KEY_PATH  string `json:"key_file"`

	SQL_SSL_MODE    string `validate:"required" json:"SQL_SSL_MODE"`
	SQL_RETRY_COUNT int    `validate:"required" json:"SQL_RETRY_COUNT"`
	SQL_DELAY_TIME  int    `validate:"required" json:"SQL_DELAY_TIME"`
}

type APP_CONFIG struct {
	PORT string `validate:"required" json:"PORT"`

	APP_NAME      string `validate:"required" json:"APP_NAME"`
	LOG_PATH      string `validate:"required" json:"LOG_PATH"`
	LOG_FILE_NAME string `validate:"required" json:"LOG_FILE_NAME"`
	LOG_LEVEL     string `validate:"required" json:"LOG_LEVEL"`
}

// USER_REGISTRATION_REQUEST ...
type USER_REGISTRATION_REQUEST struct {
	USER_NAME  string `json:"username"`
	USER_EMAIL string `json:"email"`
	FIRST_NAME string `json:"firstname"`
	LAST_NAME  string `json:"lastname"`
	PASSWORD   string `json:"password"`
}

// USER_LOGIN_REQUEST ...
type USER_LOGIN_REQUEST struct {
	USER_EMAIL string `json:"email"`
	PASSWORD   string `json:"password"`
}

// ERROR_RESPONSE_STRCUT ...
type ERROR_RESPONSE_STRCUT struct {
	TransactionID    string `json:"TransactionID"`
	ErrorDescription string `json:"message"`
}
