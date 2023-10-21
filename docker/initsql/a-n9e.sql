set names utf8mb4;

drop database if exists n9e_v6;
create database n9e_v6;
use n9e_v6;

CREATE TABLE `users`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `username` varchar(64)  NOT NULL COMMENT 'login name, cannot rename',
  `nickname` varchar(64)  NOT NULL COMMENT 'display name, chinese name',
  `password` varchar(128)  NOT NULL DEFAULT '',
  `phone` varchar(16)  NOT NULL DEFAULT '',
  `email` varchar(64)  NOT NULL DEFAULT '',
  `portrait` varchar(255)  NOT NULL DEFAULT '' COMMENT 'portrait image url',
  `roles` varchar(255)  NOT NULL COMMENT 'Admin | Standard | Guest, split by space',
  `contacts` varchar(1024)  NULL DEFAULT NULL COMMENT 'json e.g. {wecom:xx, dingtalk_robot_token:yy}',
  `maintainer` tinyint(1) NOT NULL DEFAULT 0,
  `create_at` bigint NOT NULL DEFAULT 0,
  `create_by` varchar(64)  NOT NULL DEFAULT '',
  `update_at` bigint NOT NULL DEFAULT 0,
  `update_by` varchar(64)  NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY (`username`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4;

insert into `users`(id, username, nickname, password, roles, create_at, create_by, update_at, update_by) values(1, 'root', '超管', 'root.2020', 'Admin', unix_timestamp(now()), 'system', unix_timestamp(now()), 'system');

CREATE TABLE `user_group`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` varchar(128)  NOT NULL DEFAULT '',
  `note` varchar(255)  NOT NULL DEFAULT '',
  `create_at` bigint NOT NULL DEFAULT 0,
  `create_by` varchar(64)  NOT NULL DEFAULT '',
  `update_at` bigint NOT NULL DEFAULT 0,
  `update_by` varchar(64)  NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY (`create_by`),
  KEY (`update_at`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4;

insert into user_group(id, name, create_at, create_by, update_at, update_by) values(1, 'demo-root-group', unix_timestamp(now()), 'root', unix_timestamp(now()), 'root');

CREATE TABLE `user_group_member`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `group_id` bigint UNSIGNED NOT NULL,
  `user_id` bigint UNSIGNED NOT NULL,
  PRIMARY KEY (`id`),
  KEY (`group_id`),
  KEY (`user_id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4;

insert into user_group_member(group_id, user_id) values(1, 1);

CREATE TABLE `configs`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `ckey` varchar(191)  NOT NULL,
  `cval` text  NULL COMMENT 'config value',
  `note` varchar(1024)  NULL DEFAULT '' COMMENT 'note',
  `external` bigint NULL DEFAULT 0 COMMENT '0:built-in 1:external',
  `encrypted` bigint NULL DEFAULT 0 COMMENT '0:plaintext 1:ciphertext',
  `create_at` bigint NULL DEFAULT 0 COMMENT 'create_at',
  `create_by` varchar(64)  NULL DEFAULT '' COMMENT 'cerate_by',
  `update_at` bigint NULL DEFAULT 0 COMMENT 'update_at',
  `update_by` varchar(64)  NULL DEFAULT '' COMMENT 'update_by',
  PRIMARY KEY (`id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4;

CREATE TABLE `role`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` varchar(191)  NOT NULL DEFAULT '',
  `note` varchar(255)  NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY (`name`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4;

insert into `role`(name, note) values('Admin', 'Administrator role');
insert into `role`(name, note) values('Standard', 'Ordinary user role');
insert into `role`(name, note) values('Guest', 'Readonly user role');

CREATE TABLE `role_operation`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `role_name` varchar(128)  NOT NULL,
  `operation` varchar(191)  NOT NULL,
  PRIMARY KEY (`id`),
  KEY (`role_name`),
  KEY (`operation`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4;

-- Admin is special, who has no concrete operation but can do anything.
INSERT INTO `role_operation`(role_name, operation) VALUES ('Guest', '/metric/explorer');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Guest', '/object/explorer');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Guest', '/log/explorer');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Guest', '/trace/explorer');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Guest', '/help/version');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Guest', '/help/contact');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/metric/explorer');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/object/explorer');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/log/explorer');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/trace/explorer');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/help/version');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/help/contact');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/help/servers');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/help/migrate');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/alert-rules-built-in');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/dashboards-built-in');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/trace/dependencies');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Admin', '/help/source');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Admin', '/help/sso');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Admin', '/help/notification-tpls');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Admin', '/help/notification-settings');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/users');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/user-groups');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/user-groups/add');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/user-groups/put');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/user-groups/del');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/busi-groups');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/busi-groups/add');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/busi-groups/put');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/busi-groups/del');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/targets');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/targets/add');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/targets/put');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/targets/del');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/dashboards');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/dashboards/add');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/dashboards/put');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/dashboards/del');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/alert-rules');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/alert-rules/add');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/alert-rules/put');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/alert-rules/del');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/alert-mutes');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/alert-mutes/add');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/alert-mutes/del');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/alert-subscribes');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/alert-subscribes/add');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/alert-subscribes/put');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/alert-subscribes/del');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/alert-cur-events');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/alert-cur-events/del');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/alert-his-events');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/job-tpls');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/job-tpls/add');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/job-tpls/put');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/job-tpls/del');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/job-tasks');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/job-tasks/add');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/job-tasks/put');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/recording-rules');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/recording-rules/add');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/recording-rules/put');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/recording-rules/del');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/alert-mutes/put');
INSERT INTO `role_operation`(role_name, operation) VALUES ('Standard', '/log/index-patterns');

