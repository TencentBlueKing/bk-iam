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
  `signature` CHAR(32) NOT NULL,
  `expired_at` int(10) unsigned NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`pk`),
  UNIQUE KEY `idx_uk_subject_action` (`subject_pk`,`action_pk`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `bkiam`.`rbac_group_alter_event` (
  `uuid` char(32) NOT NULL, -- uuid
  `group_pk` int(10) unsigned NOT NULL,
  `action_pks` mediumtext NOT NULL, -- json [action_pk]
  `subject_pks` mediumtext NOT NULL, -- json [subject_pk]
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `bkiam`.`rbac_subject_action_alter_message` (
  `uuid` char(32) NOT NULL, --uuid
  `data` int(10) unsigned NOT NULL, -- json
  `status` int(10) unsigned NOT NULL DEFAULT '0',
  `check_count` int(10) unsigned NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`uuid`),
  KEY `idx_created_at_status` (`created_at`,`status`),
  KEY `idx_updated_at_status` (`updated_at`,`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
