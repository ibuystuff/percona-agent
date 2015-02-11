package mysql

import (
	"fmt"
	"net"

	"github.com/go-sql-driver/mysql"
)

func MySQLErrorCode(err error) uint16 {
	if val, ok := err.(*mysql.MySQLError); ok {
		return val.Number
	}

	return 0 // not a mysql error
}

func FormatError(err error) string {
	switch err.(type) {
	case *net.OpError:
		e := err.(*net.OpError)
		if e.Op == "dial" {
			return fmt.Sprintf("%s: %s", e.Err, e.Addr)
		}
	}
	return fmt.Sprintf("%s", err)
}

// MySQL error codes
const (
	ER_SPECIFIC_ACCESS_DENIED_ERROR = 1227
	ER_SYNTAX_ERROR                 = 1064
)