-- for alert_rule | collect_rule | mute | dashboard grouping
CREATE TABLE `busi_group`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` varchar(191)  NOT NULL,
  `label_enable` tinyint(1) NOT NULL DEFAULT 0,
  `label_value` varchar(191)  NOT NULL DEFAULT '' COMMENT 'if label_enable: label_value can not be blank',
  `create_at` bigint NOT NULL DEFAULT 0,
  `create_by` varchar(64)  NOT NULL DEFAULT '',
  `update_at` bigint NOT NULL DEFAULT 0,
  `update_by` varchar(64)  NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY (`name`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4;

insert into busi_group(id, name, create_at, create_by, update_at, update_by) values(1, 'Default Busi Group', unix_timestamp(now()), 'root', unix_timestamp(now()), 'root');

CREATE TABLE `busi_group_member`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `busi_group_id` bigint NOT NULL COMMENT 'busi group id',
  `user_group_id` bigint NOT NULL COMMENT 'user group id',
  `perm_flag` char(2)  NOT NULL COMMENT 'ro | rw',
  PRIMARY KEY (`id`),
  KEY (`busi_group_id`),
  KEY (`user_group_id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4;

insert into busi_group_member(busi_group_id, user_group_id, perm_flag) values(1, 1, 'rw');

-- for dashboard new version
CREATE TABLE `board`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `group_id` bigint NOT NULL DEFAULT 0 COMMENT 'busi group id',
  `name` varchar(191)  NOT NULL,
  `ident` varchar(200)  NOT NULL DEFAULT '',
  `tags` varchar(255)  NOT NULL COMMENT 'split by space',
  `public` tinyint(1) NOT NULL DEFAULT 0 COMMENT '0:false 1:true',
  `built_in` tinyint(1) NOT NULL DEFAULT 0 COMMENT '0:false 1:true',
  `hide` tinyint(1) NOT NULL DEFAULT 0 COMMENT '0:false 1:true',
  `create_at` bigint NOT NULL DEFAULT 0,
  `create_by` varchar(64)  NOT NULL DEFAULT '',
  `update_at` bigint NOT NULL DEFAULT 0,
  `update_by` varchar(64)  NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY (`group_id`, `name`),
  KEY(`ident`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4;

-- for dashboard new version
CREATE TABLE `board_payload`  (
  `id` bigint UNSIGNED NOT NULL COMMENT 'dashboard id',
  `payload` mediumtext  NOT NULL,
  UNIQUE KEY (`id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4;

-- deprecated
CREATE TABLE `dashboard`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `group_id` bigint NOT NULL DEFAULT 0 COMMENT 'busi group id',
  `name` varchar(191)  NOT NULL,
  `tags` varchar(255)  NOT NULL COMMENT 'split by space',
  `configs` varchar(8192)  NULL DEFAULT NULL COMMENT 'dashboard variables',
  `create_at` bigint NOT NULL DEFAULT 0,
  `create_by` varchar(64)  NOT NULL DEFAULT '',
  `update_at` bigint NOT NULL DEFAULT 0,
  `update_by` varchar(64)  NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY (`group_id`, `name`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4;

-- deprecated
-- auto create the first subclass 'Default chart group' of dashboard
CREATE TABLE `chart_group`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `dashboard_id` bigint UNSIGNED NOT NULL,
  `name` varchar(255)  NOT NULL,
  `weight` int NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY (`dashboard_id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4;

-- deprecated
CREATE TABLE `chart`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `group_id` bigint UNSIGNED NOT NULL COMMENT 'chart group id',
  `configs` text  NULL,
  `weight` int NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY (`group_id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4;

CREATE TABLE `chart_share`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `cluster` varchar(128)  NOT NULL,
  `datasource_id` bigint NOT NULL DEFAULT 0 COMMENT 'datasource id',
  `configs` text  NULL,
  `create_at` bigint NOT NULL DEFAULT 0,
  `create_by` varchar(64)  NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  key (`create_at`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4;

CREATE TABLE `alert_rule`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `group_id` bigint NOT NULL DEFAULT 0 COMMENT 'busi group id',
  `cate` varchar(128)  NOT NULL,
  `datasource_ids` varchar(255)  NOT NULL DEFAULT '' COMMENT 'datasource ids',
  `cluster` varchar(128)  NOT NULL,
  `name` varchar(255)  NOT NULL,
  `note` varchar(1024)  NOT NULL DEFAULT '',
  `prod` varchar(255)  NOT NULL DEFAULT '',
  `algorithm` varchar(255)  NOT NULL DEFAULT '',
  `algo_params` varchar(255)  NULL DEFAULT NULL,
  `delay` int NOT NULL DEFAULT 0,
  `severity` tinyint(1) NOT NULL COMMENT '1:Emergency 2:Warning 3:Notice',
  `disabled` tinyint(1) NOT NULL COMMENT '0:enabled 1:disabled',
  `prom_for_duration` int NOT NULL COMMENT 'prometheus for, unit:s',
  `rule_config` text  NOT NULL COMMENT 'rule_config',
  `prom_ql` text  NOT NULL COMMENT 'promql',
  `prom_eval_interval` int NOT NULL COMMENT 'evaluate interval',
  `enable_stime` varchar(255)  NOT NULL DEFAULT '00:00',
  `enable_etime` varchar(255)  NOT NULL DEFAULT '23:59',
  `enable_days_of_week` varchar(255)  NOT NULL DEFAULT '' COMMENT 'split by space: 0 1 2 3 4 5 6',
  `enable_in_bg` tinyint(1) NOT NULL DEFAULT 0 COMMENT '1: only this bg 0: global',
  `notify_recovered` tinyint(1) NOT NULL COMMENT 'whether notify when recovery',
  `notify_channels` varchar(255)  NOT NULL DEFAULT '' COMMENT 'split by space: sms voice email dingtalk wecom',
  `notify_groups` varchar(255)  NOT NULL DEFAULT '' COMMENT 'split by space: 233 43',
  `notify_repeat_step` int NOT NULL DEFAULT 0 COMMENT 'unit: min',
  `notify_max_number` int NOT NULL DEFAULT 0,
  `recover_duration` int NOT NULL DEFAULT 0 COMMENT 'unit: s',
  `callbacks` varchar(255)  NOT NULL DEFAULT '' COMMENT 'split by space: http://a.com/api/x http://a.com/api/y',
  `runbook_url` varchar(255)  NULL DEFAULT NULL,
  `append_tags` varchar(255)  NOT NULL DEFAULT '' COMMENT 'split by space: service=n9e mod=api',
  `annotations` text  NOT NULL COMMENT 'annotations',
  `extra_config` text  NULL,
  `create_at` bigint NOT NULL DEFAULT 0,
  `create_by` varchar(64)  NOT NULL DEFAULT '',
  `update_at` bigint NOT NULL DEFAULT 0,
  `update_by` varchar(64)  NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY (`group_id`),
  KEY (`update_at`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4;

CREATE TABLE `alert_mute`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `group_id` bigint NOT NULL DEFAULT 0 COMMENT 'busi group id',
  `prod` varchar(255)  NOT NULL DEFAULT '',
  `note` varchar(1024)  NOT NULL DEFAULT '',
  `cate` varchar(128)  NOT NULL,
  `cluster` varchar(128)  NOT NULL,
  `datasource_ids` varchar(255)  NOT NULL DEFAULT '' COMMENT 'datasource ids',
  `tags` varchar(4096)  NOT NULL DEFAULT '' COMMENT 'json,map,tagkey->regexp|value',
  `cause` varchar(255)  NOT NULL DEFAULT '',
  `btime` bigint NOT NULL DEFAULT 0 COMMENT 'begin time',
  `etime` bigint NOT NULL DEFAULT 0 COMMENT 'end time',
  `disabled` tinyint(1) NOT NULL DEFAULT 0 COMMENT '0:enabled 1:disabled',
  `mute_time_type` tinyint(1) NOT NULL DEFAULT 0,
  `periodic_mutes` varchar(4096)  NOT NULL DEFAULT '',
  `severities` varchar(32)  NOT NULL DEFAULT '',
  `create_at` bigint NOT NULL DEFAULT 0,
  `create_by` varchar(64)  NOT NULL DEFAULT '',
  `update_at` bigint NOT NULL DEFAULT 0,
  `update_by` varchar(64)  NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY (`create_at`),
  KEY (`group_id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4;

CREATE TABLE `alert_subscribe`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` varchar(255)  NOT NULL DEFAULT '',
  `disabled` tinyint(1) NOT NULL DEFAULT 0 COMMENT '0:enabled 1:disabled',
  `group_id` bigint NOT NULL DEFAULT 0 COMMENT 'busi group id',
  `prod` varchar(255)  NOT NULL DEFAULT '',
  `cate` varchar(128)  NOT NULL,
  `datasource_ids` varchar(255)  NOT NULL DEFAULT '' COMMENT 'datasource ids',
  `cluster` varchar(128)  NOT NULL,
  `rule_id` bigint NOT NULL DEFAULT 0,
  `severities` varchar(32)  NOT NULL DEFAULT '',
  `tags` varchar(4096)  NOT NULL DEFAULT '' COMMENT 'json,map,tagkey->regexp|value',
  `redefine_severity` tinyint(1) NULL DEFAULT 0 COMMENT 'is redefine severity?',
  `new_severity` tinyint(1) NOT NULL COMMENT '0:Emergency 1:Warning 2:Notice',
  `redefine_channels` tinyint(1) NULL DEFAULT 0 COMMENT 'is redefine channels?',
  `new_channels` varchar(255)  NOT NULL DEFAULT '' COMMENT 'split by space: sms voice email dingtalk wecom',
  `user_group_ids` varchar(250)  NOT NULL COMMENT 'split by space 1 34 5, notify cc to user_group_ids',
  `webhooks` text  NOT NULL,
  `extra_config` text  NULL,
  `redefine_webhooks` tinyint(1) NULL DEFAULT 0,
  `for_duration` bigint NOT NULL DEFAULT 0,
  `create_at` bigint NOT NULL DEFAULT 0,
  `create_by` varchar(64)  NOT NULL DEFAULT '',
  `update_at` bigint NOT NULL DEFAULT 0,
  `update_by` varchar(64)  NOT NULL DEFAULT '',
  `busi_groups` varchar(4096)  NOT NULL DEFAULT '[]',
  PRIMARY KEY (`id`),
  KEY (`update_at`),
  KEY (`group_id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4;

CREATE TABLE `target`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `group_id` bigint NOT NULL DEFAULT 0 COMMENT 'busi group id',
  `ident` varchar(191)  NOT NULL COMMENT 'target id',
  `note` varchar(255)  NOT NULL DEFAULT '' COMMENT 'append to alert event as field',
  `tags` varchar(512)  NOT NULL DEFAULT '' COMMENT 'append to series data as tags, split by space, append external space at suffix',
  `update_at` bigint NOT NULL DEFAULT 0,
  `host_ip` varchar(191)  NULL DEFAULT '' COMMENT 'IPv4 string',
  PRIMARY KEY (`id`),
  UNIQUE KEY (`ident`),
  INDEX idx_host_ip(`host_ip`),
  KEY (`group_id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4;



-- case1: target_idents; case2: target_tags
-- CREATE TABLE `collect_rule` (
--     `id` bigint unsigned not null auto_increment,
--     `group_id` bigint not null default 0 comment 'busi group id',
--     `cluster` varchar(128) not null,
--     `target_idents` varchar(512) not null default '' comment 'ident list, split by space',
--     `target_tags` varchar(512) not null default '' comment 'filter targets by tags, split by space',
--     `name` varchar(191) not null default '',
--     `note` varchar(255) not null default '',
--     `step` int not null,
--     `type` varchar(64) not null comment 'e.g. port proc log plugin',
--     `data` text not null,
--     `append_tags` varchar(255) not null default '' comment 'split by space: e.g. mod=n9e dept=cloud',
--     `create_at` bigint not null default 0,
--     `create_by` varchar(64) not null default '',
--     `update_at` bigint not null default 0,
--     `update_by` varchar(64) not null default '',
--     PRIMARY KEY (`id`),
--     KEY (`group_id`, `type`, `name`)
-- ) ENGINE=InnoDB DEFAULT CHARSET = utf8mb4;

CREATE TABLE `metric_view`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` varchar(191)  NOT NULL DEFAULT '',
  `cate` tinyint(1) NOT NULL COMMENT '0: preset 1: custom',
  `configs` varchar(8192)  NOT NULL DEFAULT '',
  `create_at` bigint NOT NULL DEFAULT 0,
  `create_by` bigint NOT NULL DEFAULT 0 COMMENT 'user id',
  `update_at` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY (`create_by`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4;

insert into metric_view(name, cate, configs) values('Host View', 0, '{"filters":[{"oper":"=","label":"__name__","value":"cpu_usage_idle"}],"dynamicLabels":[],"dimensionLabels":[{"label":"ident","value":""}]}');

CREATE TABLE `recording_rule`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `group_id` bigint NOT NULL DEFAULT 0 COMMENT 'group_id',
  `datasource_ids` varchar(255)  NOT NULL DEFAULT '' COMMENT 'datasource ids',
  `cluster` varchar(128)  NOT NULL,
  `name` varchar(255)  NOT NULL COMMENT 'new metric name',
  `note` varchar(255)  NOT NULL COMMENT 'rule note',
  `disabled` tinyint(1) NOT NULL DEFAULT 0 COMMENT '0:enabled 1:disabled',
  `prom_ql` varchar(8192)  NOT NULL COMMENT 'promql',
  `prom_eval_interval` int NOT NULL COMMENT 'evaluate interval',
  `append_tags` varchar(255)  NULL DEFAULT '' COMMENT 'split by space: service=n9e mod=api',
  `query_configs` text  NOT NULL,
  `create_at` bigint NULL DEFAULT 0,
  `create_by` varchar(64)  NULL DEFAULT '',
  `update_at` bigint NULL DEFAULT 0,
  `update_by` varchar(64)  NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `group_id` (`group_id`),
  KEY `update_at` (`update_at`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4;

CREATE TABLE `alert_aggr_view`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` varchar(191)  NOT NULL DEFAULT '',
  `rule` varchar(2048)  NOT NULL DEFAULT '',
  `cate` tinyint(1) NOT NULL COMMENT '0: preset 1: custom',
  `create_at` bigint NOT NULL DEFAULT 0,
  `create_by` bigint NOT NULL DEFAULT 0 COMMENT 'user id',
  `update_at` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY (`create_by`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4;

insert into alert_aggr_view(name, rule, cate) values('By BusiGroup, Severity', 'field:group_name::field:severity', 0);
insert into alert_aggr_view(name, rule, cate) values('By RuleName', 'field:rule_name', 0);

CREATE TABLE `alert_cur_event`  (
  `id` bigint UNSIGNED NOT NULL COMMENT 'use alert_his_event.id',
  `cate` varchar(128)  NOT NULL,
  `datasource_id` bigint NOT NULL DEFAULT 0 COMMENT 'datasource id',
  `cluster` varchar(128)  NOT NULL,
  `group_id` bigint UNSIGNED NOT NULL COMMENT 'busi group id of rule',
  `group_name` varchar(255)  NOT NULL DEFAULT '' COMMENT 'busi group name',
  `hash` varchar(64)  NOT NULL COMMENT 'rule_id + vector_pk',
  `rule_id` bigint UNSIGNED NOT NULL,
  `rule_name` varchar(255)  NOT NULL,
  `rule_note` varchar(2048)  NOT NULL DEFAULT 'alert rule note',
  `rule_prod` varchar(255)  NOT NULL DEFAULT '',
  `rule_algo` varchar(255)  NOT NULL DEFAULT '',
  `severity` tinyint(1) NOT NULL COMMENT '0:Emergency 1:Warning 2:Notice',
  `prom_for_duration` int NOT NULL COMMENT 'prometheus for, unit:s',
  `prom_ql` varchar(8192)  NOT NULL COMMENT 'promql',
  `prom_eval_interval` int NOT NULL COMMENT 'evaluate interval',
  `callbacks` varchar(255)  NOT NULL DEFAULT '' COMMENT 'split by space: http://a.com/api/x http://a.com/api/y',
  `runbook_url` varchar(255)  NULL DEFAULT NULL,
  `notify_recovered` tinyint(1) NOT NULL COMMENT 'whether notify when recovery',
  `notify_channels` varchar(255)  NOT NULL DEFAULT '' COMMENT 'split by space: sms voice email dingtalk wecom',
  `notify_groups` varchar(255)  NOT NULL DEFAULT '' COMMENT 'split by space: 233 43',
  `notify_repeat_next` bigint NOT NULL DEFAULT 0 COMMENT 'next timestamp to notify, get repeat settings from rule',
  `notify_cur_number` int NOT NULL DEFAULT 0,
  `target_ident` varchar(191)  NOT NULL DEFAULT '' COMMENT 'target ident, also in tags',
  `target_note` varchar(191)  NOT NULL DEFAULT '' COMMENT 'target note',
  `first_trigger_time` bigint NULL DEFAULT NULL,
  `trigger_time` bigint NOT NULL,
  `trigger_value` varchar(255)  NOT NULL,
  `annotations` text  NOT NULL COMMENT 'annotations',
  `rule_config` text  NOT NULL COMMENT 'annotations',
  `tags` varchar(1024)  NOT NULL DEFAULT '' COMMENT 'merge data_tags rule_tags, split by ,,',
  PRIMARY KEY (`id`),
  KEY (`hash`),
  KEY (`rule_id`),
  KEY (`trigger_time`, `group_id`),
  KEY (`notify_repeat_next`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4;

CREATE TABLE `alert_his_event`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `is_recovered` tinyint(1) NOT NULL,
  `cate` varchar(128)  NOT NULL,
  `datasource_id` bigint NOT NULL DEFAULT 0 COMMENT 'datasource id',
  `cluster` varchar(128)  NOT NULL,
  `group_id` bigint UNSIGNED NOT NULL COMMENT 'busi group id of rule',
  `group_name` varchar(255)  NOT NULL DEFAULT '' COMMENT 'busi group name',
  `hash` varchar(64)  NOT NULL COMMENT 'rule_id + vector_pk',
  `rule_id` bigint UNSIGNED NOT NULL,
  `rule_name` varchar(255)  NOT NULL,
  `rule_note` varchar(2048)  NOT NULL DEFAULT 'alert rule note',
  `rule_prod` varchar(255)  NOT NULL DEFAULT '',
  `rule_algo` varchar(255)  NOT NULL DEFAULT '',
  `severity` tinyint(1) NOT NULL COMMENT '0:Emergency 1:Warning 2:Notice',
  `prom_for_duration` int NOT NULL COMMENT 'prometheus for, unit:s',
  `prom_ql` varchar(8192)  NOT NULL COMMENT 'promql',
  `prom_eval_interval` int NOT NULL COMMENT 'evaluate interval',
  `callbacks` varchar(255)  NOT NULL DEFAULT '' COMMENT 'split by space: http://a.com/api/x http://a.com/api/y',
  `runbook_url` varchar(255)  NULL DEFAULT NULL,
  `notify_recovered` tinyint(1) NOT NULL COMMENT 'whether notify when recovery',
  `notify_channels` varchar(255)  NOT NULL DEFAULT '' COMMENT 'split by space: sms voice email dingtalk wecom',
  `notify_groups` varchar(255)  NOT NULL DEFAULT '' COMMENT 'split by space: 233 43',
  `notify_cur_number` int NOT NULL DEFAULT 0,
  `target_ident` varchar(191)  NOT NULL DEFAULT '' COMMENT 'target ident, also in tags',
  `target_note` varchar(191)  NOT NULL DEFAULT '' COMMENT 'target note',
  `first_trigger_time` bigint NULL DEFAULT NULL,
  `trigger_time` bigint NOT NULL,
  `trigger_value` varchar(255)  NOT NULL,
  `recover_time` bigint NOT NULL DEFAULT 0,
  `last_eval_time` bigint NOT NULL DEFAULT 0 COMMENT 'for time filter',
  `tags` varchar(1024)  NOT NULL DEFAULT '' COMMENT 'merge data_tags rule_tags, split by ,,',
  `annotations` text  NOT NULL COMMENT 'annotations',
  `rule_config` text  NOT NULL COMMENT 'annotations',
  PRIMARY KEY (`id`),
  KEY (`hash`),
  KEY (`rule_id`),
  KEY (`trigger_time`, `group_id`),
  KEY (`last_eval_time`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4;

CREATE TABLE `task_tpl`  (
  `id` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `group_id` int UNSIGNED NOT NULL COMMENT 'busi group id',
  `title` varchar(255)  NOT NULL DEFAULT '',
  `account` varchar(64)  NOT NULL,
  `batch` int UNSIGNED NOT NULL DEFAULT 0,
  `tolerance` int UNSIGNED NOT NULL DEFAULT 0,
  `timeout` int UNSIGNED NOT NULL DEFAULT 0,
  `pause` varchar(255)  NOT NULL DEFAULT '',
  `script` text  NOT NULL,
  `args` varchar(512)  NOT NULL DEFAULT '',
  `tags` varchar(255)  NOT NULL DEFAULT '' COMMENT 'split by space',
  `create_at` bigint NOT NULL DEFAULT 0,
  `create_by` varchar(64)  NOT NULL DEFAULT '',
  `update_at` bigint NOT NULL DEFAULT 0,
  `update_by` varchar(64)  NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY (`group_id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4;

CREATE TABLE `task_tpl_host`  (
  `ii` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `id` int UNSIGNED NOT NULL COMMENT 'task tpl id',
  `host` varchar(128)  NOT NULL COMMENT 'ip or hostname',
  PRIMARY KEY (`ii`),
  KEY (`id`, `host`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4;

CREATE TABLE `task_record`  (
  `id` bigint UNSIGNED NOT NULL COMMENT 'ibex task id',
  `event_id` bigint NOT NULL DEFAULT 0 COMMENT 'event id',
  `group_id` bigint NOT NULL COMMENT 'busi group id',
  `ibex_address` varchar(128)  NOT NULL,
  `ibex_auth_user` varchar(128)  NOT NULL DEFAULT '',
  `ibex_auth_pass` varchar(128)  NOT NULL DEFAULT '',
  `title` varchar(255)  NOT NULL DEFAULT '',
  `account` varchar(64)  NOT NULL,
  `batch` int UNSIGNED NOT NULL DEFAULT 0,
  `tolerance` int UNSIGNED NOT NULL DEFAULT 0,
  `timeout` int UNSIGNED NOT NULL DEFAULT 0,
  `pause` varchar(255)  NOT NULL DEFAULT '',
  `script` text  NOT NULL,
  `args` varchar(512)  NOT NULL DEFAULT '',
  `create_at` bigint NOT NULL DEFAULT 0,
  `create_by` varchar(64)  NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY (`create_at`, `group_id`),
  KEY (`create_by`),
  KEY (`event_id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4;

CREATE TABLE `alerting_engines`  (
  `id` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `instance` varchar(128)  NOT NULL DEFAULT '' COMMENT 'instance identification, e.g. 10.9.0.9:9090',
  `datasource_id` bigint NOT NULL DEFAULT 0 COMMENT 'datasource id',
  `engine_cluster` varchar(128)  NOT NULL DEFAULT '' COMMENT 'n9e-alert cluster',
  `clock` bigint NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4;

CREATE TABLE `datasource`  (
  `id` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` varchar(191)  NOT NULL DEFAULT '',
  `description` varchar(255)  NOT NULL DEFAULT '',
  `category` varchar(255)  NOT NULL DEFAULT '',
  `plugin_id` int UNSIGNED NOT NULL DEFAULT 0,
  `plugin_type` varchar(255)  NOT NULL DEFAULT '',
  `plugin_type_name` varchar(255)  NOT NULL DEFAULT '',
  `cluster_name` varchar(255)  NOT NULL DEFAULT '',
  `settings` text  NOT NULL,
  `status` varchar(255)  NOT NULL DEFAULT '',
  `http` varchar(4096)  NOT NULL DEFAULT '',
  `auth` varchar(8192)  NOT NULL DEFAULT '',
  `created_at` bigint NOT NULL DEFAULT 0,
  `created_by` varchar(64)  NOT NULL DEFAULT '',
  `updated_at` bigint NOT NULL DEFAULT 0,
  `updated_by` varchar(64)  NOT NULL DEFAULT '',
  `is_default` tinyint(1) NOT NULL DEFAULT 0 COMMENT 'is default datasource',
  PRIMARY KEY (`id`),
  UNIQUE KEY (`name`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4;

CREATE TABLE `builtin_cate`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` varchar(191)  NOT NULL,
  `user_id` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4;

CREATE TABLE `notify_tpl`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `channel` varchar(32)  NOT NULL,
  `name` varchar(255)  NOT NULL,
  `content` text  NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY (`channel`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4;

INSERT INTO `notify_tpl`(id,channel,name,content) VALUES (1, 'telegram', 'telegram', '**级别状态**: {{if .IsRecovered}}<font color=\"info\">S{{.Severity}} Recovered</font>{{else}}<font color=\"warning\">S{{.Severity}} Triggered</font>{{end}}   \n**规则标题**: {{.RuleName}}{{if .RuleNote}}   \n**规则备注**: {{.RuleNote}}{{end}}{{if .TargetIdent}}   \n**监控对象**: {{.TargetIdent}}{{end}}   \n**监控指标**: {{.TagsJSON}}{{if not .IsRecovered}}   \n**触发时值**: {{.TriggerValue}}{{end}}   \n{{if .IsRecovered}}**恢复时间**: {{timeformat .LastEvalTime}}{{else}}**首次触发时间**: {{timeformat .FirstTriggerTime}}{{end}}   \n{{$time_duration := sub now.Unix .FirstTriggerTime }}{{if .IsRecovered}}{{$time_duration = sub .LastEvalTime .FirstTriggerTime }}{{end}}**距离首次告警**: {{humanizeDurationInterface $time_duration}}\n**发送时间**: {{timestamp}}');
INSERT INTO `notify_tpl`(id,channel,name,content) VALUES (2, 'wecom', 'wecom', '**级别状态**: {{if .IsRecovered}}<font color=\"info\">S{{.Severity}} Recovered</font>{{else}}<font color=\"warning\">S{{.Severity}} Triggered</font>{{end}}   \n**规则标题**: {{.RuleName}}{{if .RuleNote}}   \n**规则备注**: {{.RuleNote}}{{end}}{{if .TargetIdent}}   \n**监控对象**: {{.TargetIdent}}{{end}}   \n**监控指标**: {{.TagsJSON}}{{if not .IsRecovered}}   \n**触发时值**: {{.TriggerValue}}{{end}}   \n{{if .IsRecovered}}**恢复时间**: {{timeformat .LastEvalTime}}{{else}}**首次触发时间**: {{timeformat .FirstTriggerTime}}{{end}}   \n{{$time_duration := sub now.Unix .FirstTriggerTime }}{{if .IsRecovered}}{{$time_duration = sub .LastEvalTime .FirstTriggerTime }}{{end}}**距离首次告警**: {{humanizeDurationInterface $time_duration}}\n**发送时间**: {{timestamp}}');
INSERT INTO `notify_tpl`(id,channel,name,content) VALUES (3, 'dingtalk', 'dingtalk', '#### {{if .IsRecovered}}<font color=\"#008800\">S{{.Severity}} - Recovered - {{.RuleName}}</font>{{else}}<font color=\"#FF0000\">S{{.Severity}} - Triggered - {{.RuleName}}</font>{{end}}\n\n---\n\n- **规则标题**: {{.RuleName}}{{if .RuleNote}}\n- **规则备注**: {{.RuleNote}}{{end}}\n{{if not .IsRecovered}}- **触发时值**: {{.TriggerValue}}{{end}}\n{{if .TargetIdent}}- **监控对象**: {{.TargetIdent}}{{end}}\n- **监控指标**: {{.TagsJSON}}\n- {{if .IsRecovered}}**恢复时间**: {{timeformat .LastEvalTime}}{{else}}**触发时间**: {{timeformat .TriggerTime}}{{end}}\n- **发送时间**: {{timestamp}}\n	');
INSERT INTO `notify_tpl`(id,channel,name,content) VALUES (4, 'email', 'email', '<!DOCTYPE html>\n	<html lang=\"en\">\n	<head>\n		<meta charset=\"UTF-8\">\n		<meta http-equiv=\"X-UA-Compatible\" content=\"ie=edge\">\n		<title>夜莺告警通知</title>\n		<style type=\"text/css\">\n			.wrapper {\n				background-color: #f8f8f8;\n				padding: 15px;\n				height: 100%;\n			}\n			.main {\n				width: 600px;\n				padding: 30px;\n				margin: 0 auto;\n				background-color: #fff;\n				font-size: 12px;\n				font-family: verdana,\'Microsoft YaHei\',Consolas,\'Deja Vu Sans Mono\',\'Bitstream Vera Sans Mono\';\n			}\n			header {\n				border-radius: 2px 2px 0 0;\n			}\n			header .title {\n				font-size: 14px;\n				color: #333333;\n				margin: 0;\n			}\n			header .sub-desc {\n				color: #333;\n				font-size: 14px;\n				margin-top: 6px;\n				margin-bottom: 0;\n			}\n			hr {\n				margin: 20px 0;\n				height: 0;\n				border: none;\n				border-top: 1px solid #e5e5e5;\n			}\n			em {\n				font-weight: 600;\n			}\n			table {\n				margin: 20px 0;\n				width: 100%;\n			}\n	\n			table tbody tr{\n				font-weight: 200;\n				font-size: 12px;\n				color: #666;\n				height: 32px;\n			}\n	\n			.succ {\n				background-color: green;\n				color: #fff;\n			}\n	\n			.fail {\n				background-color: red;\n				color: #fff;\n			}\n	\n			.succ th, .succ td, .fail th, .fail td {\n				color: #fff;\n			}\n	\n			table tbody tr th {\n				width: 80px;\n				text-align: right;\n			}\n			.text-right {\n				text-align: right;\n			}\n			.body {\n				margin-top: 24px;\n			}\n			.body-text {\n				color: #666666;\n				-webkit-font-smoothing: antialiased;\n			}\n			.body-extra {\n				-webkit-font-smoothing: antialiased;\n			}\n			.body-extra.text-right a {\n				text-decoration: none;\n				color: #333;\n			}\n			.body-extra.text-right a:hover {\n				color: #666;\n			}\n			.button {\n				width: 200px;\n				height: 50px;\n				margin-top: 20px;\n				text-align: center;\n				border-radius: 2px;\n				background: #2D77EE;\n				line-height: 50px;\n				font-size: 20px;\n				color: #FFFFFF;\n				cursor: pointer;\n			}\n			.button:hover {\n				background: rgb(25, 115, 255);\n				border-color: rgb(25, 115, 255);\n				color: #fff;\n			}\n			footer {\n				margin-top: 10px;\n				text-align: right;\n			}\n			.footer-logo {\n				text-align: right;\n			}\n			.footer-logo-image {\n				width: 108px;\n				height: 27px;\n				margin-right: 10px;\n			}\n			.copyright {\n				margin-top: 10px;\n				font-size: 12px;\n				text-align: right;\n				color: #999;\n				-webkit-font-smoothing: antialiased;\n			}\n		</style>\n	</head>\n	<body>\n	<div class=\"wrapper\">\n		<div class=\"main\">\n			<header>\n				<h3 class=\"title\">{{.RuleName}}</h3>\n				<p class=\"sub-desc\"></p>\n			</header>\n	\n			<hr>\n	\n			<div class=\"body\">\n				<table cellspacing=\"0\" cellpadding=\"0\" border=\"0\">\n					<tbody>\n					{{if .IsRecovered}}\n					<tr class=\"succ\">\n						<th>级别状态：</th>\n						<td>S{{.Severity}} Recovered</td>\n					</tr>\n					{{else}}\n					<tr class=\"fail\">\n						<th>级别状态：</th>\n						<td>S{{.Severity}} Triggered</td>\n					</tr>\n					{{end}}\n	\n					<tr>\n						<th>策略备注：</th>\n						<td>{{.RuleNote}}</td>\n					</tr>\n					<tr>\n						<th>设备备注：</th>\n						<td>{{.TargetNote}}</td>\n					</tr>\n					{{if not .IsRecovered}}\n					<tr>\n						<th>触发时值：</th>\n						<td>{{.TriggerValue}}</td>\n					</tr>\n					{{end}}\n	\n					{{if .TargetIdent}}\n					<tr>\n						<th>监控对象：</th>\n						<td>{{.TargetIdent}}</td>\n					</tr>\n					{{end}}\n					<tr>\n						<th>监控指标：</th>\n						<td>{{.TagsJSON}}</td>\n					</tr>\n	\n					{{if .IsRecovered}}\n					<tr>\n						<th>恢复时间：</th>\n						<td>{{timeformat .LastEvalTime}}</td>\n					</tr>\n					{{else}}\n					<tr>\n						<th>触发时间：</th>\n						<td>\n							{{timeformat .TriggerTime}}\n						</td>\n					</tr>\n					{{end}}\n	\n					<tr>\n						<th>发送时间：</th>\n						<td>\n							{{timestamp}}\n						</td>\n					</tr>\n					</tbody>\n				</table>\n	\n				<hr>\n	\n				<footer>\n					<div class=\"copyright\" style=\"font-style: italic\">\n						报警太多？使用 <a href=\"https://flashcat.cloud/product/flashduty/\" target=\"_blank\">FlashDuty</a> 做告警聚合降噪、排班OnCall！\n					</div>\n				</footer>\n			</div>\n		</div>\n	</div>\n	</body>\n	</html>');
INSERT INTO `notify_tpl`(id,channel,name,content) VALUES (5, 'feishu', 'feishu', '级别状态: S{{.Severity}} {{if .IsRecovered}}Recovered{{else}}Triggered{{end}}   \n规则名称: {{.RuleName}}{{if .RuleNote}}   \n规则备注: {{.RuleNote}}{{end}}   \n监控指标: {{.TagsJSON}}\n{{if .IsRecovered}}恢复时间：{{timeformat .LastEvalTime}}{{else}}触发时间: {{timeformat .TriggerTime}}\n触发时值: {{.TriggerValue}}{{end}}\n发送时间: {{timestamp}}');
INSERT INTO `notify_tpl`(id,channel,name,content) VALUES (6, 'feishucard', 'feishucard', '{{ if .IsRecovered }}\n{{- if ne .Cate \"host\"}}\n**告警集群:** {{.Cluster}}{{end}}   \n**级别状态:** S{{.Severity}} Recovered   \n**告警名称:** {{.RuleName}}   \n**恢复时间:** {{timeformat .LastEvalTime}}   \n**告警描述:** **服务已恢复**   \n{{- else }}\n{{- if ne .Cate \"host\"}}   \n**告警集群:** {{.Cluster}}{{end}}   \n**级别状态:** S{{.Severity}} Triggered   \n**告警名称:** {{.RuleName}}   \n**触发时间:** {{timeformat .TriggerTime}}   \n**发送时间:** {{timestamp}}   \n**触发时值:** {{.TriggerValue}}   \n{{if .RuleNote }}**告警描述:** **{{.RuleNote}}**{{end}}   \n{{- end -}}');
INSERT INTO `notify_tpl`(id,channel,name,content) VALUES (7, 'mailsubject', 'mailsubject', '{{if .IsRecovered}}Recovered{{else}}Triggered{{end}}: {{.RuleName}} {{.TagsJSON}}');
INSERT INTO `notify_tpl`(id,channel,name,content) VALUES (8, 'mm', 'mm', '级别状态: S{{.Severity}} {{if .IsRecovered}}Recovered{{else}}Triggered{{end}}   \n规则名称: {{.RuleName}}{{if .RuleNote}}   \n规则备注: {{.RuleNote}}{{end}}   \n监控指标: {{.TagsJSON}}   \n{{if .IsRecovered}}恢复时间：{{timeformat .LastEvalTime}}{{else}}触发时间: {{timeformat .TriggerTime}}   \n触发时值: {{.TriggerValue}}{{end}}   \n发送时间: {{timestamp}}');

CREATE TABLE `sso_config`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` varchar(191)  NOT NULL,
  `content` text  NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY (`name`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4;

CREATE TABLE `es_index_pattern`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `datasource_id` bigint NOT NULL DEFAULT 0 COMMENT 'datasource id',
  `name` varchar(191)  NOT NULL,
  `time_field` varchar(128)  NOT NULL DEFAULT '@timestamp',
  `allow_hide_system_indices` tinyint(1) NOT NULL DEFAULT 0,
  `fields_format` varchar(4096)  NOT NULL DEFAULT '',
  `create_at` bigint NULL DEFAULT 0,
  `create_by` varchar(64)  NULL DEFAULT '',
  `update_at` bigint NULL DEFAULT 0,
  `update_by` varchar(64)  NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY (`datasource_id`, `name`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4;