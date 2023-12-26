CREATE TABLE `bkiam`.`subject_template_group` (
  `pk` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `subject_pk` int(10) unsigned NOT NULL,
  `template_id` int(10) unsigned NOT NULL,
  `group_pk` int(10) unsigned NOT NULL,
  `expired_at` int(10) unsigned NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`pk`),
  UNIQUE KEY `idx_uk_subject_template_group` (`subject_pk`,`group_pk`,`template_id`),
  INDEX `idx_group_template_subject` (`group_pk`, `template_id`,`subject_pk`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;