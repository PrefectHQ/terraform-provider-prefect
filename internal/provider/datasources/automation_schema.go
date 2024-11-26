package datasources

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/utils"
)

// AutomationSchema returns the total schema for an Automation.
// This includes all of the root-level attributes for an Automation,
// as well as the Trigger and Actions schemas.
func AutomationSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Required:    true,
			CustomType:  customtypes.UUIDType{},
			Description: "Automation ID (UUID)",
		},
		"created": schema.StringAttribute{
			Computed:    true,
			CustomType:  customtypes.TimestampType{},
			Description: "Timestamp of when the resource was created (RFC3339)",
		},
		"updated": schema.StringAttribute{
			Computed:    true,
			CustomType:  customtypes.TimestampType{},
			Description: "Timestamp of when the resource was updated (RFC3339)",
		},
		"name": schema.StringAttribute{
			Description: "Name of the automation",
			Computed:    true,
		},
		"description": schema.StringAttribute{
			Description: "Description of the automation",
			Computed:    true,
		},
		"enabled": schema.BoolAttribute{
			Description: "Whether the automation is enabled",
			Computed:    true,
		},
		"account_id": schema.StringAttribute{
			Optional:    true,
			CustomType:  customtypes.UUIDType{},
			Description: "Account ID (UUID), defaults to the account set in the provider",
		},
		"workspace_id": schema.StringAttribute{
			Optional:    true,
			CustomType:  customtypes.UUIDType{},
			Description: "Workspace ID (UUID), defaults to the workspace set in the provider",
		},
		"trigger":            TriggerSchema(),
		"actions":            ActionsSchema(),
		"actions_on_trigger": ActionsSchema(),
		"actions_on_resolve": ActionsSchema(),
	}
}

// TriggerSchema returns the combined resource schema for an Automation Trigger.
// This combines Resource Triggers and Composite Triggers.
// We construct the TriggerSchema this way (and not in a single schema from the start)
// because Composite Triggers are higher-order Triggers that utilize Resource Triggers.
func TriggerSchema() schema.SingleNestedAttribute {
	// (1) We start with the Resource Trigger Schema Attributes
	combinedAttributes := ResourceTriggerSchemaAttributes()

	// (2) Here we add Composite Triggers to the schema
	combinedAttributes["compound"] = schema.SingleNestedAttribute{
		Computed:    true,
		Description: "A composite trigger that requires some number of triggers to have fired within the given time period",
		Attributes: func() map[string]schema.Attribute {
			attrs := CompositeTriggerSchemaAttributes()
			attrs["require"] = schema.DynamicAttribute{
				Computed:    true,
				Description: "How many triggers must fire ('any', 'all', or a number)",
			}

			return attrs
		}(),
	}
	combinedAttributes["sequence"] = schema.SingleNestedAttribute{
		Computed:    true,
		Description: "A composite trigger that requires triggers to fire in a specific order",
		Attributes:  CompositeTriggerSchemaAttributes(),
	}

	// (3) We return the combined Triggers schema
	return schema.SingleNestedAttribute{
		Computed:    true,
		Description: "The criteria for which events this Automation covers and how it will respond",
		Attributes:  combinedAttributes,
	}
}

// ResourceTriggerSchemaAttributes returns the attributes for a Resource Trigger.
// A Resource Trigger is an `event` or `metric` trigger.
func ResourceTriggerSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"event": schema.SingleNestedAttribute{
			Computed:    true,
			Description: "A trigger that fires based on the presence or absence of events within a given period of time",
			Attributes: map[string]schema.Attribute{
				"posture": schema.StringAttribute{
					Computed:    true,
					Description: "The posture of this trigger, either Reactive or Proactive",
					Validators: []validator.String{
						stringvalidator.OneOf("Reactive", "Proactive"),
					},
				},
				"match": schema.StringAttribute{
					Computed:    true,
					Description: "(JSON) Resource specification labels which this trigger will match. Use `jsonencode()`.",
					CustomType:  jsontypes.NormalizedType{},
				},
				"match_related": schema.StringAttribute{
					Computed:    true,
					Description: "(JSON) Resource specification labels for related resources which this trigger will match. Use `jsonencode()`.",
					CustomType:  jsontypes.NormalizedType{},
				},
				"after": schema.ListAttribute{
					Computed:    true,
					Description: "The event(s) which must first been seen to fire this trigger. If empty, then fire this trigger immediately",
					ElementType: types.StringType,
				},
				"expect": schema.ListAttribute{
					Computed:    true,
					Description: "The event(s) this trigger is expecting to see. If empty, this trigger will match any event",
					ElementType: types.StringType,
				},
				"for_each": schema.ListAttribute{
					Computed:    true,
					Description: "Evaluate the trigger separately for each distinct value of these labels on the resource",
					ElementType: types.StringType,
				},
				"threshold": schema.Int64Attribute{
					Computed:    true,
					Description: "The number of events required for this trigger to fire (Reactive) or expected (Proactive)",
				},
				"within": schema.Float64Attribute{
					Computed:    true,
					Description: "The time period in seconds over which the events must occur",
				},
			},
		},
		"metric": schema.SingleNestedAttribute{
			Computed:    true,
			Description: "A trigger that fires based on the results of a metric query",
			Attributes: map[string]schema.Attribute{
				"match": schema.StringAttribute{
					Computed:    true,
					Description: "(JSON) Resource specification labels which this trigger will match. Use `jsonencode()`.",
					CustomType:  jsontypes.NormalizedType{},
				},
				"match_related": schema.StringAttribute{
					Computed:    true,
					Description: "(JSON) Resource specification labels for related resources which this trigger will match. Use `jsonencode()`.",
					CustomType:  jsontypes.NormalizedType{},
				},
				"metric": schema.SingleNestedAttribute{
					Computed: true,
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the metric to query",
						},
						"operator": schema.StringAttribute{
							Computed:    true,
							Description: "The comparative operator used to evaluate the query result against the threshold value",
						},
						"threshold": schema.Float64Attribute{
							Computed:    true,
							Description: "The threshold value against which we'll compare the query results",
						},
						"range": schema.Float64Attribute{
							Computed:    true,
							Description: "The lookback duration (seconds) for a metric query. This duration is used to determine the time range over which the query will be executed.",
						},
						"firing_for": schema.Float64Attribute{
							Computed:    true,
							Description: "The duration (seconds) for which the metric query must breach OR resolve continuously before the state is updated and actions are triggered.",
						},
					},
				},
			},
		},
	}
}

