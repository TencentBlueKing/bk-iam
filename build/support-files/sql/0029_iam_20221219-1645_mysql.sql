-- drop index
ALTER TABLE `bkiam`.`subject_role` DROP INDEX `idx_uk_type_system_subject`;

-- create index
CREATE UNIQUE INDEX `idx_uk_subject_system_role_type` ON `bkiam`.`subject_role` (`subject_pk`,`system_id`,`role_type`);
