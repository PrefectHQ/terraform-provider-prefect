package resources

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	"github.com/prefecthq/terraform-provider-prefect/internal/utils"
)

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
		Optional:    true,
		Description: "A composite trigger that requires some number of triggers to have fired within the given time period",
		Attributes: func() map[string]schema.Attribute {
			attrs := CompositeTriggerSchemaAttributes()
			attrs["require"] = schema.DynamicAttribute{
				Required:    true,
				Description: "How many triggers must fire ('any', 'all', or a number)",
			}

			return attrs
		}(),
	}
	combinedAttributes["sequence"] = schema.SingleNestedAttribute{
		Optional:    true,
		Description: "A composite trigger that requires triggers to fire in a specific order",
		Attributes:  CompositeTriggerSchemaAttributes(),
	}

	// (3) We return the combined Triggers schema
	return schema.SingleNestedAttribute{
		Required:    true,
		Description: "The criteria for which events this Automation covers and how it will respond",
		Attributes:  combinedAttributes,
	}
}

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
				"match": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: "(JSON) Resource specification labels which this trigger will match. Use `jsonencode()`.",
					CustomType:  jsontypes.NormalizedType{},
					Default:     stringdefault.StaticString("{}"),
				},
				"match_related": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: "(JSON) Resource specification labels for related resources which this trigger will match. Use `jsonencode()`.",
					CustomType:  jsontypes.NormalizedType{},
					Default:     stringdefault.StaticString("{}"),
				},
				"after": schema.ListAttribute{
					Optional:    true,
					Computed:    true,
					Description: "The event(s) which must first been seen to fire this trigger. If empty, then fire this trigger immediately",
					ElementType: types.StringType,
					Default:     listdefault.StaticValue(basetypes.NewListValueMust(types.StringType, []attr.Value{})),
				},
				"expect": schema.ListAttribute{
					Optional:    true,
					Computed:    true,
					Description: "The event(s) this trigger is expecting to see. If empty, this trigger will match any event",
					ElementType: types.StringType,
					Default:     listdefault.StaticValue(basetypes.NewListValueMust(types.StringType, []attr.Value{})),
				},
				"for_each": schema.ListAttribute{
					Optional:    true,
					Computed:    true,
					Description: "Evaluate the trigger separately for each distinct value of these labels on the resource",
					ElementType: types.StringType,
					Default:     listdefault.StaticValue(basetypes.NewListValueMust(types.StringType, []attr.Value{})),
				},
				"threshold": schema.Int64Attribute{
					Optional:    true,
					Description: "The number of events required for this trigger to fire (Reactive) or expected (Proactive)",
				},
				"within": schema.Float64Attribute{
					Optional:    true,
					Description: "The time period in seconds over which the events must occur",
				},
			},
		},
		"metric": schema.SingleNestedAttribute{
			Optional:    true,
			Description: "A trigger that fires based on the results of a metric query",
			Attributes: map[string]schema.Attribute{
				"match": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: "(JSON) Resource specification labels which this trigger will match. Use `jsonencode()`.",
					CustomType:  jsontypes.NormalizedType{},
					Default:     stringdefault.StaticString("{}"),
				},
				"match_related": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: "(JSON) Resource specification labels for related resources which this trigger will match. Use `jsonencode()`.",
					CustomType:  jsontypes.NormalizedType{},
					Default:     stringdefault.StaticString("{}"),
				},
				"metric": schema.SingleNestedAttribute{
					Required: true,
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "The name of the metric to query",
							Validators: []validator.String{
								stringvalidator.OneOf(utils.AllTriggerMetricNames...),
							},
						},
						"operator": schema.StringAttribute{
							Required:    true,
							Description: "The comparative operator used to evaluate the query result against the threshold value",
							Validators: []validator.String{
								stringvalidator.OneOf(utils.AllMetricOperators...),
							},
						},
						"threshold": schema.Float64Attribute{
							Required:    true,
							Description: "The threshold value against which we'll compare the query results",
						},
						"range": schema.Float64Attribute{
							Required:    true,
							Description: "The lookback duration (seconds) for a metric query. This duration is used to determine the time range over which the query will be executed.",
						},
						"firing_for": schema.Float64Attribute{
							Required:    true,
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
			Required:    true,
			Description: "The ordered list of triggers that must fire in sequence",
			NestedObject: schema.NestedAttributeObject{
				Attributes: ResourceTriggerSchemaAttributes(),
			},
		},
		"within": schema.Float64Attribute{
			Optional:    true,
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
		Optional:    true,
		Computed:    true,
		Default: listdefault.StaticValue(basetypes.NewListValueMust(
			types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"type":              types.StringType,
					"source":            types.StringType,
					"automation_id":     customtypes.UUIDType{},
					"block_document_id": customtypes.UUIDType{},
					"deployment_id":     customtypes.UUIDType{},
					"work_pool_id":      customtypes.UUIDType{},
					"work_queue_id":     customtypes.UUIDType{},
					"subject":           types.StringType,
					"body":              types.StringType,
					"payload":           types.StringType,
					"name":              types.StringType,
					"state":             types.StringType,
					"message":           types.StringType,
					"parameters":        jsontypes.NormalizedType{},
					"job_variables":     jsontypes.NormalizedType{},
				},
			},
			[]attr.Value{},
		)),
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"type": schema.StringAttribute{
					Required:    true,
					Description: "The type of action to perform",
					Validators: []validator.String{
						stringvalidator.OneOf(utils.AllAutomationActionTypes...),
					},
				},
				"source": schema.StringAttribute{
					Description: "(Deployment / Work Pool / Work Queue / Automation) Whether this action applies to a specific selected resource or to a specific resource by ID - 'selected' or 'inferred'",
					Optional:    true,
					Validators: []validator.String{
						stringvalidator.OneOf("selected", "inferred"),
					},
				},
				"deployment_id": schema.StringAttribute{
					Description: "(Deployment) ID of the deployment to apply this action to",
					Optional:    true,
					CustomType:  customtypes.UUIDType{},
				},
				"parameters": schema.StringAttribute{
					Description: "(Deployment) (JSON) Parameters to pass to the deployment. Use `jsonencode()`.",
					Optional:    true,
					Computed:    true,
					CustomType:  jsontypes.NormalizedType{},
					Default:     stringdefault.StaticString("{}"),
				},
				"job_variables": schema.StringAttribute{
					Description: "(Deployment) (JSON) Job variables to pass to the created flow run. Use `jsonencode()`.",
					Optional:    true,
					Computed:    true,
					CustomType:  jsontypes.NormalizedType{},
					Default:     stringdefault.StaticString("{}"),
				},
				"name": schema.StringAttribute{
					Description: "(Flow Run State Change) Name of the state to change the flow run to",
					Optional:    true,
				},
				"state": schema.StringAttribute{
					Description: "(Flow Run State Change) Type of state to change the flow run to",
					Optional:    true,
				},
				"message": schema.StringAttribute{
					Description: "(Flow Run State Change) Message to associate with the state change",
					Optional:    true,
				},
				"work_queue_id": schema.StringAttribute{
					Description: "(Work Queue) ID of the work queue to apply this action to",
					Optional:    true,
					CustomType:  customtypes.UUIDType{},
				},
				"block_document_id": schema.StringAttribute{
					Description: "(Webhook / Notification) ID of the block to use",
					Optional:    true,
					CustomType:  customtypes.UUIDType{},
				},
				"subject": schema.StringAttribute{
					Description: "(Notification) Subject of the notification",
					Optional:    true,
				},
				"body": schema.StringAttribute{
					Description: "(Notification) Body of the notification",
					Optional:    true,
				},
				"payload": schema.StringAttribute{
					Description: "(Webhook) Payload to send when calling the webhook",
					Optional:    true,
				},
				"automation_id": schema.StringAttribute{
					Description: "(Automation) ID of the automation to apply this action to",
					Optional:    true,
					CustomType:  customtypes.UUIDType{},
				},
				"work_pool_id": schema.StringAttribute{
					Description: "(Work Pool) ID of the work pool to apply this action to",
					Optional:    true,
					CustomType:  customtypes.UUIDType{},
				},
			},
		},
	}
}
