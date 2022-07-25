package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAlternatorDatabaseSchema() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Fetch SQL database schemas.",
		ReadContext: dataSourceAlternatorDatabaseSchemaRead,

		Schema: map[string]*schema.Schema{
			"database": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Target database name.",
			},
			"remote_schema": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Actual remote database schema definition.",
			},
		},
	}
}

func dataSourceAlternatorDatabaseSchemaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	database := d.Get("database").(string)
	pp := meta.(*ProviderParams)
	client, err := newAlternator(database, pp)
	if err != nil {
		return diag.FromErr(err)
	}
	defer client.Close()

	remoteSchema, err := client.FetchSchemas()
	if err != nil {
		return diag.FromErr(err)
	}
	remoteSchemaStr := ""
	for _, s := range remoteSchema {
		remoteSchemaStr += fmt.Sprintf("%s\n", s)
	}
	tflog.Debug(ctx, fmt.Sprintf("@read remote_schema: %s", remoteSchemaStr))

	d.SetId(database)
	d.Set("remote_schema", remoteSchemaStr)

	return nil
}
