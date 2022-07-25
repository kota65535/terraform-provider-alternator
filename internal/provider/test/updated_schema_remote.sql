CREATE DATABASE `example`
    DEFAULT COLLATE = utf8mb4_0900_bin;
CREATE TABLE `example`.`greeting`
(
    `id`         int          NOT NULL AUTO_INCREMENT,
    `body`       varchar(256),
    `created_at` datetime     DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime     DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
);
