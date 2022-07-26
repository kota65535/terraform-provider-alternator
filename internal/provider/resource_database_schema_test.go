package provider

import (
	"database/sql"
	_ "embed"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

//go:embed test/provider.tf
var provider string

//go:embed test/initial_schema.sql
var initialSchema string

//go:embed test/initial_schema_remote.sql
var initialSchemaRemote string

//go:embed test/updated_schema.sql
var updatedSchema string

//go:embed test/updated_schema_remote.sql
var updatedSchemaRemote string

//go:embed test/another_schema.sql
var anotherSchema string

//go:embed test/another_schema_remote.sql
var anotherSchemaRemote string

func TestAccResourceAlternatorDatabaseSchema(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			// Create
			{
				PreConfig: func() {
					db, err := sql.Open("mysql", "root@tcp(localhost:23306)/?multiStatements=true")
					require.NoError(t, err)
					// drop target table
					_, err = db.Exec("DROP DATABASE IF EXISTS example")
					require.NoError(t, err)
					// create another table
					_, err = db.Exec("DROP DATABASE IF EXISTS example2")
					require.NoError(t, err)
					_, err = db.Exec(anotherSchema)
					require.NoError(t, err)
				},
				Config: testAccResourceAlternatorDatabaseSchemaInitialConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("alternator_database_schema.main", "id", "example"),
					resource.TestCheckResourceAttr("alternator_database_schema.main", "remote_schema", initialSchemaRemote),
				),
			},
			// Update schema
			{
				Config: testAccResourceAlternatorDatabaseSchemaUpdatedConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("alternator_database_schema.main", "id", "example"),
					resource.TestCheckResourceAttr("alternator_database_schema.main", "remote_schema", updatedSchemaRemote),
				),
			},
			// Change schema from outside
			{
				PreConfig: func() {
					db, err := sql.Open("mysql", "root@tcp(localhost:23306)/")
					if err != nil {
						panic(err)
					}
					_, err = db.Exec("ALTER TABLE example.greeting MODIFY COLUMN body varchar(200)")
					if err != nil {
						panic(err)
					}
				},
				Config: testAccResourceAlternatorDatabaseSchemaUpdatedConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("alternator_database_schema.main", "id", "example"),
					resource.TestCheckResourceAttr("alternator_database_schema.main", "remote_schema", updatedSchemaRemote),
				),
			},
			// Import another database schema
			{
				Config:       testAccResourceAlternatorDatabaseSchemaAnotherConfig(),
				ResourceName: "alternator_database_schema.main",
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return "example2", nil
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("alternator_database_schema.main", "id", "example2"),
					resource.TestCheckResourceAttr("alternator_database_schema.main", "remote_schema", anotherSchemaRemote),
				),
			},
			// Recreate
			{
				Config: testAccResourceAlternatorDatabaseSchemaInitialConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("alternator_database_schema.main", "id", "example"),
					resource.TestCheckResourceAttr("alternator_database_schema.main", "remote_schema", initialSchemaRemote),
				),
			},
		},
	})
}

func testAccResourceAlternatorDatabaseSchemaInitialConfig() string {
	return fmt.Sprintf(`
    %s
	resource "alternator_database_schema" "main" {
        database = "example"
        schema = <<EOT
		%s
		EOT
	}
	`, provider, initialSchema)
}

func testAccResourceAlternatorDatabaseSchemaUpdatedConfig() string {
	return fmt.Sprintf(`
    %s
	resource "alternator_database_schema" "main" {
        database = "example"
        schema = <<EOT
		%s
		EOT
	}
	`, provider, updatedSchema)
}

func testAccResourceAlternatorDatabaseSchemaAnotherConfig() string {
	return fmt.Sprintf(`
    %s
	resource "alternator_database_schema" "main" {
        database = "example2"
        schema = <<EOT
		%s
		EOT
	}
	`, provider, anotherSchema)
}
