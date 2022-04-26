ALTER TABLE `bkiam`.`saas_action` ADD COLUMN `sensitivity` TinyInt(3) unsigned NOT NULL DEFAULT 0;
ALTER TABLE `bkiam`.`saas_action` ADD COLUMN `usage` varchar(32) NOT NULL DEFAULT 'all',
ALTER TABLE `bkiam`.`saas_action_resource_type` ADD COLUMN `sensitivity` TinyInt(3) unsigned NOT NULL DEFAULT 0;