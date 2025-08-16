CREATE TABLE IF NOT EXISTS `active_api_key` (
    `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
    `account_id` bigint UNSIGNED NOT NULL DEFAULT 0,
    `api_key` VARCHAR(64) NOT NULL,
    `expired_at` DATETIME NOT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    PRIMARY KEY (`id`),
    UNIQUE INDEX (`api_key`)
) DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

