CREATE USER 'devadmin'@'%' IDENTIFIED BY 'devadmin';
GRANT ALL PRIVILEGES ON *.* TO 'devadmin'@'%';

create database `anomaly_tracker`;
create table `anomaly_tracker`.`anomalies` (
    `id` int not null auto_increment,
    `anom_id` varchar(7) not null,
    `anom_system` varchar(32) not null,
    `anom_type` varchar(16) not null,
    `anom_name` varchar(64) not null,
    `user_id` int not null,
    `group_id` int not null,
    `created_dttm` datetime not null default NOW(),
    primary key (`id`),
    unique key (`anom_id`, `group_id`)
);

create table `anomaly_tracker`.`api_keys` (
    `id` int not null auto_increment,
    `key` varchar(64) not null,
    `type` varchar(32) not null default "user",
    `user_id` int not null,
    `group_id` int not null,
    `created_by` int not null,
    `created_dttm` datetime not null default NOW(),
    primary key (`id`),
    unique key (`key`),
    unique key (`user_id`, `group_id`)
);

create table `anomaly_tracker`.`users` (
    `id` int not null auto_increment,
    `username` varchar(16) not null,
    `created` datetime not null default NOW(),
    primary key (`id`)
);

create table `anomaly_tracker`.`user_groups` (
    `id` int not null auto_increment,
    `name` varchar(32),
    `created_by` int not null,
    primary key (`id`)
);

create table `anomaly_tracker`.`user_groups_members` (
    `group_id` int not null,
    `user_id` int not null,
    `created_by` int not null,
    `created_dttm` datetime not null default NOW(),
    primary key (`group_id`, `user_id`)
);

insert anomaly_tracker.users (`username`) 
    values ('rboss'), ('bpowers'), ('zharvest');

insert anomaly_tracker.user_groups (`name`, `created_by`) values
	('Rick\'s Group', 0),
    ('Zach\'s Group', 1);

insert anomaly_tracker.api_keys (`key`, `user_id`, `group_id`, `created_by`) values
    ('00000-00000-0000-00000-00000', 0, 0, 0),
    ('00000-00000-0000-00000-00001', 1, 0, 0),
    ('00000-00000-0000-00000-00002', 2, 1, 2);

insert anomaly_tracker.anomalies 
        (`anom_id`, `anom_system`, `anom_type`, `anom_name`, `user_id`)
    values
        ('HE1-FA9', 'Jita', 'Combat', 'Guristas Refuge', 0),
        ('AIP-81J', 'Perimiter', 'Combat', 'Guristas Hideaway', 0),
        ('POK-184', 'Perimiter', 'Combat', 'Guristas Hidden Hideaway', 0),
        ('KNQ-91M', 'Jan', 'Ice', 'Ice Field', 1),
        ('NOK-K8J', 'Jan', 'Combat', 'Sansha Rally Point', 1),
        ('MNU-I09', 'Jan', 'Combat', 'Guristas Forlorn Refuge', 1),
        ('NYA-813', 'Noni', 'Combat', 'Blood Raider Post', 2),
        ('HE1-FA9', 'Jita', 'Combat', 'Guristas Refuge', 2),
        ('Y17-FM8', 'Noni', 'Combat', 'Guristas Hideaway', 2);
