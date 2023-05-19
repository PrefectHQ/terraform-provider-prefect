package prefect

import (
	"context"
	hc "github.com/prefecthq/terraform-provider-prefect/prefect_api"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceBlock() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBlockCreate,
		ReadContext:   resourceBlockRead,
		UpdateContext: resourceBlockUpdate,
		DeleteContext: resourceBlockDelete,
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
			"block_schema_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"block_type_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"is_anonymous": &schema.Schema{
				Type:      schema.TypeBool,
				Optional:  true,
				Sensitive: true,
			},
			"data": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceBlockRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	c, err := getClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	blockDocumentId := d.Id()
	workspaceId := d.Get("workspace_id").(string)

	blockDocument, err := c.GetBlockDocument(ctx, blockDocumentId, workspaceId)
	if err != nil {
		return diag.FromErr(err)
	}

	// Might need to rewrite the checks below

	if err := d.Set("name", blockDocument.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("block_schema_id", blockDocument.BlockSchemaId); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("block_type_id", blockDocument.BlockTypeId); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("is_anonymous", blockDocument.IsAnonymous); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("data", blockDocument.Data); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceBlockCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*hc.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	workspaceId := d.Get("workspace_id").(string)

	blockDocument := hc.BlockDocument{
		Name:          d.Get("name").(string),
		BlockSchemaId: d.Get("block_schema_id").(string),
		BlockTypeId:   d.Get("block_type_id").(string),
		IsAnonymous:   d.Get("is_anonymous").(bool),
		Data:          d.Get("data"),
	}

	o, err := c.CreateBlockDocument(ctx, blockDocument, workspaceId)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(o.Id)

	resourceBlockRead(ctx, d, m)

	return diags
}

func resourceBlockUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*hc.Client)

	blockDocumentId := d.Id()
	workspaceId := d.Get("workspace_id").(string)

	blockDocument := hc.BlockDocument{
		BlockSchemaId: d.Get("block_schema_id").(string),
		IsAnonymous:   d.Get("is_anonymous").(bool),
		Data:          d.Get("data"),
	}

	_, err := c.UpdateBlockDocument(ctx, blockDocument, blockDocumentId, workspaceId)

	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("last_updated", time.Now().Format(time.RFC850))

	return resourceBlockRead(ctx, d, m)
}

func resourceBlockDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*hc.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	blockDocumentId := d.Id()
	workspaceId := d.Get("workspace_id").(string)

	err := c.DeleteBlockDocument(ctx, blockDocumentId, workspaceId)
	if err != nil {
		return diag.FromErr(err)
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}
