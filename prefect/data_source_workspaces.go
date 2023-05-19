package prefect

import (
	"context"
	"strconv"
	"time"

	hc "github.com/prefecthq/terraform-provider-prefect/prefect_api"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceWorkspaces() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceWorkspacesRead,
		Schema: map[string]*schema.Schema{
			"workspaces": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"created": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"updated": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"account_id": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"name": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"description": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"handle": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"default_workspace_role_id": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceWorkspacesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	c, err := getClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	workspaces, err := c.GetAllWorkspaces(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	workspaces_output := tfWorkspacesSchemaOutput(workspaces)

	if err := d.Set("workspaces", workspaces_output); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

func tfWorkspacesSchemaOutput(workspaces []hc.Workspace) []interface{} {
	schema_output := make([]interface{}, len(workspaces), len(workspaces))

	for i, workspace := range workspaces {
		schema := make(map[string]interface{})

		schema["id"] = workspace.Id
		schema["created"] = workspace.Created
		schema["updated"] = workspace.Updated
		schema["account_id"] = workspace.AccountId
		schema["name"] = workspace.Name
		schema["description"] = workspace.Description
		schema["handle"] = workspace.Handle
		schema["default_workspace_role_id"] = workspace.DefaultWorkspaceRoleId

		schema_output[i] = schema
	}
	return schema_output
}
