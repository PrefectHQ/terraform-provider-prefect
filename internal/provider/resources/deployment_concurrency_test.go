package resources // nolint:testpackage // need access to private concurrencyUpdateValues function

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
)

func TestConcurrencyUpdateValues(t *testing.T) {
	t.Parallel()

	limitID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	tests := []struct {
		name              string
		config            DeploymentResourceModel
		plan              DeploymentResourceModel
		prior             DeploymentResourceModel
		wantLimit         string
		wantOptions       string
		wantGlobalLimitID string
	}{
		{
			name: "concurrency limit set",
			config: DeploymentResourceModel{
				ConcurrencyLimit:         types.Int64Value(2),
				GlobalConcurrencyLimitID: customtypes.NewUUIDNull(),
			},
			plan: DeploymentResourceModel{
				ConcurrencyLimit:         types.Int64Value(2),
				GlobalConcurrencyLimitID: customtypes.NewUUIDUnknown(),
			},
			wantLimit:         "2",
			wantOptions:       "null",
			wantGlobalLimitID: "<omitted>",
		},
		{
			name: "concurrency limit unknown",
			config: DeploymentResourceModel{
				ConcurrencyLimit:         types.Int64Unknown(),
				GlobalConcurrencyLimitID: customtypes.NewUUIDNull(),
			},
			plan: DeploymentResourceModel{
				ConcurrencyLimit:         types.Int64Unknown(),
				GlobalConcurrencyLimitID: customtypes.NewUUIDUnknown(),
			},
			wantLimit:         "<omitted>",
			wantOptions:       "null",
			wantGlobalLimitID: "<omitted>",
		},
		{
			name: "global concurrency limit ID set",
			config: DeploymentResourceModel{
				ConcurrencyLimit:         types.Int64Null(),
				GlobalConcurrencyLimitID: customtypes.NewUUIDValue(limitID),
			},
			plan: DeploymentResourceModel{
				ConcurrencyLimit:         types.Int64Null(),
				GlobalConcurrencyLimitID: customtypes.NewUUIDValue(limitID),
			},
			wantLimit:         "<omitted>",
			wantOptions:       "null",
			wantGlobalLimitID: `"11111111-1111-1111-1111-111111111111"`,
		},
		{
			name: "global concurrency limit ID unknown",
			config: DeploymentResourceModel{
				ConcurrencyLimit:         types.Int64Null(),
				GlobalConcurrencyLimitID: customtypes.NewUUIDUnknown(),
			},
			plan: DeploymentResourceModel{
				ConcurrencyLimit:         types.Int64Null(),
				GlobalConcurrencyLimitID: customtypes.NewUUIDUnknown(),
			},
			prior: DeploymentResourceModel{
				ConcurrencyLimit:         types.Int64Null(),
				GlobalConcurrencyLimitID: customtypes.NewUUIDValue(limitID),
			},
			wantLimit:         "<omitted>",
			wantOptions:       "null",
			wantGlobalLimitID: "<omitted>",
		},
		{
			name: "nothing configured with no prior limit",
			config: DeploymentResourceModel{
				ConcurrencyLimit:         types.Int64Null(),
				GlobalConcurrencyLimitID: customtypes.NewUUIDNull(),
			},
			plan: DeploymentResourceModel{
				ConcurrencyLimit:         types.Int64Null(),
				GlobalConcurrencyLimitID: customtypes.NewUUIDUnknown(),
			},
			prior: DeploymentResourceModel{
				ConcurrencyLimit:         types.Int64Null(),
				GlobalConcurrencyLimitID: customtypes.NewUUIDNull(),
			},
			wantLimit:         "null",
			wantOptions:       "null",
			wantGlobalLimitID: "<omitted>",
		},
		{
			name: "nothing configured with prior concurrency limit",
			config: DeploymentResourceModel{
				ConcurrencyLimit:         types.Int64Null(),
				GlobalConcurrencyLimitID: customtypes.NewUUIDNull(),
			},
			plan: DeploymentResourceModel{
				ConcurrencyLimit:         types.Int64Null(),
				GlobalConcurrencyLimitID: customtypes.NewUUIDUnknown(),
			},
			prior: DeploymentResourceModel{
				ConcurrencyLimit:         types.Int64Value(2),
				GlobalConcurrencyLimitID: customtypes.NewUUIDNull(),
			},
			wantLimit:         "null",
			wantOptions:       "null",
			wantGlobalLimitID: "<omitted>",
		},
		{
			name: "global concurrency limit ID removed",
			config: DeploymentResourceModel{
				ConcurrencyLimit:         types.Int64Null(),
				GlobalConcurrencyLimitID: customtypes.NewUUIDNull(),
			},
			plan: DeploymentResourceModel{
				ConcurrencyLimit:         types.Int64Null(),
				GlobalConcurrencyLimitID: customtypes.NewUUIDUnknown(),
			},
			prior: DeploymentResourceModel{
				ConcurrencyLimit:         types.Int64Null(),
				GlobalConcurrencyLimitID: customtypes.NewUUIDValue(limitID),
			},
			wantLimit:         "<omitted>",
			wantOptions:       "null",
			wantGlobalLimitID: "null",
		},
		{
			name: "concurrency options set",
			config: DeploymentResourceModel{
				ConcurrencyLimit:         types.Int64Null(),
				GlobalConcurrencyLimitID: customtypes.NewUUIDNull(),
			},
			plan: DeploymentResourceModel{
				ConcurrencyLimit:         types.Int64Null(),
				GlobalConcurrencyLimitID: customtypes.NewUUIDUnknown(),
				ConcurrencyOptions: &ConcurrencyOptions{
					CollisionStrategy: types.StringValue("ENQUEUE"),
				},
			},
			wantLimit:         "null",
			wantOptions:       `{"collision_strategy":"ENQUEUE"}`,
			wantGlobalLimitID: "<omitted>",
		},
		{
			name: "concurrency options removed with no prior options",
			config: DeploymentResourceModel{
				ConcurrencyLimit:         types.Int64Null(),
				GlobalConcurrencyLimitID: customtypes.NewUUIDNull(),
			},
			plan: DeploymentResourceModel{
				ConcurrencyLimit:         types.Int64Null(),
				GlobalConcurrencyLimitID: customtypes.NewUUIDUnknown(),
			},
			wantLimit:         "null",
			wantOptions:       "null",
			wantGlobalLimitID: "<omitted>",
		},
		{
			name: "concurrency options removed with prior options",
			config: DeploymentResourceModel{
				ConcurrencyLimit:         types.Int64Null(),
				GlobalConcurrencyLimitID: customtypes.NewUUIDNull(),
			},
			plan: DeploymentResourceModel{
				ConcurrencyLimit:         types.Int64Null(),
				GlobalConcurrencyLimitID: customtypes.NewUUIDUnknown(),
			},
			prior: DeploymentResourceModel{
				ConcurrencyLimit:         types.Int64Null(),
				GlobalConcurrencyLimitID: customtypes.NewUUIDNull(),
				ConcurrencyOptions: &ConcurrencyOptions{
					CollisionStrategy: types.StringValue("ENQUEUE"),
				},
			},
			wantLimit:         "null",
			wantOptions:       "null",
			wantGlobalLimitID: "<omitted>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := concurrencyUpdateValues(tt.config, tt.plan, tt.prior)

			if got := rawMessageString(got.concurrencyLimit); got != tt.wantLimit {
				t.Errorf("concurrency_limit = %q, want %q", got, tt.wantLimit)
			}
			if got := rawMessageString(got.concurrencyOptions); got != tt.wantOptions {
				t.Errorf("concurrency_options = %q, want %q", got, tt.wantOptions)
			}
			if got := rawMessageString(got.globalConcurrencyLimitID); got != tt.wantGlobalLimitID {
				t.Errorf("global_concurrency_limit_id = %q, want %q", got, tt.wantGlobalLimitID)
			}

			if got.concurrencyLimit != nil && got.globalConcurrencyLimitID != nil {
				t.Fatal("concurrency_limit and global_concurrency_limit_id must not both be sent")
			}
		})
	}
}

func rawMessageString(value json.RawMessage) string {
	if value == nil {
		return "<omitted>"
	}

	return string(value)
}
