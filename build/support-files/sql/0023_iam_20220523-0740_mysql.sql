CREATE TABLE group_resource_policy (
   `pk` int(10) unsigned NOT NULL AUTO_INCREMENT,
   `signature` CHAR(32) NOT NULL,
   `group_pk` int(10) unsigned NOT NULL,
   `template_id` int(10) unsigned NOT NULL,
   `system_id` varchar(32) NOT NULL,
   `action_pks` TEXT NOT NULL,  /* JSON */
   `action_related_resource_type_pk` int(10) unsigned NOT NULL,
   `resource_type_pk` int(10) unsigned NOT NULL,
   `resource_id` varchar(32) NOT NULL,
   `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
   `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
   PRIMARY KEY (`pk`),
   UNIQUE KEY `idx_uk` (`signature`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='policy with resource';