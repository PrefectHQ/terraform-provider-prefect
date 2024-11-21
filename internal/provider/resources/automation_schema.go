package resources

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/customtypes"
)

// ResourceTriggerSchemaAttributes returns the attributes for a Resource Trigger.
// A Resource Trigger is an `event` or `metric` trigger.
func ResourceTriggerSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"event": schema.SingleNestedAttribute{
			Optional:    true,
			Description: "A trigger that fires based on the presence or absence of events within a given period of time",
			Attributes: map[string]schema.Attribute{
				"posture": schema.StringAttribute{
					Required:    true,
					Description: "The posture of this trigger, either Reactive or Proactive",
					Validators: []validator.String{
						stringvalidator.OneOf("Reactive", "Proactive"),
					},
				},
				"match": schema.MapAttribute{
					Optional:    true,
					Description: "Labels for resources which this trigger will match",
					ElementType: types.StringType,
				},
				"match_related": schema.MapAttribute{
					Optional:    true,
					Description: "Labels for related resources which this trigger will match",
					ElementType: types.ListType{
						ElemType: types.StringType,
					},
				},
				"after": schema.SetAttribute{
					Optional:    true,
					Description: "The event(s) which must first been seen to fire this trigger. If empty, then fire this trigger immediately",
					ElementType: types.StringType,
				},
				"expect": schema.SetAttribute{
					Optional:    true,
					Description: "The event(s) this trigger is expecting to see. If empty, this trigger will match any event",
					ElementType: types.StringType,
				},
				"for_each": schema.SetAttribute{
					Optional:    true,
					Description: "Evaluate the trigger separately for each distinct value of these labels on the resource",
					ElementType: types.StringType,
				},
				"threshold": schema.Int64Attribute{
					Optional:    true,
					Description: "The number of events required for this trigger to fire (Reactive) or expected (Proactive)",
					Validators: []validator.Int64{
						int64validator.AtLeast(1),
					},
				},
				"within": schema.Int64Attribute{
					Optional:    true,
					Description: "The time period in seconds over which the events must occur",
					Validators: []validator.Int64{
						int64validator.AtLeast(0),
					},
				},
			},
		},
		"metric": schema.SingleNestedAttribute{
			Optional:    true,
			Description: "A trigger that fires based on the results of a metric query",
			Attributes: map[string]schema.Attribute{
				"posture": schema.StringAttribute{
					Required:    true,
					Description: "The posture of this trigger (Metric)",
					Validators: []validator.String{
						stringvalidator.OneOf("Metric"),
					},
				},
				"metric": schema.SingleNestedAttribute{
					Required: true,
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "The name of the metric to query",
						},
						"operator": schema.StringAttribute{
							Required:    true,
							Description: "The operator to compare the metric value against the threshold",
						},
						"threshold": schema.Float64Attribute{
							Required:    true,
							Description: "The threshold value to compare against",
						},
						"range": schema.Int64Attribute{
							Required:    true,
							Description: "The time range in seconds over which to evaluate the metric",
						},
					},
				},
			},
		},
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
		Optional:    true,
		Description: "A composite trigger that requires some number of triggers to have fired within the given time period",
		Attributes: map[string]schema.Attribute{
			"triggers": schema.ListNestedAttribute{
				Required:    true,
				Description: "The list of triggers that make up this compound trigger",
				NestedObject: schema.NestedAttributeObject{
					Attributes: ResourceTriggerSchemaAttributes(),
				},
			},
			"require": schema.StringAttribute{
				Required:    true,
				Description: "How many triggers must fire ('any', 'all', or a number)",
			},
		},
	}
	combinedAttributes["sequence"] = schema.SingleNestedAttribute{
		Optional:    true,
		Description: "A composite trigger that requires triggers to fire in a specific order",
		Attributes: map[string]schema.Attribute{
			"triggers": schema.ListNestedAttribute{
				Required:    true,
				Description: "The ordered list of triggers that must fire in sequence",
				NestedObject: schema.NestedAttributeObject{
					Attributes: ResourceTriggerSchemaAttributes(),
				},
			},
		},
	}

	// (3) We return the combined Triggers schema
	return schema.SingleNestedAttribute{
		Required:    true,
		Description: "The criteria for which events this Automation covers and how it will respond",
		Validators: []validator.Object{
			objectvalidator.ExactlyOneOf(
				path.MatchRoot("event"),
				path.MatchRoot("metric"),
				path.MatchRoot("compound"),
				path.MatchRoot("sequence"),
			),
		},
		Attributes: combinedAttributes,
	}
}

