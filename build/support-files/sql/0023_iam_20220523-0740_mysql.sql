CREATE TABLE group_resource_policy (
   `pk` int(10) unsigned NOT NULL AUTO_INCREMENT,
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
   UNIQUE KEY `idx_uk_rgtsa` (`resource_id`, `resource_type_pk`, `system_id`, `action_related_resource_type_pk`, `group_pk`, `template_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='policy with resource';