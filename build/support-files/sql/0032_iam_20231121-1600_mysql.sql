CREATE INDEX `idx_updated_at` ON `bkiam`.`rbac_group_resource_policy` (`updated_at`);
CREATE INDEX `idx_resource_type` ON `bkiam`.`rbac_group_resource_policy` (`resource_type_pk`,`action_related_resource_type_pk`,`system_id`,`resource_id`);
DROP INDEX `idx_resource` ON `bkiam`.`rbac_group_resource_policy`;