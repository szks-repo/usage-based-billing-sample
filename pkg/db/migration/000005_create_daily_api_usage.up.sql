CREATE TABLE IF NOT EXISTS `daily_api_usage` (
    `account_id` bigint UNSIGNED NOT NULL DEFAULT 0,
    `date` VARCHAR(8) NOT NULL, -- 20240102
    `usage` bigint UNSIGNED NOT NULL DEFAULT 0,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP(),
    PRIMARY KEY (`account_id`, `munite_key`),
    CONSTRAINT `fk_account_id`
        FOREIGN KEY (`account_id`)
        REFERENCES `accounts` (`id`)
) DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
