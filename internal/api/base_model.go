package api

import (
	"time"

	"github.com/google/uuid"
)

// BaseModel is embedded in all other types and defines fields
// common to all Prefect data models.
type BaseModel struct {
	ID      uuid.UUID  `json:"id"`
	Created *time.Time `json:"created"`
	Updated *time.Time `json:"updated"`
}
