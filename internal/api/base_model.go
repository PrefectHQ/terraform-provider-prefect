package api

import (
	"time"

	"github.com/google/uuid"
)

type BaseModel struct {
	ID      uuid.UUID  `json:"id"`
	Created *time.Time `json:"created"`
	Updated *time.Time `json:"updated"`
}
