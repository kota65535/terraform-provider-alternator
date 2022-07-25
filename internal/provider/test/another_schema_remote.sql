CREATE DATABASE `example2`
    DEFAULT COLLATE = utf8mb4_0900_bin;
CREATE TABLE `example2`.`greeting`
(
    `id`         int          NOT NULL AUTO_INCREMENT,
    `body`       varchar(255),
    `created_at` datetime     DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime     DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
);
