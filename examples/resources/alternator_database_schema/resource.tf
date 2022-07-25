resource "alternator_database_schema" "example" {
  database = "example"
  schema   = <<EOF
CREATE DATABASE example;
USE example;
CREATE TABLE users
(
    id   int PRIMARY KEY,
    name varchar(100)
);
CREATE TABLE blog_posts
(
    id        int PRIMARY KEY,
    title     varchar(100),
    body      text,
    author_id int,
    FOREIGN KEY (author_id) REFERENCES users (id)
);
EOF
}