// ActionsSchema returns the schema for an Automation's Actions.
// Actions are polymorphic and can have different schemas based on the action type.
func ActionsSchema() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Description: "List of actions to perform when the automation is triggered",
		Optional:    true,
		Computed:    true,
		Default:     listdefault.StaticValue(basetypes.NewListValueMust(types.ObjectType{}, []attr.Value{})),
		NestedObject: schema.NestedAttributeObject{
			Validators: []validator.Object{
				objectvalidator.ExactlyOneOf(
					path.MatchRoot("do-nothing"),
					path.MatchRoot("run-deployment"),
					path.MatchRoot("pause-deployment"),
					path.MatchRoot("resume-deployment"),
					path.MatchRoot("cancel-flow-run"),
					path.MatchRoot("change-flow-run-state"),
					path.MatchRoot("pause-work-queue"),
					path.MatchRoot("resume-work-queue"),
					path.MatchRoot("send-notification"),
					path.MatchRoot("call-webhook"),
					path.MatchRoot("pause-automation"),
					path.MatchRoot("resume-automation"),
					path.MatchRoot("suspend-flow-run"),
					path.MatchRoot("resume-flow-run"),
					path.MatchRoot("declare-incident"),
					path.MatchRoot("pause-work-pool"),
					path.MatchRoot("resume-work-pool"),
				),
			},
			Attributes: map[string]schema.Attribute{
				"do-nothing": schema.SingleNestedAttribute{
					Description: "Do nothing when automation is triggered",
					Optional:    true,
					Attributes:  map[string]schema.Attribute{},
				},
				"run-deployment": schema.SingleNestedAttribute{
					Description: "Run a deployment",
					Optional:    true,
					Attributes: map[string]schema.Attribute{
						"source": schema.StringAttribute{
							Description: "How to determine the deployment - 'selected' or 'inferred'",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("selected", "inferred"),
							},
						},
						"deployment_id": schema.StringAttribute{
							Description: "ID of the deployment to run",
							Optional:    true,
						},
						"parameters": schema.MapAttribute{
							Description: "Parameters to pass to the deployment",
							Optional:    true,
							ElementType: types.StringType,
						},
						"job_variables": schema.MapAttribute{
							Description: "Job variables to pass to the created flow run",
							Optional:    true,
							ElementType: types.StringType,
						},
					},
				},
				"pause-deployment": schema.SingleNestedAttribute{
					Description: "Pause a deployment",
					Optional:    true,
					Attributes: map[string]schema.Attribute{
						"source": schema.StringAttribute{
							Description: "How to determine the deployment - 'selected' or 'inferred'",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("selected", "inferred"),
							},
						},
						"deployment_id": schema.StringAttribute{
							Description: "ID of the deployment to pause",
							Optional:    true,
						},
					},
				},
				"resume-deployment": schema.SingleNestedAttribute{
					Description: "Resume a deployment",
					Optional:    true,
					Attributes: map[string]schema.Attribute{
						"source": schema.StringAttribute{
							Description: "How to determine the deployment - 'selected' or 'inferred'",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("selected", "inferred"),
							},
						},
						"deployment_id": schema.StringAttribute{
							Description: "ID of the deployment to resume",
							Optional:    true,
						},
					},
				},
				"cancel-flow-run": schema.SingleNestedAttribute{
					Description: "Cancel a flow run",
					Optional:    true,
					Attributes:  map[string]schema.Attribute{},
				},
				"change-flow-run-state": schema.SingleNestedAttribute{
					Description: "Change the state of a flow run",
					Optional:    true,
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Name of the state to change the flow run to",
							Optional:    true,
						},
						"state": schema.StringAttribute{
							Description: "Type of state to change the flow run to",
							Required:    true,
						},
						"message": schema.StringAttribute{
							Description: "Message to associate with the state change",
							Optional:    true,
						},
					},
				},
				"pause-work-queue": schema.SingleNestedAttribute{
					Description: "Pause a work queue",
					Optional:    true,
					Attributes: map[string]schema.Attribute{
						"source": schema.StringAttribute{
							Description: "How to determine the work queue - 'selected' or 'inferred'",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("selected", "inferred"),
							},
						},
						"work_queue_id": schema.StringAttribute{
							Description: "ID of the work queue to pause",
							Optional:    true,
						},
					},
				},
				"resume-work-queue": schema.SingleNestedAttribute{
					Description: "Resume a work queue",
					Optional:    true,
					Attributes: map[string]schema.Attribute{
						"source": schema.StringAttribute{
							Description: "How to determine the work queue - 'selected' or 'inferred'",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("selected", "inferred"),
							},
						},
						"work_queue_id": schema.StringAttribute{
							Description: "ID of the work queue to resume",
							Optional:    true,
						},
					},
				},
				"send-notification": schema.SingleNestedAttribute{
					Description: "Send a notification",
					Optional:    true,
					Attributes: map[string]schema.Attribute{
						"block_document_id": schema.StringAttribute{
							Description: "ID of the notification block to use",
							Required:    true,
						},
						"subject": schema.StringAttribute{
							Description: "Subject of the notification",
							Required:    true,
						},
						"body": schema.StringAttribute{
							Description: "Body of the notification",
							Required:    true,
						},
					},
				},
				"call-webhook": schema.SingleNestedAttribute{
					Description: "Call a webhook",
					Optional:    true,
					Attributes: map[string]schema.Attribute{
						"block_document_id": schema.StringAttribute{
							Description: "ID of the webhook block to use",
							Required:    true,
						},
						"payload": schema.StringAttribute{
							Description: "Payload to send when calling the webhook",
							Optional:    true,
						},
					},
				},
				"pause-automation": schema.SingleNestedAttribute{
					Description: "Pause an automation",
					Optional:    true,
					Attributes: map[string]schema.Attribute{
						"source": schema.StringAttribute{
							Description: "How to determine the automation - 'selected' or 'inferred'",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("selected", "inferred"),
							},
						},
						"automation_id": schema.StringAttribute{
							Description: "ID of the automation to pause",
							Optional:    true,
						},
					},
				},
				"resume-automation": schema.SingleNestedAttribute{
					Description: "Resume an automation",
					Optional:    true,
					Attributes: map[string]schema.Attribute{
						"source": schema.StringAttribute{
							Description: "How to determine the automation - 'selected' or 'inferred'",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("selected", "inferred"),
							},
						},
						"automation_id": schema.StringAttribute{
							Description: "ID of the automation to resume",
							Optional:    true,
						},
					},
				},
				"suspend-flow-run": schema.SingleNestedAttribute{
					Description: "Suspend a flow run",
					Optional:    true,
					Attributes:  map[string]schema.Attribute{},
				},
				"resume-flow-run": schema.SingleNestedAttribute{
					Description: "Resume a flow run",
					Optional:    true,
					Attributes:  map[string]schema.Attribute{},
				},
				"declare-incident": schema.SingleNestedAttribute{
					Description: "Declare an incident",
					Optional:    true,
					Attributes:  map[string]schema.Attribute{},
				},
				"pause-work-pool": schema.SingleNestedAttribute{
					Description: "Pause a work pool",
					Optional:    true,
					Attributes: map[string]schema.Attribute{
						"source": schema.StringAttribute{
							Description: "How to determine the work pool - 'selected' or 'inferred'",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("selected", "inferred"),
							},
						},
						"work_pool_id": schema.StringAttribute{
							Description: "ID of the work pool to pause",
							Optional:    true,
						},
					},
				},
				"resume-work-pool": schema.SingleNestedAttribute{
					Description: "Resume a work pool",
					Optional:    true,
					Attributes: map[string]schema.Attribute{
						"source": schema.StringAttribute{
							Description: "How to determine the work pool - 'selected' or 'inferred'",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("selected", "inferred"),
							},
						},
						"work_pool_id": schema.StringAttribute{
							Description: "ID of the work pool to resume",
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

// AutomationSchema returns the total schema for an Automation.
// This includes all of the root-level attributes for an Automation,
// as well as the Trigger and Actions schemas.
func AutomationSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
			// We cannot use a CustomType due to a conflict with PlanModifiers; see
			// https://github.com/hashicorp/terraform-plugin-framework/issues/763
			// https://github.com/hashicorp/terraform-plugin-framework/issues/754
			Description: "Automation ID (UUID)",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"created": schema.StringAttribute{
			Computed:    true,
			CustomType:  customtypes.TimestampType{},
			Description: "Timestamp of when the resource was created (RFC3339)",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"updated": schema.StringAttribute{
			Computed:    true,
			CustomType:  customtypes.TimestampType{},
			Description: "Timestamp of when the resource was updated (RFC3339)",
		},
		"account_id": schema.StringAttribute{
			CustomType:  customtypes.UUIDType{},
			Description: "Account ID (UUID), defaults to the account set in the provider",
			Optional:    true,
		},
		"name": schema.StringAttribute{
			Description: "Name of the automation",
			Required:    true,
		},
		"description": schema.StringAttribute{
			Description: "Description of the automation",
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(""),
		},
		"enabled": schema.BoolAttribute{
			Description: "Whether the automation is enabled",
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(true),
		},
		"trigger":            TriggerSchema(),
		"actions":            ActionsSchema(),
		"actions_on_trigger": ActionsSchema(),
		"actions_on_resolve": ActionsSchema(),
	}
}
