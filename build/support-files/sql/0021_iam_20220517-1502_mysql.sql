-- drop index
ALTER TABLE `bkiam`.`subject_relation` DROP INDEX idx_uk_subject_parent;
ALTER TABLE `bkiam`.`subject_relation` DROP INDEX idx_subject;
ALTER TABLE `bkiam`.`subject_relation` DROP INDEX idx_parent;
ALTER TABLE `bkiam`.`subject_relation` DROP INDEX idx_subject_pk_expire;

-- drop column
ALTER TABLE `bkiam`.`subject_relation` DROP COLUMN `subject_type`;
ALTER TABLE `bkiam`.`subject_relation` DROP COLUMN `subject_id`;
ALTER TABLE `bkiam`.`subject_relation` DROP COLUMN `parent_type`;
ALTER TABLE `bkiam`.`subject_relation` DROP COLUMN `parent_id`;

-- create index
CREATE UNIQUE INDEX idx_uk_parent_subject ON `bkiam`.`subject_relation` (`parent_pk`,`subject_pk`);
CREATE INDEX idx_subject_parent_expire ON `bkiam`.`subject_relation` (`subject_pk`,`parent_pk`,`policy_expired_at`);
