package helpers

import "strings"

const (
	Err404 = "status_code=404"
)

func Is404Error(err error) bool {
	return strings.Contains(err.Error(), Err404)
}
