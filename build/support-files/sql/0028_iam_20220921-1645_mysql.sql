CREATE INDEX `idx_expire_subject_parent` ON `bkiam`.`subject_relation` (`expired_at`,`subject_pk`,`parent_pk`);
