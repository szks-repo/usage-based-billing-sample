CREATE DATETIME IF NOT EXISTS `usage_based_billing`;

CREATE USER IF NOT EXISTS 'user'@'%' IDENTIFIED BY 'password';

GRANT ALL PRIVILEGES ON `usage_based_billing`.* TO 'user'@'%';

FLUSH PRIVILEGES;