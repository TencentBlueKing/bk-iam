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
(105, 'user','user105', 'user105'),
(106, 'user','user106', 'user106'),
(107, 'user','user107', 'user107'),
(2001, 'group','group001', 'group001'),
(2100, 'group','group100', 'group100'),
(2101, 'group','group101', 'group101'),
(2102, 'group','group102', 'group102'),
(2103, 'group','group103', 'group103'),
(2104, 'group','group104', 'group104'),
(2105, 'group','group105', 'group105'),
(2106, 'group','group106', 'group106'),
(2107, 'group','group107', 'group107'),
-- user - dept with group
(120, 'user','user120', 'user120'),
(121, 'user','user121', 'user121'),
(122, 'user','user122', 'user122'),
(123, 'user','user123', 'user123'),
(124, 'user','user124', 'user124'),
(125, 'user','user125', 'user125'),
(126, 'user','user126', 'user126'),
(127, 'user','user127', 'user127'),
(1120, 'department','dept120', 'dept120'),
(1121, 'department','dept121', 'dept121'),
(1122, 'department','dept122', 'dept122'),
(1123, 'department','dept123', 'dept123'),
(1124, 'department','dept124', 'dept124'),
(1125, 'department','dept125', 'dept125'),
(1126, 'department','dept126', 'dept126'),
(1127, 'department','dept127', 'dept127'),
(2120, 'group','group120', 'group120'),
(2121, 'group','group121', 'group121'),
(2122, 'group','group122', 'group122'),
(2123, 'group','group123', 'group123'),
(2124, 'group','group124', 'group124'),
(2125, 'group','group125', 'group125'),
(2126, 'group','group126', 'group126'),
(2127, 'group','group127', 'group127'),
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
(125, '1125','2021-06-21 06:55:24', '2021-06-21 06:55:24'),
(126, '1126','2021-06-21 06:55:24', '2021-06-21 06:55:24'),
(127, '1127','2021-06-21 06:55:24', '2021-06-21 06:55:24'),
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
(105, 2105, 4102444800),
(106, 2106, 4102444800),
(107, 2107, 4102444800),
(1120, 2120, 4102444800),
(1121, 2121, 4102444800),
(1122, 2122, 4102444800),
(1123, 2123, 4102444800),
(1124, 2124, 4102444800),
(1125, 2125, 4102444800),
(1126, 2126, 4102444800),
(1127, 2127, 4102444800),
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
INSERT INTO `action` VALUES
(1,'demo','access_developer_center','2021-06-24 07:23:13','2021-06-24 07:23:13'),
(2,'demo','develop_app','2021-06-24 07:31:28','2021-06-24 07:31:28'),
(3,'demo','view_app','2021-06-24 07:31:28','2021-06-24 07:31:28'),
(4,'demo','edit_app','2021-06-24 07:31:28','2021-06-24 07:31:28');
INSERT INTO `action_resource_type` VALUES
(1,'demo','develop_app','demo','app','2021-06-24 07:31:28','2021-06-24 07:31:28'),
(2,'demo','view_app','demo','app','2021-06-24 07:31:28','2021-06-24 07:31:28'),
(3,'demo','edit_app','demo','app','2021-06-24 07:31:28','2021-06-24 07:31:28');
INSERT INTO `saas_action` VALUES
(1,'demo','access_developer_center','访问开发者中心','access developer center','一个用户是否能访问开发者中心','Is allowed to access the developer center','null','','create', 0, 1,'2021-06-24 07:23:13','2021-06-24 07:23:13','null'),
(2,'demo','develop_app','开发SaaS应用','develop app','一个用户是否能够开发SaaS','Is allowed to develop SaaS app','[\"access_developer_center\"]','','', 0, 1,'2021-06-24 07:31:28','2021-06-24 07:31:28','null'),
(3,'demo','view_app','查看应用','view app','一个用户是否能够查看应用','Is allowed to view app','[\"access_developer_center\"]','rbac','', 0, 1,'2021-06-24 07:31:28','2021-06-24 07:31:28','null'),
(4,'demo','edit_app','编辑应用','view app','一个用户是否能够编辑应用','Is allowed to edit app','[\"access_developer_center\"]','rbac','', 0, 1,'2021-06-24 07:31:28','2021-06-24 07:31:28','null');
INSERT INTO `saas_action_resource_type` VALUES
(1,'demo','develop_app','demo','app','','','instance','[{\"id\":\"app_view\",\"ignore_iam_path\":false,\"system_id\":\"demo\"}]','2021-06-24 07:31:28','2021-06-24 07:31:28'),
(2,'demo','view_app','demo','app','','','instance','[{\"id\":\"app_view\",\"ignore_iam_path\":false,\"system_id\":\"demo\"}]','2021-06-24 07:31:28','2021-06-24 07:31:28'),
(3,'demo','edit_app','demo','app','','','instance','[{\"id\":\"app_view\",\"ignore_iam_path\":false,\"system_id\":\"demo\"}]','2021-06-24 07:31:28','2021-06-24 07:31:28');
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

