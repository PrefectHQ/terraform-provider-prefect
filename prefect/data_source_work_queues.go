package prefect

import (
	"context"
	"strconv"
	"time"

	hc "github.com/prefecthq/terraform-provider-prefect/prefect_api"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceWorkQueues() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceWorkQueuesRead,
		Schema: map[string]*schema.Schema{
			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"work_queue_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"work_queues": &schema.Schema{
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
						"name": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"description": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"is_paused": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
						},
						"concurrency_limit": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
						},
						"priority": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
						},
						"work_pool_id": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"filter": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"last_polled": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceWorkQueuesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	c, err := getClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	workspaceId := d.Get("workspace_id").(string)
	workQueueId := d.Get("work_queue_id").(string)

	var workQueuesOutput []interface{}

	if workQueueId != "" {
		workQueue, err := c.GetWorkQueue(ctx, workQueueId, workspaceId)
		if err != nil {
			return diag.FromErr(err)
		}

		workQueues := []hc.WorkQueue{*workQueue}
		workQueuesOutput = tfWorkQueuesSchemaOutput(workQueues)
	} else {
		workQueues, err := c.GetAllWorkQueues(ctx, workspaceId)
		if err != nil {
			return diag.FromErr(err)
		}
		workQueuesOutput = tfWorkQueuesSchemaOutput(workQueues)
	}

	if err := d.Set("work_queues", workQueuesOutput); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

func tfWorkQueuesSchemaOutput(work_queues []hc.WorkQueue) []interface{} {
	schemaOutput := make([]interface{}, len(work_queues), len(work_queues))

	for i, work_queue := range work_queues {
		schema := make(map[string]interface{})

		schema["id"] = work_queue.Id
		schema["created"] = work_queue.Created
		schema["updated"] = work_queue.Updated
		schema["name"] = work_queue.Name
		schema["description"] = work_queue.Description
		schema["is_paused"] = work_queue.IsPaused
		schema["concurrency_limit"] = work_queue.ConcurrencyLimit
		schema["priority"] = work_queue.Priority
		schema["work_pool_id"] = work_queue.WorkPoolId
		schema["filter"] = work_queue.Filter
		schema["last_polled"] = work_queue.LastPolled

		schemaOutput[i] = schema
	}
	return schemaOutput
}
