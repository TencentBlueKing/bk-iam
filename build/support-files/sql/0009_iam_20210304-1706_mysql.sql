ALTER TABLE `bkiam`.`expression` MODIFY COLUMN `action_pk` int(10) unsigned NOT NULL DEFAULT 0;
ALTER TABLE `bkiam`.`expression` ADD COLUMN `type` SMALLINT(5) unsigned NOT NULL DEFAULT 0 AFTER `signature`;
UPDATE `bkiam`.`expression` SET `type`=1 WHERE `template_id` != 0;
ALTER TABLE `bkiam`.`expression` DROP INDEX `idx_template_version`;
