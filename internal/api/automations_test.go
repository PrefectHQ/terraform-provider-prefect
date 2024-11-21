package api_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
)

func TestResourceSpecificationModel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		json string
		want api.ResourceSpecification
	}{
		{
			name: "Single String",
			json: `{"prefect.resource.id":"flow-123"}`,
			want: api.ResourceSpecification{"prefect.resource.id": api.StringOrSlice{String: "flow-123", IsList: false}},
		},
		{
			name: "String List",
			json: `{"prefect.resource.id":["flow-123","flow-456"]}`,
			want: api.ResourceSpecification{"prefect.resource.id": api.StringOrSlice{StringList: []string{"flow-123", "flow-456"}, IsList: true}},
		},
		{
			name: "Empty Dict",
			json: `{}`,
			want: api.ResourceSpecification{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got api.ResourceSpecification
			err := json.Unmarshal([]byte(tt.json), &got)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)

			bytes, err := json.Marshal(got)
			assert.NoError(t, err)
			assert.Equal(t, tt.json, string(bytes))
		})
	}
}

func TestActionModel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		json string
		want api.Action
	}{
		{
			name: "DoNothing",
			json: `{"type": "do-nothing"}`,
			want: api.Action{Type: "do-nothing"},
		},
		{
			name: "RunDeployment",
			json: `{
							"type": "run-deployment",
							"source": "selected",
							"deployment_id": "123e4567-e89b-12d3-a456-426614174000",
							"parameters": {"foo": "bar"},
							"job_variables": {"env": "prod"}
					}`,
			want: api.Action{
				Type:         "run-deployment",
				Source:       ptr.To("selected"),
				DeploymentID: ptr.To(uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")),
				Parameters:   map[string]interface{}{"foo": "bar"},
				JobVariables: map[string]interface{}{"env": "prod"},
			},
		},
		{
			name: "PauseDeployment",
			json: `{
							"type": "pause-deployment",
							"source": "selected",
							"deployment_id": "123e4567-e89b-12d3-a456-426614174000"
					}`,
			want: api.Action{
				Type:         "pause-deployment",
				Source:       ptr.To("selected"),
				DeploymentID: ptr.To(uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")),
			},
		},
		{
			name: "ResumeDeployment",
			json: `{
							"type": "resume-deployment",
							"source": "selected",
							"deployment_id": "123e4567-e89b-12d3-a456-426614174000"
					}`,
			want: api.Action{
				Type:         "resume-deployment",
				Source:       ptr.To("selected"),
				DeploymentID: ptr.To(uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")),
			},
		},
		{
			name: "CancelFlowRun",
			json: `{"type": "cancel-flow-run"}`,
			want: api.Action{Type: "cancel-flow-run"},
		},
		{
			name: "ChangeFlowRunState",
			json: `{
							"type": "change-flow-run-state",
							"name": "Failed",
							"state": "FAILED",
							"message": "Flow run failed"
					}`,
			want: api.Action{
				Type:    "change-flow-run-state",
				Name:    ptr.To("Failed"),
				State:   ptr.To("FAILED"),
				Message: ptr.To("Flow run failed"),
			},
		},
		{
			name: "PauseWorkQueue",
			json: `{
							"type": "pause-work-queue",
							"source": "selected",
							"work_queue_id": "123e4567-e89b-12d3-a456-426614174000"
					}`,
			want: api.Action{
				Type:        "pause-work-queue",
				Source:      ptr.To("selected"),
				WorkQueueID: ptr.To(uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")),
			},
		},
		{
			name: "ResumeWorkQueue",
			json: `{
							"type": "resume-work-queue",
							"source": "selected",
							"work_queue_id": "123e4567-e89b-12d3-a456-426614174000"
					}`,
			want: api.Action{
				Type:        "resume-work-queue",
				Source:      ptr.To("selected"),
				WorkQueueID: ptr.To(uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")),
			},
		},
		{
			name: "SendNotification",
			json: `{
							"type": "send-notification",
							"block_document_id": "123e4567-e89b-12d3-a456-426614174000",
							"subject": "Alert",
							"body": "Something happened"
					}`,
			want: api.Action{
				Type:            "send-notification",
				BlockDocumentID: ptr.To(uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")),
				Subject:         ptr.To("Alert"),
				Body:            ptr.To("Something happened"),
			},
		},
		{
			name: "CallWebhook",
			json: `{
							"type": "call-webhook",
							"block_document_id": "123e4567-e89b-12d3-a456-426614174000",
							"payload": "{\"message\": \"test\"}"
					}`,
			want: api.Action{
				Type:            "call-webhook",
				BlockDocumentID: ptr.To(uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")),
				Payload:         ptr.To("{\"message\": \"test\"}"),
			},
		},
		{
			name: "PauseAutomation",
			json: `{
							"type": "pause-automation",
							"source": "selected",
							"automation_id": "123e4567-e89b-12d3-a456-426614174000"
					}`,
			want: api.Action{
				Type:         "pause-automation",
				Source:       ptr.To("selected"),
				AutomationID: ptr.To(uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")),
			},
		},
		{
			name: "ResumeAutomation",
			json: `{
							"type": "resume-automation",
							"source": "selected",
							"automation_id": "123e4567-e89b-12d3-a456-426614174000"
					}`,
			want: api.Action{
				Type:         "resume-automation",
				Source:       ptr.To("selected"),
				AutomationID: ptr.To(uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")),
			},
		},
		{
			name: "SuspendFlowRun",
			json: `{"type": "suspend-flow-run"}`,
			want: api.Action{Type: "suspend-flow-run"},
		},
		{
			name: "ResumeFlowRun",
			json: `{"type": "resume-flow-run"}`,
			want: api.Action{Type: "resume-flow-run"},
		},
		{
			name: "DeclareIncident",
			json: `{"type": "declare-incident"}`,
			want: api.Action{Type: "declare-incident"},
		},
		{
			name: "PauseWorkPool",
			json: `{
							"type": "pause-work-pool",
							"source": "selected",
							"work_pool_id": "123e4567-e89b-12d3-a456-426614174000"
					}`,
			want: api.Action{
				Type:       "pause-work-pool",
				Source:     ptr.To("selected"),
				WorkPoolID: ptr.To(uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")),
			},
		},
		{
			name: "ResumeWorkPool",
			json: `{
							"type": "resume-work-pool",
							"source": "selected",
							"work_pool_id": "123e4567-e89b-12d3-a456-426614174000"
					}`,
			want: api.Action{
				Type:       "resume-work-pool",
				Source:     ptr.To("selected"),
				WorkPoolID: ptr.To(uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got api.Action
			err := json.Unmarshal([]byte(tt.json), &got)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got, "Expected %+v but got %+v", tt.want, got)
		})
	}
}
