ALTER TABLE `bkiam`.`saas_action` ADD COLUMN `auth_type` VARCHAR(32) NOT NULL DEFAULT "" AFTER `related_actions`;
