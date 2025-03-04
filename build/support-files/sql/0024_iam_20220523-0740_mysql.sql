CREATE TABLE IF NOT EXISTS `bkiam`.`rbac_group_resource_policy` (
   `pk` int(10) unsigned NOT NULL AUTO_INCREMENT,
   `signature` CHAR(32) NOT NULL,
   `group_pk` int(10) unsigned NOT NULL,
   `template_id` int(10) unsigned NOT NULL,
   `system_id` varchar(32) NOT NULL,
   `action_pks` TEXT NOT NULL,  /* JSON */
   `action_related_resource_type_pk` int(10) unsigned NOT NULL,
   `resource_type_pk` int(10) unsigned NOT NULL,
   `resource_id` varchar(36) NOT NULL,
   `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
   `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
   PRIMARY KEY (`pk`),
   UNIQUE KEY `idx_uk` (`signature`),
   INDEX `idx_resource` (`resource_id`(9), `action_related_resource_type_pk`, `resource_type_pk`, `system_id`),
   INDEX `idx_group_action_resource_type` (`group_pk`,`action_related_resource_type_pk`,`system_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='policy with resource';
