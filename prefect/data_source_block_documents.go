package prefect

import (
	"context"
	"strconv"
	"time"

	hc "github.com/prefecthq/terraform-provider-prefect/prefect_api"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceBlockDocuments() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceBlockDocumentsRead,
		Schema: map[string]*schema.Schema{
			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"block_document_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"block_documents": &schema.Schema{
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
						"data": &schema.Schema{
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"block_schema_id": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"block_type_id": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"block_document_references": &schema.Schema{
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"is_anonymous": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceBlockDocumentsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	c, err := getClient(m)
	if err != nil {
		return diag.FromErr(err)
	}

	workspaceId := d.Get("workspace_id").(string)
	blockDocumentId := d.Get("block_document_id").(string)

	var blockDocumentsOutput []interface{}

	if blockDocumentId != "" {
		blockDocument, err := c.GetBlockDocument(ctx, blockDocumentId, workspaceId)
		if err != nil {
			return diag.FromErr(err)
		}

		blockDocuments := []hc.BlockDocument{*blockDocument}
		blockDocumentsOutput = tfBlockDocumentschemaOutput(blockDocuments)
	} else {
		blockDocuments, err := c.GetAllBlockDocuments(ctx, workspaceId)

		if err != nil {
			return diag.FromErr(err)
		}

		blockDocumentsOutput = tfBlockDocumentschemaOutput(blockDocuments)
	}

	if err := d.Set("block_documents", blockDocumentsOutput); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

func tfBlockDocumentschemaOutput(block_documents []hc.BlockDocument) []interface{} {
	schemaOutput := make([]interface{}, len(block_documents), len(block_documents))

	for i, block_document := range block_documents {
		schema := make(map[string]interface{})

		schema["id"] = block_document.Id
		schema["created"] = block_document.Created
		schema["updated"] = block_document.Updated
		schema["name"] = block_document.Name
		schema["block_schema_id"] = block_document.BlockSchemaId
		schema["block_type_id"] = block_document.BlockTypeId
		schema["is_anonymous"] = block_document.IsAnonymous
		schema["data"] = block_document.Data
		schema["block_document_references"] = block_document.BlockDocumentReferences

		schemaOutput[i] = schema
	}
	return schemaOutput
}
