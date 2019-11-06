#!/bin/bash

echo "Waiting for mysql to get up"
sleep 30
export MYSQL_PWD=$MYSQL_ROOT_PASSWORD

if [ -z $MYSQL_REPLICATION_INIT]; then
  echo "restarting replication"

  mysql --host mysql -uroot -e "SET GLOBAL rpl_semi_sync_master_enabled = 1;"

  mysql --host mysql-slave -uroot -e "START SLAVE;"
  mysql --host mysql-slave -uroot -e "STOP SLAVE IO_THREAD;"
  mysql --host mysql-slave -uroot -e "SET GLOBAL rpl_semi_sync_slave_enabled = 1;"
  mysql --host mysql-slave -uroot -e "START SLAVE IO_THREAD;"

  mysql --host mysql-slave2 -uroot -e "START SLAVE;"
  mysql --host mysql-slave2 -uroot -e "STOP SLAVE IO_THREAD;"
  mysql --host mysql-slave2 -uroot -e "SET GLOBAL rpl_semi_sync_slave_enabled = 1;"
  mysql --host mysql-slave2 -uroot -e "START SLAVE IO_THREAD;"
else
  echo "initializing replication"
  echo "create user on master"
  mysql --host mysql -uroot -e "CREATE USER 'repl_user'@'%' identified by 'repl_pwd'; GRANT REPLICATION SLAVE ON *.* TO repl_user@'%'; FLUSH PRIVILEGES;"
  mysql --host mysql -uroot -e "INSTALL PLUGIN rpl_semi_sync_master SONAME 'semisync_master.so';"
  mysql --host mysql -uroot -e "SET GLOBAL rpl_semi_sync_master_enabled = 1;"
  # create master dump
  echo "create master dump"
  mysqldump --host mysql -uroot --add-drop-database --triggers --routines --events --single-transaction --databases hiload-db | gzip -6 -c > /tmp/backup.sql.gz
  # load to slaves
  echo "slave"
  zcat /tmp/backup.sql.gz | mysql --host mysql-slave -uroot
  mysql --host mysql-slave -uroot -e "CHANGE MASTER TO MASTER_HOST='mysql', MASTER_PORT=3306, MASTER_USER='repl_user', MASTER_PASSWORD='repl_pwd', MASTER_AUTO_POSITION=1;START SLAVE;"
  mysql --host mysql-slave -uroot -e "INSTALL PLUGIN rpl_semi_sync_slave SONAME 'semisync_slave.so';"
  mysql --host mysql-slave -uroot -e "STOP SLAVE IO_THREAD;"
  mysql --host mysql-slave -uroot -e "SET GLOBAL rpl_semi_sync_slave_enabled = 1;"
  mysql --host mysql-slave -uroot -e "START SLAVE IO_THREAD;"
  #
  echo "slave2"
  zcat /tmp/backup.sql.gz | mysql --host mysql-slave2 -uroot
  mysql --host mysql-slave2 -uroot -e "CHANGE MASTER TO MASTER_HOST='mysql', MASTER_PORT=3306, MASTER_USER='repl_user', MASTER_PASSWORD='repl_pwd', MASTER_AUTO_POSITION=1;START SLAVE;"
  mysql --host mysql-slave2 -uroot -e "INSTALL PLUGIN rpl_semi_sync_slave SONAME 'semisync_slave.so';"
  mysql --host mysql-slave2 -uroot -e "STOP SLAVE IO_THREAD;"
  mysql --host mysql-slave2 -uroot -e "SET GLOBAL rpl_semi_sync_slave_enabled = 1;"
  mysql --host mysql-slave2 -uroot -e "START SLAVE IO_THREAD;"
fi
echo "script finished"

