provider "alternator" {}

terraform {
  required_providers {
    alternator = {
      source  = "kota65535/alternator"
      version = ">= 0.0.1"
    }
  }
  required_version = ">= 1.1.0"
}

resource "alternator_mysql" "main" {
  host     = "localhost:23306"
  user     = "root"
  database = "example"
  schema   = file("${path.root}/schema.sql")
}
