package utils

import (
	"time"

	"github.com/avast/retry-go/v4"
)

const RetryAttempts uint = 5
const RetryDelay time.Duration = 500 * time.Millisecond

var DefaultRetryOptions = []retry.Option{
	retry.Attempts(RetryAttempts),
	retry.Delay(RetryDelay),
}
