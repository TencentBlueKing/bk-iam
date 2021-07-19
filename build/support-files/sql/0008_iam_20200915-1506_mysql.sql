CREATE TABLE IF NOT EXISTS `bkiam`.`subject_role` (
  `pk` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `role_type` VARCHAR(32) NOT NULL,
  `system_id` VARCHAR(32) NOT NULL DEFAULT "",
  `subject_pk` INT UNSIGNED NOT NULL,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`pk`),
  UNIQUE KEY `idx_uk_type_system_subject` (`role_type`, `system_id`, `subject_pk`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
