package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	schema.DescriptionKind = schema.StringMarkdown

	// Customize the content of descriptions when output. For example you can add defaults on
	// to the exported descriptions if present.
	schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
		desc := s.Description
		if s.Default != nil {
			desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
		}
		return strings.TrimSpace(desc)
	}
}

type ProviderParams struct {
	Host     string
	Dialect  string
	User     string
	Password string
}

func New() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"host": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Host on which database server is located. If port number is not specified, the default value is used according to the SQL dialect (ex: mysql -> 3306). ",
				DefaultFunc: schema.EnvDefaultFunc("ALTERNATOR_HOST", ""),
			},
			"dialect": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "SQL dialect. Currently, only \"mysql\" is supported.",
				DefaultFunc:  schema.EnvDefaultFunc("ALTERNATOR_DIALECT", ""),
				ValidateFunc: validation.StringInSlice([]string{"mysql"}, true),
			},
			"user": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "User name to use when connecting to server",
				DefaultFunc: schema.EnvDefaultFunc("ALTERNATOR_USER", ""),
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Password to use when connecting to server",
				DefaultFunc: schema.EnvDefaultFunc("ALTERNATOR_PASSWORD", ""),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"alternator_database_schema": resourceAlternatorDatabaseSchema(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"alternator_database_schema": dataSourceAlternatorDatabaseSchema(),
		},
		ConfigureContextFunc: configure(),
	}
}

func configure() func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(cxt context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		params := &ProviderParams{
			Host:     d.Get("host").(string),
			Dialect:  d.Get("dialect").(string),
			User:     d.Get("user").(string),
			Password: d.Get("password").(string),
		}
		return params, nil
	}
}
