ALTER TABLE `bkiam`.`expression` ADD INDEX `idx_signature` (`signature`);
ALTER TABLE `bkiam`.`expression` DROP INDEX `idx_signature_action`;
ALTER TABLE `bkiam`.`expression` DROP COLUMN `action_pk`;
ALTER TABLE `bkiam`.`expression` DROP COLUMN `template_id`;
ALTER TABLE `bkiam`.`expression` DROP COLUMN `template_version`;