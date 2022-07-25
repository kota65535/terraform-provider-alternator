resource "alternator_database_schema" "example" {
  database = "example"
  schema   = <<EOF
CREATE DATABASE example
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_0900_bin;

USE example;

CREATE TABLE greeting
(
    id         int AUTO_INCREMENT,
    body       varchar(255),
    created_at datetime DEFAULT CURRENT_TIMESTAMP,
    updated_at datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
);
EOF
}
