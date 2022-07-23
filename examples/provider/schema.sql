CREATE DATABASE `example`
    DEFAULT CHARACTER SET = utf8mb4
    DEFAULT COLLATE = utf8mb4_0900_bin
    DEFAULT ENCRYPTION = 'N';
USE `example`;
CREATE TABLE `greeting`
(
    `id`         int          NOT NULL AUTO_INCREMENT,
    `body`       varchar(250),
    `created_at` datetime     DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime     DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
)