package provider

import (
	"database/sql"
	_ "embed"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

//go:embed test/initial_schema.sql
var initialSchema string

//go:embed test/initial_schema_remote.sql
var initialSchemaRemote string

//go:embed test/updated_schema.sql
var updatedSchema string

//go:embed test/updated_schema_remote.sql
var updatedSchemaRemote string

func TestAccResourceAlternatorMySql(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					db, err := sql.Open("mysql", "root@tcp(localhost:23306)/")
					if err != nil {
						panic(err)
					}
					_, err = db.Exec("DROP DATABASE IF EXISTS example")
					if err != nil {
						panic(err)
					}
				},
				Config: testAccResourceAlternatorMySqlInitialConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("alternator_mysql.main", "remote_schema", initialSchemaRemote),
				),
			},
			{
				Config: testAccResourceAlternatorMySqlUpdatedConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("alternator_mysql.main", "remote_schema", updatedSchemaRemote),
				),
			},
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
				Config: testAccResourceAlternatorMySqlUpdatedConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("alternator_mysql.main", "remote_schema", updatedSchemaRemote),
				),
			},
		},
	})
}

func testAccResourceAlternatorMySqlInitialConfig() string {
	return fmt.Sprintf(`
	resource "alternator_mysql" "main" {
        schema = <<EOT
		%s
		EOT
        
 		host     = "localhost:23306"
        user     = "root"
        database = "example"
	}
	`, initialSchema)
}

func testAccResourceAlternatorMySqlUpdatedConfig() string {
	return fmt.Sprintf(`
	resource "alternator_mysql" "main" {
        schema = <<EOT
		%s
		EOT
        
 		host     = "localhost:23306"
        user     = "root"
        database = "example"
	}
	`, updatedSchema)
}
