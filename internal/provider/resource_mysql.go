package provider

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/kota65535/alternator/cmd"
	"strings"
)

func resourceAlternatorMysql() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Manage a MySQL database schema by Alternator.",
		CreateContext: resourceAlternatorMySqlCreate,
		ReadContext:   resourceAlternatorMySqlRead,
		UpdateContext: resourceAlternatorMySqlUpdate,
		DeleteContext: resourceAlternatorMySqlDelete,

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
			"schema": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "SQL Database schema definition, composed by DDL statements",
			},
			"remote_schema": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Actual remote database schema definition",
			},
			"changed": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Used by the provider internal",
			},
		},
		CustomizeDiff: func(ctx context.Context, diff *schema.ResourceDiff, i interface{}) error {
			changed := diff.Get("changed").(bool)
			if changed {
				// Fetch current remote database schemas
				schemaStr := diff.Get("schema").(string)
				client, err := newAlternator(diff)
				if err != nil {
					return err
				}
				localSchema, err := client.ReadSchemas(schemaStr)
				if err != nil {
					return err
				}
				newRemoteSchemaStr := ""
				for _, s := range localSchema {
					newRemoteSchemaStr += fmt.Sprintf("%s\n", s)
				}
				tflog.Debug(ctx, fmt.Sprintf("@diff remote_schema: %s", newRemoteSchemaStr))
				diff.SetNew("remote_schema", newRemoteSchemaStr)
			}

			return nil
		},
	}
}

func resourceAlternatorMySqlCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	schemaStr := d.Get("schema").(string)

	client, err := newAlternator(d)
	if err != nil {
		return diag.FromErr(err)
	}
	defer client.Close()

	// Create remote database
	statements := []string{}
	for _, s := range strings.Split(schemaStr, ";") {
		trimmed := strings.TrimSpace(s)
		if trimmed != "" {
			statements = append(statements, trimmed)
		}
	}
	for _, s := range statements {
		_, err := client.Db.Exec(s)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// Fetch current remote database schemas
	remoteSchema, err := client.FetchSchemas()
	if err != nil {
		return diag.FromErr(err)
	}
	remoteSchemaStr := ""
	for _, s := range remoteSchema {
		remoteSchemaStr += fmt.Sprintf("%s\n", s)
	}

	tflog.Debug(ctx, fmt.Sprintf("@create remote_schema: %s", remoteSchemaStr))

	h := sha1.New()
	h.Write([]byte(client.DbUri.Host + client.DbUri.DbName))
	sha1 := hex.EncodeToString(h.Sum(nil))
	d.SetId(sha1)

	d.Set("remote_schema", remoteSchemaStr)
	d.Set("changed", false)

	return nil
}

func resourceAlternatorMySqlRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	schemaStr := d.Get("schema").(string)

	client, err := newAlternator(d)
	if err != nil {
		return diag.FromErr(err)
	}
	defer client.Close()

	alt, remoteSchema, _, err := client.GetAlterations(schemaStr)
	if err != nil {
		return diag.FromErr(err)
	}
	remoteSchemaStr := ""
	for _, s := range remoteSchema {
		remoteSchemaStr += fmt.Sprintf("%s\n", s)
	}

	changed := len(alt.Statements()) > 0

	tflog.Debug(ctx, fmt.Sprintf("@read remote_schema: %s", remoteSchemaStr))
	tflog.Debug(ctx, fmt.Sprintf("@read changed: %t", changed))

	d.Set("remote_schema", remoteSchemaStr)
	d.Set("changed", changed)

	return nil
}

func resourceAlternatorMySqlUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	schemaStr := d.Get("schema").(string)

	client, err := newAlternator(d)
	if err != nil {
		return diag.FromErr(err)
	}
	defer client.Close()

	// Update remote database schemas
	alt, _, _, err := client.GetAlterations(schemaStr)
	if err != nil {
		return diag.FromErr(err)
	}
	for _, s := range alt.Statements() {
		tflog.Debug(ctx, fmt.Sprintf("Executing statements: %s", s))
		_, err := client.Db.Exec(s)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// Fetch current remote database schemas
	remoteSchema, err := client.FetchSchemas()
	if err != nil {
		return diag.FromErr(err)
	}
	remoteSchemaStr := ""
	for _, s := range remoteSchema {
		remoteSchemaStr += fmt.Sprintf("%s\n", s)
	}

	tflog.Debug(ctx, fmt.Sprintf("@update remote_schema: %s", remoteSchema))

	d.Set("remote_schema", remoteSchemaStr)
	d.Set("changed", false)

	return nil
}

func resourceAlternatorMySqlDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := newAlternator(d)
	if err != nil {
		return diag.FromErr(err)
	}
	defer client.Close()

	// Delete remote database schemas
	alt, _, _, err := client.GetAlterations("")
	if err != nil {
		return diag.FromErr(err)
	}
	for _, s := range alt.Statements() {
		tflog.Debug(ctx, fmt.Sprintf("Executing statements: %s", s))
		_, err := client.Db.Exec(s)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

type DataMap interface {
	Get(key string) interface{}
}

func newAlternator(d DataMap) (*cmd.Alternator, error) {
	database := d.Get("database").(string)
	host := d.Get("host").(string)
	user := d.Get("user").(string)
	password := d.Get("password").(string)

	dbUri := &cmd.DatabaseUri{
		Dialect:  "mysql",
		Host:     host,
		DbName:   database,
		User:     user,
		Password: password,
	}

	return cmd.NewAlternator(dbUri)
}
