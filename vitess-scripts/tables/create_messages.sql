create table messages (
    id bigint not null auto_increment,
    chat_id bigint not null,
    user_id bigint not null,
    text character varying (2000) not null,
    created_at datetime NOT NULL,
    primary key (id)
) engine=innodb;
