
CREATE TABLE IF NOT EXISTS `bkiam`.`saas_instance_selection`(
   `pk` INT UNSIGNED AUTO_INCREMENT,
   `system_id` VARCHAR(32) NOT NULL,
   `id` VARCHAR(32) NOT NULL,
   `name` VARCHAR(255) NOT NULL,
   `name_en` VARCHAR(255) NOT NULL,
   `resource_type_chain` TEXT NOT NULL,  /* JSON */
   `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
   `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
   PRIMARY KEY ( `pk` ),
   UNIQUE KEY `idx_uk_system_id` (`system_id`, `id`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;
