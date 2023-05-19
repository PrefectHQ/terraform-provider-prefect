package prefect

import (
	"context"
	"strconv"
	"time"

	hc "github.com/prefecthq/terraform-provider-prefect/prefect_api"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceBlockSchemas() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceBlockSchemasRead,
		Schema: map[string]*schema.Schema{
			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"block_schema_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"checksum": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"block_schemas": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"block_type_id": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceBlockSchemasRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	c, err := getClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	workspaceId := d.Get("workspace_id").(string)
	blockSchemaId := d.Get("block_schema_id").(string)
	checksum := d.Get("checksum").(string)

	var blockSchemasOutput []interface{}

	if blockSchemaId != "" {
		blockSchema, err := c.GetBlockSchemaById(ctx, blockSchemaId, workspaceId)
		if err != nil {
			return diag.FromErr(err)
		}

		blockSchemas := []hc.BlockSchema{*blockSchema}
		blockSchemasOutput = tfBlockSchemaSchemaOutput(blockSchemas)
	} else if checksum != "" {
		blockSchema, err := c.GetBlockSchemaByChecksum(ctx, checksum, workspaceId)
		if err != nil {
			return diag.FromErr(err)
		}

		blockSchemas := []hc.BlockSchema{*blockSchema}
		blockSchemasOutput = tfBlockSchemaSchemaOutput(blockSchemas)
	} else {
		blockSchemas, err := c.GetAllBlockSchemas(ctx, workspaceId)
		if err != nil {
			return diag.FromErr(err)
		}
		blockSchemasOutput = tfBlockSchemaSchemaOutput(blockSchemas)
	}

	if err := d.Set("block_schemas", blockSchemasOutput); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

func tfBlockSchemaSchemaOutput(block_schemas []hc.BlockSchema) []interface{} {
	schemaOutput := make([]interface{}, len(block_schemas), len(block_schemas))

	for i, block_schema := range block_schemas {
		schema := make(map[string]interface{})

		schema["id"] = block_schema.Id
		schema["block_type_id"] = block_schema.BlockTypeId

		schemaOutput[i] = schema
	}
	return schemaOutput
}
