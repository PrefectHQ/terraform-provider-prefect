package prefect

import (
	"context"
	hc "github.com/prefecthq/terraform-provider-prefect/prefect_api"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceWorkQueue() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWorkQueueCreate,
		ReadContext:   resourceWorkQueueRead,
		UpdateContext: resourceWorkQueueUpdate,
		DeleteContext: resourceWorkQueueDelete,
		Schema: map[string]*schema.Schema{
			"workspace_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"last_updated": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"is_paused": &schema.Schema{
				Type:     schema.TypeBool,
				Required: true,
			},
			"concurrency_limit": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"priority": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
		},
	}
}

func resourceWorkQueueRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*hc.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	workQueueId := d.Id()
	workspaceId := d.Get("workspace_id").(string)

	workQueue, err := c.GetWorkQueue(ctx, workQueueId, workspaceId)
	if err != nil {
		return diag.FromErr(err)
	}

	// Might need to rewrite the checks below

	if err := d.Set("name", workQueue.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("description", workQueue.Description); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("is_paused", workQueue.IsPaused); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("concurrency_limit", workQueue.ConcurrencyLimit); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("priority", workQueue.Priority); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceWorkQueueCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*hc.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	workspaceId := d.Get("workspace_id").(string)

	wqq := hc.WorkQueue{
		Name:             d.Get("name").(string),
		Description:      d.Get("description").(string),
		IsPaused:         d.Get("is_paused").(bool),
		ConcurrencyLimit: d.Get("concurrency_limit").(int),
		Priority:         d.Get("priority").(int),
	}

	o, err := c.CreateWorkQueue(ctx, wqq, workspaceId)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(o.Id)

	resourceWorkQueueRead(ctx, d, m)

	return diags
}

func resourceWorkQueueUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*hc.Client)

	workQueueId := d.Id()
	workspaceId := d.Get("workspace_id").(string)

	wqq := hc.WorkQueue{
		Name:             d.Get("name").(string),
		Description:      d.Get("description").(string),
		IsPaused:         d.Get("is_paused").(bool),
		ConcurrencyLimit: d.Get("concurrency_limit").(int),
		Priority:         d.Get("priority").(int),
	}

	_, err := c.UpdateWorkQueue(ctx, wqq, workQueueId, workspaceId)

	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("last_updated", time.Now().Format(time.RFC850))

	return resourceWorkQueueRead(ctx, d, m)
}

func resourceWorkQueueDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	c, err := getClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	workQueueId := d.Id()
	workspaceId := d.Get("workspace_id").(string)

	err = c.DeleteWorkQueue(ctx, workQueueId, workspaceId)
	if err != nil {
		return diag.FromErr(err)
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}
