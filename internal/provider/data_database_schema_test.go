package provider

import (
	"database/sql"
	_ "embed"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAlternatorDatabaseSchema(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					db, err := sql.Open("mysql", "root@tcp(localhost:23306)/?multiStatements=true")
					if err != nil {
						panic(err)
					}
					_, err = db.Exec("DROP DATABASE IF EXISTS example")
					if err != nil {
						panic(err)
					}
					_, err = db.Exec(initialSchema)
					if err != nil {
						panic(err)
					}
				},
				Config: testAccDataSourceAlternatorDatabaseSchemaInitialConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.alternator_database_schema.main", "remote_schema", initialSchemaRemote),
				),
			},
		},
	})
}

func testAccDataSourceAlternatorDatabaseSchemaInitialConfig() string {
	return fmt.Sprintf(`
	%s
	data "alternator_database_schema" "main" {
        database = "example"
	}
	`, provider)
}
