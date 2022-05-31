/*  subject: user - department - groups*/

/* NOTE: new case should build there own subject-dept-group relations/ploicy/expression, should not use old case's data */

-- the PK rule:
-- single user 1-99
-- user with dept or group 100-999
-- department 1xxx
-- group      2xxx

INSERT INTO `subject` (`pk`, `type`, `id`, `name`) VALUES
-- admin
(99999, 'user','admin', '管理员'),
-- user001 - no permissions
(1, 'user','user001', 'user001'),
-- user002/003/004/005 - has systemA permissions
(2, 'user','user002', 'user002'),
(3, 'user','user003', 'user003'),
(4, 'user','user004', 'user004'),
(5, 'user','user005', 'user005'),
(6, 'user','user006', 'user006'),
(7, 'user','user007', 'user007'),
(8, 'user','user008', 'user008'),
(9, 'user','user009', 'user009'),
(10, 'user','user010', 'user010'),
(11, 'user','user011', 'user011'),
-- user with group
(100, 'user','user100', 'user100'),
(101, 'user','user101', 'user101'),
(102, 'user','user102', 'user102'),
(103, 'user','user103', 'user103'),
(104, 'user','user104', 'user104'),
(2001, 'group','group001', 'group001'),
(2100, 'group','group100', 'group100'),
(2101, 'group','group101', 'group101'),
(2102, 'group','group102', 'group102'),
(2103, 'group','group103', 'group103'),
(2104, 'group','group104', 'group104'),
-- user - dept with group
(120, 'user','user120', 'user120'),
(121, 'user','user121', 'user121'),
(122, 'user','user122', 'user122'),
(123, 'user','user123', 'user123'),
(124, 'user','user124', 'user124'),
(1120, 'department','dept120', 'dept120'),
(1121, 'department','dept121', 'dept121'),
(1122, 'department','dept122', 'dept122'),
(1123, 'department','dept123', 'dept123'),
(1124, 'department','dept124', 'dept124'),
(2120, 'group','group120', 'group120'),
(2121, 'group','group121', 'group121'),
(2122, 'group','group122', 'group122'),
(2123, 'group','group123', 'group123'),
(2124, 'group','group124', 'group124'),
-- user - group / user - dept - group
-- case 1: user140 - dept  1140 - group 2140
-- case 2: user141 - group 2141  / user141 - dept 1142 - group 2142
-- case 3: user143 - group 2143  / user143 - dept 1144 - group 2144
-- case 4: user145 - group 2145  / user145 - dept 1146 - group 2146
-- case 5: user147 - group 2147  / user147 - dept 1148 - group 2148
(140, 'user','user140', 'user140'),
(141, 'user','user141', 'user141'),
(143, 'user','user143', 'user143'),
(145, 'user','user145', 'user145'),
(147, 'user','user147', 'user147'),
(1140, 'department','dept140', 'dept140'),
(1142, 'department','dept142', 'dept142'),
(1144, 'department','dept144', 'dept144'),
(1146, 'department','dept146', 'dept146'),
(1148, 'department','dept148', 'dept148'),
(2140, 'group','group140', 'group140'),
(2141, 'group','group141', 'group141'),
(2142, 'group','group142', 'group142'),
(2143, 'group','group143', 'group143'),
(2144, 'group','group144', 'group144'),
(2145, 'group','group145', 'group145'),
(2146, 'group','group146', 'group146'),
(2147, 'group','group147', 'group147'),
(2148, 'group','group148', 'group148');