// CompositeTriggerSchemaAttributes returns the shared attributes for a Composite Trigger.
// A Composite Trigger is a `compound` or `sequence` trigger.
func CompositeTriggerSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"triggers": schema.ListNestedAttribute{
			Computed:    true,
			Description: "The ordered list of triggers that must fire in sequence",
			NestedObject: schema.NestedAttributeObject{
				Attributes: ResourceTriggerSchemaAttributes(),
			},
		},
		"within": schema.Float64Attribute{
			Computed:    true,
			Description: "The time period in seconds over which the events must occur",
		},
	}
}

// ActionsSchema returns the schema for an Automation's Actions.
// Actions are polymorphic and can have different schemas based on the action type.
// In the resource schema here, we only make `type` required. The other attributes
// are needed based on the action type, which we'll validate in the resource layer.
func ActionsSchema() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Description: "List of actions to perform when the automation is triggered",
		Computed:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"type": schema.StringAttribute{
					Computed:    true,
					Description: "The type of action to perform",
					Validators: []validator.String{
						stringvalidator.OneOf(utils.AllAutomationActionTypes...),
					},
				},
				"source": schema.StringAttribute{
					Description: "(Deployment / Work Pool / Work Queue / Automation) Whether this action applies to a specific selected resource or to a specific resource by ID - 'selected' or 'inferred'",
					Computed:    true,
					Validators: []validator.String{
						stringvalidator.OneOf("selected", "inferred"),
					},
				},
				"deployment_id": schema.StringAttribute{
					Description: "(Deployment) ID of the deployment to apply this action to",
					Computed:    true,
					CustomType:  customtypes.UUIDType{},
				},
				"parameters": schema.StringAttribute{
					Description: "(Deployment) (JSON) Parameters to pass to the deployment. Use `jsonencode()`.",
					Computed:    true,
					CustomType:  jsontypes.NormalizedType{},
				},
				"job_variables": schema.StringAttribute{
					Description: "(Deployment) (JSON) Job variables to pass to the created flow run. Use `jsonencode()`.",
					Computed:    true,
					CustomType:  jsontypes.NormalizedType{},
				},
				"name": schema.StringAttribute{
					Description: "(Flow Run State Change) Name of the state to change the flow run to",
					Computed:    true,
				},
				"state": schema.StringAttribute{
					Description: "(Flow Run State Change) Type of state to change the flow run to",
					Computed:    true,
				},
				"message": schema.StringAttribute{
					Description: "(Flow Run State Change) Message to associate with the state change",
					Computed:    true,
				},
				"work_queue_id": schema.StringAttribute{
					Description: "(Work Queue) ID of the work queue to apply this action to",
					Computed:    true,
					CustomType:  customtypes.UUIDType{},
				},
				"block_document_id": schema.StringAttribute{
					Description: "(Webhook / Notification) ID of the block to use",
					Computed:    true,
					CustomType:  customtypes.UUIDType{},
				},
				"subject": schema.StringAttribute{
					Description: "(Notification) Subject of the notification",
					Computed:    true,
				},
				"body": schema.StringAttribute{
					Description: "(Notification) Body of the notification",
					Computed:    true,
				},
				"payload": schema.StringAttribute{
					Description: "(Webhook) Payload to send when calling the webhook",
					Computed:    true,
				},
				"automation_id": schema.StringAttribute{
					Description: "(Automation) ID of the automation to apply this action to",
					Computed:    true,
					CustomType:  customtypes.UUIDType{},
				},
				"work_pool_id": schema.StringAttribute{
					Description: "(Work Pool) ID of the work pool to apply this action to",
					Computed:    true,
					CustomType:  customtypes.UUIDType{},
				},
			},
		},
	}
}
