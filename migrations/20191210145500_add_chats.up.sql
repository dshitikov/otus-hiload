ALTER TABLE users MODIFY id bigint not null auto_increment;

CREATE TABLE chats (
    id bigint not null,
    user1_id bigint not null,
    user2_id bigint not null,
    updated_at datetime NOT NULL,
    CONSTRAINT chats_user1_id_fk FOREIGN KEY (user1_id)  REFERENCES users (id),
    CONSTRAINT chats_user2_id_fk FOREIGN KEY (user2_id)  REFERENCES users (id),
    primary key(id)
);

CREATE INDEX chats_user1_id ON chats(user1_id);
CREATE INDEX chats_user2_id ON chats(user2_id);
CREATE INDEX chats_updated_at ON chats(updated_at);
