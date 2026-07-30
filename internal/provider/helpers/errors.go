package helpers

import (
	"errors"
	"strings"
)

const (
	Err404 = "status_code=404"
)

// ErrNotFound is a sentinel for client responses that semantically mean
// "the remote object does not exist" but did not arrive as an HTTP 404
// (e.g. a 200 OK list filter that returns no matching record). Wrap it so
// resource Read methods can treat the object as deleted and drop it from state.
var ErrNotFound = errors.New("resource not found")

func Is404Error(err error) bool {
	return strings.Contains(err.Error(), Err404) || errors.Is(err, ErrNotFound)
}
