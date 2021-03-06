CREATE TABLE IF NOT EXISTS `bkiam`.`subject_black_list` (
  `pk` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `subject_pk` INT UNSIGNED NOT NULL,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`pk`),
  UNIQUE KEY `idx_uk_subject_pk` (`subject_pk`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;