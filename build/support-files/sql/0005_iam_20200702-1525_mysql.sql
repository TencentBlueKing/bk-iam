CREATE TABLE IF NOT EXISTS `bkiam`.`saas_system_config` (
  `pk` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `system_id` VARCHAR(32) NOT NULL,
  `name` VARCHAR(32) NOT NULL,
  `type` VARCHAR(32) NOT NULL DEFAULT "json",
  `value` TEXT NOT NULL,  /* JSON */
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY ( `pk` ),
  UNIQUE KEY `idx_uk_system_name` (`system_id`, `name`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;


ALTER TABLE `bkiam`.`saas_system_info` DROP COLUMN `action_topology`;
