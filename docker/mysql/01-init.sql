-- database and table schemas 

create database `anomaly_tracker`;

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
    primary key (`id`),
    foreign key (`created_by`) references users(`id`)
);

create table `anomaly_tracker`.`user_groups_members` (
    `group_id` int not null,
    `user_id` int not null,
    `created_by` int not null,
    `created_dttm` datetime not null default NOW(),
    primary key (`group_id`, `user_id`),
    foreign key (`group_id`) references user_groups(`id`),
    foreign key (`user_id`) references users(`id`),
    foreign key (`created_by`) references users(`id`)
);

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
    unique key (`anom_id`, `group_id`),
    foreign key (`user_id`) references users(`id`),
    foreign key (`group_id`) references user_groups(`id`)
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
    unique key (`user_id`, `group_id`),
    foreign key (`user_id`) references users(`id`),
    foreign key (`group_id`) references user_groups(`id`),
    foreign key (`created_by`) references users(`id`)
);
