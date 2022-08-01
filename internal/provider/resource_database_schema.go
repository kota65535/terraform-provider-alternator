package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/kota65535/alternator/cmd"
	"strings"
)

func resourceAlternatorDatabaseSchema() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage SQL database schemas by Alternator.",
		CreateContext: resourceAlternatorDatabaseSchemaCreate,
		ReadContext:   resourceAlternatorDatabaseSchemaRead,
		UpdateContext: resourceAlternatorDatabaseSchemaUpdate,
		DeleteContext: resourceAlternatorDatabaseSchemaDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceAlternatorDatabaseSchemaImport,
		},
		Schema: map[string]*schema.Schema{
			"database": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Target database name.",
			},
			"schema": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "SQL Database schema definition, composed by DDL statements.",
			},
			"remote_schema": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Actual remote database schema definition.",
			},
			"changed": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Used by the provider internal.",
			},
			"statements": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Statements to execute on apply.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
			tflog.Debug(ctx, fmt.Sprintf("@diff start"))

			// We can easily detect change of the input variables in this way
			localSchemaChanged := d.HasChange("schema")
			// As for the computed variables, we cannot simply compare their old & new value,
			// because their old value has already been updated to match the result of our read function.
			// So we have to use the dedicated boolean computed variable.
			// cf. https://discuss.hashicorp.com/t/force-new-resource-based-on-api-read-difference/29759/3
			remoteSchemaChanged := d.Get("changed").(bool)
			if localSchemaChanged || remoteSchemaChanged {
				database := d.Get("database").(string)
				schemaStr := d.Get("schema").(string)
				pp := meta.(*ProviderArguments)
				// Provider's host argument will be empty when it is a new resource output.
				if pp.Host == "" {
					tflog.Debug(ctx, fmt.Sprintf("@diff host is empty. arguments: %+v", pp))
					return nil
				}

				// If the host argument has been changed, alternator initialization may fail with the old host value.
				// We ignore the error here to continue the plan phase.
				client, err := newAlternator(database, pp)
				if err != nil {
					tflog.Debug(ctx, fmt.Sprintf("@diff failed to initialize alternator: %s", err.Error()))
					return nil
				}
				// Read local schema
				alt, _, localSchema, err := client.GetAlterations(schemaStr)
				if err != nil {
					return err
				}
				// Use local schema as new remote schema value to show diff on planing
				newRemoteSchemaStr := ""
				for _, s := range localSchema {
					newRemoteSchemaStr += fmt.Sprintf("%s\n", s)
				}
				statements := alt.Statements()
				tflog.Debug(ctx, fmt.Sprintf("@diff remote_schema: %s", newRemoteSchemaStr))
				tflog.Debug(ctx, fmt.Sprintf("@diff statements: %s", statements))
				err = d.SetNew("remote_schema", newRemoteSchemaStr)
				if err != nil {
					return err
				}
				// statements variable is only for showing diff on planning, and always empty value after applying it.
				err = d.SetNew("statements", statements)
				if err != nil {
					return err
				}
			}

			tflog.Debug(ctx, fmt.Sprintf("@diff end"))
			return nil
		},
	}
}

func resourceAlternatorDatabaseSchemaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Debug(ctx, fmt.Sprintf("@create start"))

	database := d.Get("database").(string)
	schemaStr := d.Get("schema").(string)
	pp := meta.(*ProviderArguments)

	client, err := newAlternator(database, pp)
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
		tflog.Info(ctx, fmt.Sprintf("@create executing statements: %s", s))
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

	err = d.Set("remote_schema", remoteSchemaStr)
	if err != nil {
		diag.FromErr(err)
	}
	err = d.Set("changed", false)
	if err != nil {
		diag.FromErr(err)
	}
	err = d.Set("statements", []string{})
	if err != nil {
		diag.FromErr(err)
	}
	d.SetId(database)

	tflog.Debug(ctx, fmt.Sprintf("@create end"))
	return nil
}

func resourceAlternatorDatabaseSchemaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Debug(ctx, fmt.Sprintf("@read start"))

	database := d.Get("database").(string)
	schemaStr := d.Get("schema").(string)
	pp := meta.(*ProviderArguments)
	// Provider's host argument will be empty when it is a new resource output, which leads to the resource re-creation.
	if pp.Host == "" {
		tflog.Debug(ctx, fmt.Sprintf("@read host is empty. arguments: %+v", pp))
		d.SetId("")
		return nil
	}

	client, err := newAlternator(database, pp)
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

	err = d.Set("remote_schema", remoteSchemaStr)
	if err != nil {
		diag.FromErr(err)
	}
	err = d.Set("changed", changed)
	if err != nil {
		diag.FromErr(err)
	}
	err = d.Set("statements", []string{})
	if err != nil {
		diag.FromErr(err)
	}
	d.SetId(database)

	tflog.Debug(ctx, fmt.Sprintf("@read end"))
	return nil
}

func resourceAlternatorDatabaseSchemaUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Debug(ctx, fmt.Sprintf("@update start"))

	database := d.Get("database").(string)
	schemaStr := d.Get("schema").(string)
	pp := meta.(*ProviderArguments)

	client, err := newAlternator(database, pp)
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
		tflog.Info(ctx, fmt.Sprintf("@update executing statements: %s", s))
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

	tflog.Debug(ctx, fmt.Sprintf("@update remote_schema: %s", remoteSchemaStr))

	err = d.Set("remote_schema", remoteSchemaStr)
	if err != nil {
		diag.FromErr(err)
	}
	err = d.Set("changed", false)
	if err != nil {
		diag.FromErr(err)
	}
	err = d.Set("statements", []string{})
	if err != nil {
		diag.FromErr(err)
	}
	d.SetId(database)

	tflog.Debug(ctx, fmt.Sprintf("@update end"))
	return nil
}

func resourceAlternatorDatabaseSchemaDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Debug(ctx, fmt.Sprintf("@delete start"))

	database := d.Get("database").(string)
	pp := meta.(*ProviderArguments)

	client, err := newAlternator(database, pp)
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
		tflog.Info(ctx, fmt.Sprintf("@delete executing statements: %s", s))
		_, err := client.Db.Exec(s)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId("")

	tflog.Debug(ctx, fmt.Sprintf("@delete end"))
	return nil
}

func resourceAlternatorDatabaseSchemaImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	database := d.Id()
	tflog.Debug(ctx, fmt.Sprintf("@import id = %s", database))
	err := d.Set("database", database)
	if err != nil {
		diag.FromErr(err)
	}
	d.SetId(database)
	return []*schema.ResourceData{d}, nil
}

func newAlternator(database string, p *ProviderArguments) (*cmd.Alternator, error) {
	dbUri := &cmd.DatabaseUri{
		Dialect:  p.Dialect,
		Host:     p.Host,
		User:     p.User,
		Password: p.Password,
		DbName:   database,
	}

	return cmd.NewAlternator(dbUri)
}
