CREATE TABLE IF NOT EXISTS `invoice` (
    `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
    `account_id` bigint UNSIGNED NOT NULL,
    `subscription_id` bigint UNSIGNED NOT NULL,
    `total_usage` int UNSIGNED NOT NULL DEFAULT 0,
    `tax_rate` tinyint UNSIGNED NOT NULL DEFAULT 10,
    `tax_amount` DECIMAL(20, 5) NOT NULL DEFAULT 0,
    `subtotal` DECIMAL(20, 5) NOT NULL DEFAULT 0,
    `free_credit_discount` int UNSIGNED NOT NULL DEFAULT 0,
    `total_price` DECIMAL(20, 5) NOT NULL DEFAULT 0,
    `total_price_tax_included` DECIMAL(20, 5) NOT NULL DEFAULT 0,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    PRIMARY KEY (`id`),
    KEY(`created_at`),
    FOREIGN KEY (`account_id`) REFERENCES `account`(`id`),
    FOREIGN KEY (`subscription_id`) REFERENCES `subscription`(`id`)
) DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
