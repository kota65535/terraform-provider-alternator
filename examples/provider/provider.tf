// For the local database
provider "alternator" {
  dialect = "mysql"
  host    = "localhost"
  user    = "root"
}

// For the remote database
provider "alternator" {
  dialect  = "mysql"
  host     = "dev.example.com:3307"
  user     = "bob"
  password = "secret"
}
