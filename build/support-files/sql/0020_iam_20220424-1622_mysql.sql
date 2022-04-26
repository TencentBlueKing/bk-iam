ALTER TABLE `bkiam`.`saas_action` ADD COLUMN `sensitivity` TinyInt(3) unsigned NOT NULL DEFAULT 0 AFTER `type`;
ALTER TABLE `bkiam`.`saas_action` ADD COLUMN `used_for` varchar(32) NOT NULL DEFAULT 'all' AFTER `sensitivity`;
ALTER TABLE `bkiam`.`saas_resource_type` ADD COLUMN `sensitivity` TinyInt(3) unsigned NOT NULL DEFAULT 0 AFTER `provider_config`
