CREATE TABLE IF NOT EXISTS `account_contract` (
    `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
    `account_id` bigint UNSIGNED NOT NULL,
    `from` DATETIME NOT NULL, 
    `estimated_to` DATETIME NOT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    PRIMARY KEY (`id`),
    UNIQUE(`account_id`, `from`),
    FOREIGN KEY (`account_id`) REFERENCES `account`(`id`)
) DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
