/* 后台相关表 */
CREATE TABLE IF NOT EXISTS `bkiam`.`system_info` (
   `id` VARCHAR(32) NOT NULL,
   `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
   `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
   PRIMARY KEY ( `id` )
)ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `bkiam`.`subject`(
   `pk` INT UNSIGNED AUTO_INCREMENT,
   `type` VARCHAR(32) NOT NULL,
   `id` VARCHAR(64) NOT NULL,
   `name` VARCHAR(64) NOT NULL DEFAULT "",
   `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
   `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
   PRIMARY KEY ( `pk` ),
   UNIQUE KEY `idx_uk_id_type` (`id`, `type`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `bkiam`.`subject_relation`(
   `pk` INT UNSIGNED AUTO_INCREMENT,
   `subject_pk` INT UNSIGNED NOT NULL,
   `subject_type` VARCHAR(16) NOT NULL,
   `subject_id` VARCHAR(64) NOT NULL,
   `parent_pk` INT UNSIGNED NOT NULL,
   `parent_type` VARCHAR(16) NOT NULL,
   `parent_id` VARCHAR(64) NOT NULL,
   `policy_expired_at` INT UNSIGNED NOT NULL,
   `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
   `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
   PRIMARY KEY ( `pk` ),
   INDEX `idx_subject` (`subject_id`, `subject_type`),
   INDEX `idx_parent` (`parent_id`, `parent_type`),
   UNIQUE KEY `idx_uk_subject_parent` (`subject_pk`, `parent_pk`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `bkiam`.`resource_type`(
   `pk` INT UNSIGNED AUTO_INCREMENT,
   `system_id` VARCHAR(32) NOT NULL,
   `id` VARCHAR(32) NOT NULL,
   `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
   `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
   PRIMARY KEY ( `pk` ),
   UNIQUE KEY `idx_uk_system_rt_id` (`system_id`, `id`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `bkiam`.`action`(
   `pk` INT UNSIGNED AUTO_INCREMENT,
   `system_id` VARCHAR(32) NOT NULL,
   `id` VARCHAR(32) NOT NULL,
   `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
   `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
   PRIMARY KEY ( `pk` ),
   UNIQUE KEY `idx_uk_system_action_id` (`system_id`, `id`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `bkiam`.`action_resource_type`(
   `pk` INT UNSIGNED AUTO_INCREMENT,
   `action_system_id` VARCHAR(32) NOT NULL,
   `action_id` VARCHAR(32) NOT NULL,
   `resource_type_system_id` VARCHAR(32) NOT NULL,
   `resource_type_id` VARCHAR(32) NOT NULL,
   `scope_expression` TEXT NULL, /* JSON */
   `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
   `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
   PRIMARY KEY ( `pk` ),
   INDEX `idx_action` (`action_system_id`, `action_id`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `bkiam`.`expression` (
  `pk` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `action_pk` INT UNSIGNED NOT NULL,
  `expression` TEXT NOT NULL,  /* JSON */
  `signature` CHAR(32) NOT NULL,
  `template_id` INT UNSIGNED NOT NULL DEFAULT 0,
  `template_version` INT UNSIGNED NOT NULL DEFAULT 0,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`pk`),
  KEY `idx_template_version` (`template_id`,`template_version`),
  KEY `idx_signature_action` (`signature`, `action_pk`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `bkiam`.`policy` (
  `pk` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `subject_pk` INT UNSIGNED NOT NULL,
  `action_pk` INT UNSIGNED NOT NULL,
  `expression_pk` INT NOT NULL,
  `expired_at` INT UNSIGNED NOT NULL,
  `template_id` INT UNSIGNED NOT NULL DEFAULT 0,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`pk`),
  KEY `idx_uk_subject_expression` (`subject_pk`, `expression_pk`),
  KEY `idx_subject_action_expire` (`subject_pk`,`action_pk`,`expired_at`),
  KEY `idx_subject_template` (`subject_pk`,`template_id`),
  KEY `idx_expression_expire` (`expression_pk`, `expired_at`),
  KEY `idx_action_expire` (`action_pk`, `expired_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


/* SaaS相关表 */
CREATE TABLE IF NOT EXISTS `bkiam`.`saas_system_info` (
   `id` VARCHAR(32) NOT NULL,
   `name` VARCHAR(255) NOT NULL,
   `name_en` VARCHAR(255) NOT NULL,
   `description` VARCHAR(1024) NOT NULL DEFAULT "",
   `description_en` VARCHAR(1024) NOT NULL DEFAULT "",
   `clients` VARCHAR(1024) NOT NULL,
   `provider_config` TEXT NOT NULL,  /* JSON */
   `action_topology` TEXT NOT NULL,  /* JSON */
   `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
   `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
   PRIMARY KEY ( `id` )
)ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `bkiam`.`saas_resource_type`(
   `pk` INT UNSIGNED AUTO_INCREMENT,
   `system_id` VARCHAR(32) NOT NULL,
   `id` VARCHAR(32) NOT NULL,
   `name` VARCHAR(255) NOT NULL,
   `name_en` VARCHAR(255) NOT NULL,
   `description` VARCHAR(1024) NOT NULL DEFAULT "",
   `description_en` VARCHAR(1024) NOT NULL DEFAULT "",
   `parents` TEXT NOT NULL,  /* JSON */
   `provider_config` TEXT NOT NULL,  /* JSON */
   `version` INT UNSIGNED DEFAULT 0,
   `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
   `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
   PRIMARY KEY ( `pk` ),
   UNIQUE KEY `idx_uk_system_rt_id` (`system_id`, `id`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `bkiam`.`saas_action`(
   `pk` INT UNSIGNED AUTO_INCREMENT,
   `system_id` VARCHAR(32) NOT NULL,
   `id` VARCHAR(32) NOT NULL,
   `name` VARCHAR(255) NOT NULL,
   `name_en` VARCHAR(255) NOT NULL,
   `description` VARCHAR(1024) NOT NULL DEFAULT "",
   `description_en` VARCHAR(1024) NOT NULL DEFAULT "",
   `related_actions` TEXT NOT NULL,  /* JSON */
   `type` VARCHAR(32) NOT NULL DEFAULT "",
   `version` INT UNSIGNED DEFAULT 0,
   `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
   `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
   PRIMARY KEY ( `pk` ),
   UNIQUE KEY `idx_uk_system_action_id` (`system_id`, `id`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `bkiam`.`saas_action_resource_type`(
   `pk` INT UNSIGNED AUTO_INCREMENT,
   `action_system_id` VARCHAR(32) NOT NULL,
   `action_id` VARCHAR(32) NOT NULL,
   `resource_type_system_id` VARCHAR(32) NOT NULL,
   `resource_type_id` VARCHAR(32) NOT NULL,
   `name_alias` VARCHAR(255) NOT NULL DEFAULT "",
   `name_alias_en` VARCHAR(255) NOT NULL DEFAULT "",
   `selection_mode` VARCHAR(32) NOT NULL DEFAULT "",
   `related_instance_selections` TEXT NOT NULL,  /* JSON */
   `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
   `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
   PRIMARY KEY ( `pk` ),
   INDEX `idx_action` (`action_system_id`, `action_id`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `bkiam`.`subject_department`(
   `pk` INT UNSIGNED AUTO_INCREMENT,
   `subject_pk` INT UNSIGNED NOT NULL,
   `department_pks` VARCHAR(1024) NOT NULL,
   `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
   `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
   PRIMARY KEY ( `pk` ),
   UNIQUE KEY `idx_uk_subject` (`subject_pk`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;
