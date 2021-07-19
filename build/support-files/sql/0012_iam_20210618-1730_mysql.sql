CREATE TABLE IF NOT EXISTS `bkiam`.`model_change_event` (
  `pk` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `type` VARCHAR(32) NOT NULL,
  `status` VARCHAR(32) NOT NULL,
  `system_id` VARCHAR(32) NOT NULL,
  `model_type` VARCHAR(32) NOT NULL,
  `model_id` VARCHAR(32) NOT NULL,
  `model_pk` INT UNSIGNED NOT NULL,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`pk`),
  INDEX `idx_type_model` (`status`, `type`, `model_type`, `model_pk`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
