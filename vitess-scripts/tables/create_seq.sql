create table messages_seq(
    id int,
    next_id bigint,
    cache bigint,
    primary key(id)
) comment 'vitess_sequence';

insert into messages_seq(id, next_id, cache) values(0, 1, 100);