-- user 106 view_app 001
INSERT INTO `policy` (`subject_pk`, `action_pk`, `expression_pk`, `expired_at`) VALUES (106,3,20,4102444800);
-- user 107 view_app 001
INSERT INTO `policy` (`subject_pk`, `action_pk`, `expression_pk`, `expired_at`) VALUES (107,3,21,4102444800);
-- group 2107 view_app 001
INSERT INTO `policy` (`subject_pk`, `action_pk`, `expression_pk`, `expired_at`) VALUES (2107,3,22,4102444800);
-- department 126 view_app 001
INSERT INTO `policy` (`subject_pk`, `action_pk`, `expression_pk`, `expired_at`) VALUES (126,3,23,4102444800);
-- department 127 view_app 001
INSERT INTO `policy` (`subject_pk`, `action_pk`, `expression_pk`, `expired_at`) VALUES (127,3,24,4102444800);
-- group 2127 view_app 001
INSERT INTO `policy` (`subject_pk`, `action_pk`, `expression_pk`, `expired_at`) VALUES (2127,3,25,4102444800);

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
INSERT INTO `expression` (`pk`, `expression`, `signature`) VALUES (20,
    '[{\"system\": \"demo\", \"type\": \"app\", \"expression\": {\"StringEquals\":{\"id\":[\"001\"]}}}]','EA3BD3486B9ABF5343872EDFB6799F80');
INSERT INTO `expression` (`pk`, `expression`, `signature`) VALUES (21,
    '[{\"system\": \"demo\", \"type\": \"app\", \"expression\": {\"StringEquals\":{\"id\":[\"001\"]}}}]','EA3BD3486B9ABF5343872EDFB6799F80');
INSERT INTO `expression` (`pk`, `expression`, `signature`) VALUES (22,
    '[{\"system\": \"demo\", \"type\": \"app\", \"expression\": {\"StringEquals\":{\"id\":[\"003\"]}}}]','e6fed929a53050eef5a3db05f4eeac0d');
INSERT INTO `expression` (`pk`, `expression`, `signature`) VALUES (23,
    '[{\"system\": \"demo\", \"type\": \"app\", \"expression\": {\"StringEquals\":{\"id\":[\"001\"]}}}]','EA3BD3486B9ABF5343872EDFB6799F80');
INSERT INTO `expression` (`pk`, `expression`, `signature`) VALUES (24,
    '[{\"system\": \"demo\", \"type\": \"app\", \"expression\": {\"StringEquals\":{\"id\":[\"001\"]}}}]','EA3BD3486B9ABF5343872EDFB6799F80');
INSERT INTO `expression` (`pk`, `expression`, `signature`) VALUES (25,
    '[{\"system\": \"demo\", \"type\": \"app\", \"expression\": {\"StringEquals\":{\"id\":[\"003\"]}}}]','e6fed929a53050eef5a3db05f4eeac0d');

-- temporary policy
INSERT INTO `temporary_policy` (`subject_pk`, `action_pk`, `expression`, `expired_at`) VALUES (11,2,
    '[{\"system\": \"demo\", \"type\": \"app\", \"expression\": {\"StringEquals\":{\"id\":[\"001\"]}}}]',4102444800);


-- subject_system_group
INSERT INTO `subject_system_group` (`system_id`, `subject_pk`, `groups`) VALUES ("demo", 102, "{\"2102\":4102444800}");
INSERT INTO `subject_system_group` (`system_id`, `subject_pk`, `groups`) VALUES ("demo", 103, "{\"2103\":4102444800}");
INSERT INTO `subject_system_group` (`system_id`, `subject_pk`, `groups`) VALUES ("demo", 104, "{\"2104\":4102444800}");
INSERT INTO `subject_system_group` (`system_id`, `subject_pk`, `groups`) VALUES ("demo", 141, "{\"2141\":4102444800}");
INSERT INTO `subject_system_group` (`system_id`, `subject_pk`, `groups`) VALUES ("demo", 145, "{\"2145\":4102444800}");
INSERT INTO `subject_system_group` (`system_id`, `subject_pk`, `groups`) VALUES ("demo", 147, "{\"2147\":4102444800}");
INSERT INTO `subject_system_group` (`system_id`, `subject_pk`, `groups`) VALUES ("demo", 1122, "{\"2122\":4102444800}");
INSERT INTO `subject_system_group` (`system_id`, `subject_pk`, `groups`) VALUES ("demo", 1123, "{\"2123\":4102444800}");
INSERT INTO `subject_system_group` (`system_id`, `subject_pk`, `groups`) VALUES ("demo", 1144, "{\"2144\":4102444800}");
INSERT INTO `subject_system_group` (`system_id`, `subject_pk`, `groups`) VALUES ("demo", 1146, "{\"2146\":4102444800}");
INSERT INTO `subject_system_group` (`system_id`, `subject_pk`, `groups`) VALUES ("demo", 1148, "{\"2148\":4102444800}");
INSERT INTO `subject_system_group` (`system_id`, `subject_pk`, `groups`) VALUES ("demo", 105, "{\"2105\":4102444800}");
INSERT INTO `subject_system_group` (`system_id`, `subject_pk`, `groups`) VALUES ("demo", 106, "{\"2106\":4102444800}");
INSERT INTO `subject_system_group` (`system_id`, `subject_pk`, `groups`) VALUES ("demo", 107, "{\"2107\":4102444800}");
INSERT INTO `subject_system_group` (`system_id`, `subject_pk`, `groups`) VALUES ("demo", 1125, "{\"2125\":4102444800}");
INSERT INTO `subject_system_group` (`system_id`, `subject_pk`, `groups`) VALUES ("demo", 1126, "{\"2126\":4102444800}");
INSERT INTO `subject_system_group` (`system_id`, `subject_pk`, `groups`) VALUES ("demo", 1127, "{\"2127\":4102444800}");


