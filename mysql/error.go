package mysql

import (
	"github.com/arnehormann/mysql"
)

func MySQLErrorCode(err error) uint16 {
	if val, ok := err.(*mysql.MySQLError); ok {
		return val.Number
	}

	return 0 // not a mysql error
}

// MySQL error codes
const (
	ER_SPECIFIC_ACCESS_DENIED_ERROR = 1227
)