/* sujbect-departments */
INSERT INTO `subject_department` (`subject_pk`, `department_pks`, `created_at`, `updated_at`) VALUES
(120, '1120','2021-06-21 06:55:24', '2021-06-21 06:55:24'),
(121, '1121','2021-06-21 06:55:24', '2021-06-21 06:55:24'),
(122, '1122','2021-06-21 06:55:24', '2021-06-21 06:55:24'),
(123, '1123','2021-06-21 06:55:24', '2021-06-21 06:55:24'),
(124, '1124','2021-06-21 06:55:24', '2021-06-21 06:55:24'),
-- user - group / user - dept - group
-- case 1: user140 - dept  1140 - group 2140
-- case 2: user141 - group 2141  / user141 - dept 1142 - group 2142
-- case 3: user143 - group 2143  / user143 - dept 1144 - group 2144
-- case 4: user145 - group 2145  / user145 - dept 1146 - group 2146
-- case 5: user147 - group 2147  / user147 - dept 1148 - group 2148
(140, '1140','2021-06-21 06:55:24', '2021-06-21 06:55:24'),
(141, '1142','2021-06-21 06:55:24', '2021-06-21 06:55:24'),
(143, '1144','2021-06-21 06:55:24', '2021-06-21 06:55:24'),
(145, '1146','2021-06-21 06:55:24', '2021-06-21 06:55:24'),
(147, '1148','2021-06-21 06:55:24', '2021-06-21 06:55:24');


/* subject-groups: user-group  department-group */
INSERT INTO `subject_relation` (`subject_pk`, `parent_pk`, `policy_expired_at`) VALUES
(100, 2100, 4102444800),
(101, 2101, 4102444800),
(102, 2102, 4102444800),
(103, 2103, 4102444800),
(104, 2104, 4102444800),
(1120, 2120, 4102444800),
(1121, 2121, 4102444800),
(1122, 2122, 4102444800),
(1123, 2123, 4102444800),
(1124, 2124, 4102444800),
-- user - group / user - dept - group
-- case 1: user140 - dept  1140 - group 2140
-- case 2: user141 - group 2141  / user141 - dept 1142 - group 2142
-- case 3: user143 - group 2143  / user143 - dept 1144 - group 2144
-- case 4: user145 - group 2145  / user145 - dept 1146 - group 2146
-- case 5: user147 - group 2147  / user147 - dept 1148 - group 2148
(1140, 2140, 4102444800),
(1142, 2142, 4102444800),
(1144, 2144, 4102444800),
(141, 2141, 4102444800),
(143, 2143, 4102444800),
(145, 2145, 4102444800),
(1146, 2146, 4102444800),
(147, 2147, 4102444800),
(1148, 2148, 4102444800);



/* base system info: demo */
INSERT INTO `system_info` VALUES ('demo','2021-06-24 07:23:13','2021-06-24 07:23:13');
INSERT INTO `resource_type` VALUES (1,'demo','app','2021-06-24 07:23:13','2021-06-24 07:23:13');
INSERT INTO `action` VALUES (1,'demo','access_developer_center','2021-06-24 07:23:13','2021-06-24 07:23:13'),(2,'demo','develop_app','2021-06-24 07:31:28','2021-06-24 07:31:28');
INSERT INTO `action_resource_type` VALUES (1,'demo','develop_app','demo','app','2021-06-24 07:31:28','2021-06-24 07:31:28');
INSERT INTO `saas_action` VALUES (1,'demo','access_developer_center','访问开发者中心','access developer center','一个用户是否能访问开发者中心','Is allowed to access the developer center','null','create', 0, 1,'2021-06-24 07:23:13','2021-06-24 07:23:13','null'),(2,'demo','develop_app','开发SaaS应用','develop app','一个用户是否能够开发SaaS','Is allowed to develop SaaS app','[\"access_developer_center\"]','', 0, 1,'2021-06-24 07:31:28','2021-06-24 07:31:28','null');
INSERT INTO `saas_action_resource_type` VALUES (1,'demo','develop_app','demo','app','','','instance','[{\"id\":\"app_view\",\"ignore_iam_path\":false,\"system_id\":\"demo\"}]','2021-06-24 07:31:28','2021-06-24 07:31:28');
INSERT INTO `saas_instance_selection` VALUES (1,'demo','app_view','应用视图','app_view',0,'[{\"system_id\":\"demo\",\"id\":\"app\"}]','2021-06-24 07:23:13','2021-06-24 07:23:13');
INSERT INTO `saas_resource_type` VALUES (1,'demo','app','SaaS应用','application','SaaS应用','SaaS application','[]','{\"path\":\"/api/v1/iam/apps\"}',0,1,'2021-06-24 07:23:13','2021-06-24 07:31:28');
INSERT INTO `saas_system_info` VALUES ('demo','Demo平台','Demo','A demo SaaS for quick start','A demo SaaS for quick start.','demo,bk_iam_app','{\"token\":\"63yr6hs11bsqa8u4d9i0acbpjuuyizaw\",\"host\":\"http://127.0.0.1:5000\",\"auth\":\"basic\",\"healthz\":\"/healthz/\"}','2021-06-24 07:23:13','2021-06-24 07:31:28');

