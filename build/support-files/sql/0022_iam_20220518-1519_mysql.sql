CREATE TABLE `bkiam`.`subject_system_group` (
  `pk` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `system_id` varchar(32) NOT NULL,
  `subject_pk` int(10) unsigned NOT NULL,
  `groups` mediumtext NOT NULL, -- JSON
  `updates` int(10) unsigned NOT NULL DEFAULT 0,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`pk`),
  UNIQUE KEY `idx_uk_subject_system` (`subject_pk`,`system_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `bkiam`.`group_system_auth_type` (
  `pk` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `system_id` varchar(32) NOT NULL,
  `group_pk` int(10) unsigned NOT NULL,
  `auth_type` tinyint(3) unsigned NOT NULL, -- 0 none 1 abac 2 rbac 3 rabc and abac, 支持位运算
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`pk`),
  UNIQUE KEY `idx_uk_group_system` (`group_pk`,`system_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
