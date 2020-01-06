CREATE INDEX name_idx ON users(name(10)) using btree;
CREATE INDEX last_name_idx ON users(last_name(13)) using btree;