INSERT INTO `system_info` VALUES ('demo2','2021-06-24 07:23:13','2021-06-24 07:23:13');
INSERT INTO `saas_system_info` VALUES ('demo2','Demo平台2','Demo','A demo SaaS for quick start','A demo SaaS for quick start.','demo,bk_iam_app','{\"token\":\"63yr6hs11bsqa8u4d9i0acbpjuuyizaw\",\"host\":\"http://127.0.0.1:5000\",\"auth\":\"basic\",\"healthz\":\"/healthz/\"}','2021-06-24 07:23:13','2021-06-24 07:31:28');

/* base policy and expressions */
-- actionPK
--    access_developer_center=1
--    develop_app=2
-- policy

-- user002 access_developer_center
INSERT INTO `policy` (`subject_pk`, `action_pk`, `expression_pk`, `expired_at`) VALUES (2,1,-1,4102444800);
-- user003 develop_app any
INSERT INTO `policy` (`subject_pk`, `action_pk`, `expression_pk`, `expired_at`) VALUES (3,2,1,4102444800);
-- user004 develop_app id=001
INSERT INTO `policy` (`subject_pk`, `action_pk`, `expression_pk`, `expired_at`) VALUES (4,2,2,4102444800);
-- user005 access_developer_center expired / develop_app expired
INSERT INTO `policy` (`subject_pk`, `action_pk`, `expression_pk`, `expired_at`) VALUES (5,1,-1,1624521014);
INSERT INTO `policy` (`subject_pk`, `action_pk`, `expression_pk`, `expired_at`) VALUES (5,2,3,1624521014);
-- user101  user 1, group 0
INSERT INTO `policy` (`subject_pk`, `action_pk`, `expression_pk`, `expired_at`) VALUES (101,2,4,4102444800);
-- user102  user 0, group 1
INSERT INTO `policy` (`subject_pk`, `action_pk`, `expression_pk`, `expired_at`) VALUES (2102,2,5,4102444800);
-- user103  user 1, group 1
INSERT INTO `policy` (`subject_pk`, `action_pk`, `expression_pk`, `expired_at`) VALUES (103,2,6,4102444800);
INSERT INTO `policy` (`subject_pk`, `action_pk`, `expression_pk`, `expired_at`) VALUES (2103,2,7,4102444800);
-- user104  user 0, group 1 but expired
INSERT INTO `policy` (`subject_pk`, `action_pk`, `expression_pk`, `expired_at`) VALUES (2104,2,8,1624521014);
-- user121  user 1, dept-group 0
INSERT INTO `policy` (`subject_pk`, `action_pk`, `expression_pk`, `expired_at`) VALUES (121,2,9,4102444800);
-- user122  user 0, dept-group 1
INSERT INTO `policy` (`subject_pk`, `action_pk`, `expression_pk`, `expired_at`) VALUES (2122,2,10,4102444800);
-- user 123 user 1, dept-group 1
INSERT INTO `policy` (`subject_pk`, `action_pk`, `expression_pk`, `expired_at`) VALUES (123,2,11,4102444800);
INSERT INTO `policy` (`subject_pk`, `action_pk`, `expression_pk`, `expired_at`) VALUES (2123,2,12,4102444800);
-- user - group / user - dept - group
-- case 1: user140 - dept  1140 - group 2140 => 都无
-- case 2: user141 - group 2141  / user141 - dept 1142 - group 2142  => 2141有
-- case 3: user143 - group 2143  / user143 - dept 1144 - group 2144  => 2144有
-- case 4: user145 - group 2145  / user145 - dept 1146 - group 2146  => 2145 / 2146 有
-- case 5: user147 - group 2147  / user147 - dept 1148 - group 2148  => 147 / 2147 / 2148 有
INSERT INTO `policy` (`subject_pk`, `action_pk`, `expression_pk`, `expired_at`) VALUES (2141,2,13,4102444800);
INSERT INTO `policy` (`subject_pk`, `action_pk`, `expression_pk`, `expired_at`) VALUES (2144,2,14,4102444800);
INSERT INTO `policy` (`subject_pk`, `action_pk`, `expression_pk`, `expired_at`) VALUES (2145,2,15,4102444800);
INSERT INTO `policy` (`subject_pk`, `action_pk`, `expression_pk`, `expired_at`) VALUES (2146,2,16,4102444800);
INSERT INTO `policy` (`subject_pk`, `action_pk`, `expression_pk`, `expired_at`) VALUES (147,2,17,4102444800);
INSERT INTO `policy` (`subject_pk`, `action_pk`, `expression_pk`, `expired_at`) VALUES (2147,2,18,4102444800);
INSERT INTO `policy` (`subject_pk`, `action_pk`, `expression_pk`, `expired_at`) VALUES (2148,2,19,4102444800);


