package api

import (
	"context"

	"github.com/google/uuid"
)

// VariablesClient is a client for working with variables.
type VariablesClient interface {
	Create(ctx context.Context, variable VariableCreate) (*Variable, error)
	Get(ctx context.Context, variableID uuid.UUID) (*Variable, error)
	GetByName(ctx context.Context, name string) (*Variable, error)
	List(ctx context.Context, filter VariableFilter) ([]Variable, error)
	Update(ctx context.Context, variableID uuid.UUID, variable VariableUpdate) error
	Delete(ctx context.Context, variableID uuid.UUID) error
}

// Variable is a representation of a variable.
type Variable struct {
	BaseModel
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
	Tags  []string    `json:"tags"`
}

// VariableCreate is a subset of Variable used when creating variables.
type VariableCreate struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
	Tags  []string    `json:"tags"`
}

// VariableUpdate is a subset of Variable used when updating variables.
type VariableUpdate struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
	Tags  []string    `json:"tags"`
}

// VariableFilterSettings defines settings when searching for variables.
type VariableFilterSettings struct {
	Limit     *int64          `json:"limit"`
	Offset    *int64          `json:"offset"`
	Variables *VariableFilter `json:"variables"`
	Sort      string          `json:"sort"`
}

// VariableFilter defines filters when searching for variables.
type VariableFilter struct {
	ID    *VariableFilterID    `json:"id"`
	Name  *VariableFilterName  `json:"name"`
	Value *VariableFilterValue `json:"value"`
	Tags  *VariableFilterTags  `json:"tags"`
}

// VariableFilterID defines filter criteria searching on variable IDs.
type VariableFilterID struct {
	Any string `json:"any_"`
}

// VariableFilterName defines filter criteria searching on variable names.
type VariableFilterName struct {
	Any  string `json:"any_"`
	Like string `json:"like_"`
}

// VariableFilterValue defines filter criteria searching on variable values.
type VariableFilterValue struct {
	Any  string `json:"any_"`
	Like string `json:"like_"`
}

// VariableFilterTags defines filter criteria searching on variable IDs.
type VariableFilterTags struct {
	All  string `json:"all_"`
	Null bool   `json:"is_null_"`
}
