package api_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/stretchr/testify/assert"
)

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
				Source:       new("selected"),
				DeploymentID: new(uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")),
				Parameters:   map[string]any{"foo": "bar"},
				JobVariables: map[string]any{"env": "prod"},
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
				Source:       new("selected"),
				DeploymentID: new(uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")),
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
				Source:       new("selected"),
				DeploymentID: new(uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")),
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
				Name:    new("Failed"),
				State:   new("FAILED"),
				Message: new("Flow run failed"),
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
				Source:      new("selected"),
				WorkQueueID: new(uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")),
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
				Source:      new("selected"),
				WorkQueueID: new(uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")),
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
				BlockDocumentID: new(uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")),
				Subject:         new("Alert"),
				Body:            new("Something happened"),
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
				BlockDocumentID: new(uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")),
				Payload:         new("{\"message\": \"test\"}"),
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
				Source:       new("selected"),
				AutomationID: new(uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")),
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
				Source:       new("selected"),
				AutomationID: new(uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")),
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
				Source:     new("selected"),
				WorkPoolID: new(uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")),
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
				Source:     new("selected"),
				WorkPoolID: new(uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")),
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
