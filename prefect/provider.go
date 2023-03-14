package prefect

import (
	"context"

	"terraform-provider-prefect/prefect_api"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"prefect_api_url": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PREFECT_API_URL", nil),
			},
			"prefect_api_key": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PREFECT_API_KEY", nil),
			},
			"prefect_account_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("PREFECT_ACCOUNT_ID", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"prefect_workspace":  resourceWorkspace(),
			"prefect_work_queue": resourceWorkQueue(),
			"prefect_block":      resourceBlock(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"prefect_workspaces":      dataSourceWorkspaces(),
			"prefect_work_queues":     dataSourceWorkQueues(),
			"prefect_block_types":     dataSourceBlockTypes(),
			"prefect_block_schemas":   dataSourceBlockSchemas(),
			"prefect_block_documents": dataSourceBlockDocuments(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	prefect_api_url := d.Get("prefect_api_url").(string)
	prefect_api_key := d.Get("prefect_api_key").(string)
	prefect_account_id := d.Get("prefect_account_id").(string)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	c, err := prefect_api.NewClient(prefect_api_url, prefect_api_key, prefect_account_id)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return c, diags
}
