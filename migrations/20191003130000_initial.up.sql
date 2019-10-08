create table users (
  id integer auto_increment not null,
  login character varying (255) not null,
  name character varying (255) not null,
  password_hash character varying (255) not null,
  created_at datetime NOT NULL,
  description character varying(1000) NOT NULL DEFAULT "",
  photo_file character varying(255) NOT NULL DEFAULT "",
  primary key (id)
) engine=innodb;