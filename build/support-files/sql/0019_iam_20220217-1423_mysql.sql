CREATE TABLE `bkiam`.`temporary_policy` (
  `pk` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `subject_pk` int(10) unsigned NOT NULL,
  `action_pk` int(10) unsigned NOT NULL,
  `expression` mediumtext NOT NULL,
  `expired_at` int(10) unsigned NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`pk`),
  KEY `idx_expire` (`expired_at`),
  KEY `idx_subject_action_expire` (`subject_pk`,`action_pk`,`expired_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
