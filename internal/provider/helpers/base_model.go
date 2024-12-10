package helpers

import "github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"

// BaseModel is embedded in all other types and defines fields
// common to all Prefect data models.
type BaseModel struct {
	ID      customtypes.UUIDValue      `tfsdk:"id"`
	Created customtypes.TimestampValue `tfsdk:"created"`
	Updated customtypes.TimestampValue `tfsdk:"updated"`
}
