// Basic usage
provider "alternator" {
  dialect = "mysql"
  host    = "mydb.dev.example.com"
  user    = "root"
}

// Specify host argument using a resource output
// cf. https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/db_instance
provider "alternator" {
  dialect  = "mysql"
  host     = aws_db_instance.main.endpoint
  user     = "bob"
  password = "secret"
}
