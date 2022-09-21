CREATE INDEX `idx_expire_subject_parent` ON `bkiam`.`subject_relation` (`expired_at`,`subject_pk`,`parent_pk`);
DROP INDEX `idx_subject_parent_expire` ON `bkiam`.`subject_relation`;
