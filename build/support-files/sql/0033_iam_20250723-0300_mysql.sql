ALTER TABLE `bkiam`.`saas_resource_type` ADD COLUMN `tenant_id` VARCHAR(32) NOT NULL DEFAULT '';
ALTER TABLE `bkiam`.`saas_action` ADD COLUMN `tenant_id` VARCHAR(32) NOT NULL DEFAULT '';
ALTER TABLE `bkiam`.`saas_instance_selection` ADD COLUMN `tenant_id` VARCHAR(32) NOT NULL DEFAULT '';