-- expression, not the signature should be not be the same
INSERT INTO `expression` (`pk`, `expression`, `signature`) VALUES (1,
    '[{\"system\": \"demo\", \"type\": \"app\", \"expression\": {\"Any\":{\"id\":[]}}}]','57479006515B301E80954B21E259B7BE');
INSERT INTO `expression` (`pk`, `expression`, `signature`) VALUES (2,
    '[{\"system\": \"demo\", \"type\": \"app\", \"expression\": {\"StringEquals\":{\"id\":[\"001\"]}}}]','EA3BD3486B9ABF5343872EDFB6799F80');
INSERT INTO `expression` (`pk`, `expression`, `signature`) VALUES (3,
    '[{\"system\": \"demo\", \"type\": \"app\", \"expression\": {\"StringEquals\":{\"id\":[\"001\"]}}}]','EA3BD3486B9ABF5343872EDFB6799F80');
INSERT INTO `expression` (`pk`, `expression`, `signature`) VALUES (4,
    '[{\"system\": \"demo\", \"type\": \"app\", \"expression\": {\"StringEquals\":{\"id\":[\"001\"]}}}]','EA3BD3486B9ABF5343872EDFB6799F80');
INSERT INTO `expression` (`pk`, `expression`, `signature`) VALUES (5,
    '[{\"system\": \"demo\", \"type\": \"app\", \"expression\": {\"StringEquals\":{\"id\":[\"001\"]}}}]','EA3BD3486B9ABF5343872EDFB6799F80');
INSERT INTO `expression` (`pk`, `expression`, `signature`) VALUES (6,
    '[{\"system\": \"demo\", \"type\": \"app\", \"expression\": {\"StringEquals\":{\"id\":[\"001\"]}}}]','EA3BD3486B9ABF5343872EDFB6799F80');
INSERT INTO `expression` (`pk`, `expression`, `signature`) VALUES (7,
    '[{\"system\": \"demo\", \"type\": \"app\", \"expression\": {\"StringEquals\":{\"id\":[\"001\"]}}}]','EA3BD3486B9ABF5343872EDFB6799F80');
