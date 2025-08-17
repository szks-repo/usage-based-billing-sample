CREATE TABLE IF NOT EXISTS `account_free_credit_balance` (
    `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
    `account_id` bigint UNSIGNED NOT NULL,
    `credit` int NOT NULL DEFAULT 0,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    PRIMARY KEY (`id`),
    FOREIGN KEY (`account_id`) REFERENCES `account`(`id`)
) DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
