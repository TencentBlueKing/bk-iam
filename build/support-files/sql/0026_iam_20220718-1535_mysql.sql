CREATE TABLE IF NOT EXISTS `bkiam`.`rbac_subject_action_group_resource` (
  `pk` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `subject_pk` int(10) unsigned NOT NULL,
  `action_pk` int(10) unsigned NOT NULL,
  `group_resource` mediumtext NOT NULL, -- json
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`pk`),
  UNIQUE KEY `idx_uk_subject_action` (`subject_pk`,`action_pk`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `bkiam`.`rbac_subject_action_expression` (
  `pk` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `subject_pk` int(10) unsigned NOT NULL,
  `action_pk` int(10) unsigned NOT NULL,
  `expression` mediumtext NOT NULL,
  `expired_at` int(10) unsigned NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`pk`),
  UNIQUE KEY `idx_uk_subject_action` (`subject_pk`,`action_pk`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `bkiam`.`rbac_group_alter_event` (
  `pk` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `task_id` CHAR(32) NOT NULL,
  `group_pk` int(10) unsigned NOT NULL,
  `action_pks` mediumtext NOT NULL, -- json [action_pk]
  `subject_pks` mediumtext NOT NULL, -- json [subject_pk]
  `check_count` int(10) unsigned NOT NULL DEFAULT 0,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`pk`),
  INDEX `idx_check_count_created_at` (`check_count`,`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
