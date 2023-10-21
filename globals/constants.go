package globals

import (
	"database/sql"
	"sample_api/structs"
)

var (
	USER_DB_CONN *sql.DB
	CONFIG_OBJ   structs.CONFIG_OBJECT

	LINE_FIELD         = "line"
	CONTENT_TYPE       = "application/json"
	CONTENT_TYPE_FIELD = "Content-Type"
	TRX_ID_FIELD       = "TransactionID"

	USER_EXISTS       = "User Email already exists."
	VERIFICATION_SENT = "A verification mail has been sent to your registered mail."
	USER_NOT_FOUND    = "User not found with the given email and password"

	USER_INTERRUPTED = false
)

var LOG_HEIRARCHY = map[string]int{
	"debug": 1,
	"info":  2,
	"warn":  3,
	"error": 4,
}
