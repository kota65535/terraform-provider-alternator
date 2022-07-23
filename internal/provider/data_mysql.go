package provider

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAlternatorMysql() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Manage a MySQL database schema by Alternator.",
		ReadContext: dataSourceAlternatorMySqlRead,

		Schema: map[string]*schema.Schema{
			"database": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Target database name",
			},
			"host": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Host name to connect to",
			},
			"user": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "User name to connect as",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Password to be used if the server demands password authentication",
			},
			"remote_schema": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Actual remote database schema definition",
			},
		},
	}
}

func dataSourceAlternatorMySqlRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := newAlternator(d)
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

	h := sha1.New()
	h.Write([]byte(client.DbUri.Host + client.DbUri.DbName))
	sha1 := hex.EncodeToString(h.Sum(nil))
	d.SetId(sha1)

	d.Set("remote_schema", remoteSchemaStr)

	return nil
}
