package prefect

import (
	"context"

	hc "terraform-provider-prefect/prefect_api"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceWorkspace() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWorkspaceCreate,
		ReadContext:   resourceWorkspaceRead,
		UpdateContext: resourceWorkspaceUpdate,
		DeleteContext: resourceWorkspaceDelete,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"handle": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceWorkspaceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*hc.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	wsp := hc.Workspace{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Handle:      d.Get("handle").(string),
	}

	o, err := c.CreateWorkspace(ctx, wsp)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(o.Id)

	resourceWorkspaceRead(ctx, d, m)

	return diags
}

func resourceWorkspaceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*hc.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	workspaceID := d.Id()

	workspace, err := c.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return diag.FromErr(err)
	}

	// Might need to rewrite the checks below

	if err := d.Set("name", workspace.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("description", workspace.Description); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("handle", workspace.Handle); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceWorkspaceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceWorkspaceRead(ctx, d, m)
}

func resourceWorkspaceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	c, err := getClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	workspaceID := d.Id()

	err = c.DeleteWorkspace(ctx, workspaceID)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