-- group_system_auth_type
INSERT INTO `group_system_auth_type` (`system_id`, `group_pk`, `auth_type`) VALUES ("demo", 2102, 1);
INSERT INTO `group_system_auth_type` (`system_id`, `group_pk`, `auth_type`) VALUES ("demo", 2103, 1);
INSERT INTO `group_system_auth_type` (`system_id`, `group_pk`, `auth_type`) VALUES ("demo", 2104, 1);
INSERT INTO `group_system_auth_type` (`system_id`, `group_pk`, `auth_type`) VALUES ("demo", 2141, 1);
INSERT INTO `group_system_auth_type` (`system_id`, `group_pk`, `auth_type`) VALUES ("demo", 2145, 1);
INSERT INTO `group_system_auth_type` (`system_id`, `group_pk`, `auth_type`) VALUES ("demo", 2147, 1);
INSERT INTO `group_system_auth_type` (`system_id`, `group_pk`, `auth_type`) VALUES ("demo", 2122, 1);
INSERT INTO `group_system_auth_type` (`system_id`, `group_pk`, `auth_type`) VALUES ("demo", 2123, 1);
INSERT INTO `group_system_auth_type` (`system_id`, `group_pk`, `auth_type`) VALUES ("demo", 2144, 1);
INSERT INTO `group_system_auth_type` (`system_id`, `group_pk`, `auth_type`) VALUES ("demo", 2146, 1);
INSERT INTO `group_system_auth_type` (`system_id`, `group_pk`, `auth_type`) VALUES ("demo", 2148, 1);
INSERT INTO `group_system_auth_type` (`system_id`, `group_pk`, `auth_type`) VALUES ("demo", 2105, 2);
INSERT INTO `group_system_auth_type` (`system_id`, `group_pk`, `auth_type`) VALUES ("demo", 2106, 2);
INSERT INTO `group_system_auth_type` (`system_id`, `group_pk`, `auth_type`) VALUES ("demo", 2107, 15);
INSERT INTO `group_system_auth_type` (`system_id`, `group_pk`, `auth_type`) VALUES ("demo", 2125, 2);
INSERT INTO `group_system_auth_type` (`system_id`, `group_pk`, `auth_type`) VALUES ("demo", 2126, 2);
INSERT INTO `group_system_auth_type` (`system_id`, `group_pk`, `auth_type`) VALUES ("demo", 2127, 15);


-- group_resource_policy
INSERT INTO `group_resource_policy` (`signature`, `group_pk`, `template_id`, `system_id`, `action_pks`, `action_related_resource_type_pk`, `resource_type_pk`, `resource_id`) VALUES 
("b43770b386441505ad200bf14b85ec44", 2105, 0, "demo", "[3]", 1, 1, "002");
INSERT INTO `group_resource_policy` (`signature`, `group_pk`, `template_id`, `system_id`, `action_pks`, `action_related_resource_type_pk`, `resource_type_pk`, `resource_id`) VALUES 
("ec4c2b529af94e255ab578055ddc91ae", 2106, 0, "demo", "[3]", 1, 1, "002");
INSERT INTO `group_resource_policy` (`signature`, `group_pk`, `template_id`, `system_id`, `action_pks`, `action_related_resource_type_pk`, `resource_type_pk`, `resource_id`) VALUES 
("063890f02170124bad493a0335847275", 2107, 0, "demo", "[3]", 1, 1, "002");
INSERT INTO `group_resource_policy` (`signature`, `group_pk`, `template_id`, `system_id`, `action_pks`, `action_related_resource_type_pk`, `resource_type_pk`, `resource_id`) VALUES 
("7d053f20b1afc58db479f07eaffafd50", 2125, 0, "demo", "[3]", 1, 1, "002");
INSERT INTO `group_resource_policy` (`signature`, `group_pk`, `template_id`, `system_id`, `action_pks`, `action_related_resource_type_pk`, `resource_type_pk`, `resource_id`) VALUES 
("c99c79d3321cb5c289e72b7fbd876758", 2126, 0, "demo", "[3]", 1, 1, "002");
INSERT INTO `group_resource_policy` (`signature`, `group_pk`, `template_id`, `system_id`, `action_pks`, `action_related_resource_type_pk`, `resource_type_pk`, `resource_id`) VALUES 
("e848ca9bc61dfe3dce55a3a17f46e2b9", 2127, 0, "demo", "[3]", 1, 1, "002");