INSERT INTO `expression` (`pk`, `expression`, `signature`) VALUES (8,
    '[{\"system\": \"demo\", \"type\": \"app\", \"expression\": {\"StringEquals\":{\"id\":[\"001\"]}}}]','EA3BD3486B9ABF5343872EDFB6799F80');
INSERT INTO `expression` (`pk`, `expression`, `signature`) VALUES (9,
    '[{\"system\": \"demo\", \"type\": \"app\", \"expression\": {\"StringEquals\":{\"id\":[\"001\"]}}}]','EA3BD3486B9ABF5343872EDFB6799F80');
INSERT INTO `expression` (`pk`, `expression`, `signature`) VALUES (10,
    '[{\"system\": \"demo\", \"type\": \"app\", \"expression\": {\"StringEquals\":{\"id\":[\"001\"]}}}]','EA3BD3486B9ABF5343872EDFB6799F80');
INSERT INTO `expression` (`pk`, `expression`, `signature`) VALUES (11,
    '[{\"system\": \"demo\", \"type\": \"app\", \"expression\": {\"StringEquals\":{\"id\":[\"001\"]}}}]','EA3BD3486B9ABF5343872EDFB6799F80');
INSERT INTO `expression` (`pk`, `expression`, `signature`) VALUES (12,
    '[{\"system\": \"demo\", \"type\": \"app\", \"expression\": {\"StringEquals\":{\"id\":[\"001\"]}}}]','EA3BD3486B9ABF5343872EDFB6799F80');
INSERT INTO `expression` (`pk`, `expression`, `signature`) VALUES (13,
    '[{\"system\": \"demo\", \"type\": \"app\", \"expression\": {\"StringEquals\":{\"id\":[\"001\"]}}}]','EA3BD3486B9ABF5343872EDFB6799F80');
INSERT INTO `expression` (`pk`, `expression`, `signature`) VALUES (14,
    '[{\"system\": \"demo\", \"type\": \"app\", \"expression\": {\"StringEquals\":{\"id\":[\"001\"]}}}]','EA3BD3486B9ABF5343872EDFB6799F80');
INSERT INTO `expression` (`pk`, `expression`, `signature`) VALUES (15,
    '[{\"system\": \"demo\", \"type\": \"app\", \"expression\": {\"StringEquals\":{\"id\":[\"001\"]}}}]','EA3BD3486B9ABF5343872EDFB6799F80');
INSERT INTO `expression` (`pk`, `expression`, `signature`) VALUES (16,
    '[{\"system\": \"demo\", \"type\": \"app\", \"expression\": {\"StringEquals\":{\"id\":[\"001\"]}}}]','EA3BD3486B9ABF5343872EDFB6799F80');
INSERT INTO `expression` (`pk`, `expression`, `signature`) VALUES (17,
    '[{\"system\": \"demo\", \"type\": \"app\", \"expression\": {\"StringEquals\":{\"id\":[\"001\"]}}}]','EA3BD3486B9ABF5343872EDFB6799F80');
INSERT INTO `expression` (`pk`, `expression`, `signature`) VALUES (18,
    '[{\"system\": \"demo\", \"type\": \"app\", \"expression\": {\"StringEquals\":{\"id\":[\"001\"]}}}]','EA3BD3486B9ABF5343872EDFB6799F80');
INSERT INTO `expression` (`pk`, `expression`, `signature`) VALUES (19,
    '[{\"system\": \"demo\", \"type\": \"app\", \"expression\": {\"StringEquals\":{\"id\":[\"001\"]}}}]','EA3BD3486B9ABF5343872EDFB6799F80');


-- temporary policy
INSERT INTO `temporary_policy` (`subject_pk`, `action_pk`, `expression`, `expired_at`) VALUES (11,2,
    '[{\"system\": \"demo\", \"type\": \"app\", \"expression\": {\"StringEquals\":{\"id\":[\"001\"]}}}]',4102444800);
