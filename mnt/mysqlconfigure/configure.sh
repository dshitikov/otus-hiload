#!/bin/bash

#echo "Waiting for mysql to get up"
#sleep 2
export MYSQL_PWD=$MYSQL_ROOT_PASSWORD

if [[ -z $MYSQL_REPLICATION_INIT ]]; then
  echo "restarting replication"
else
  echo "initializing replication"
  echo "create user for tarantool"
  mysql --host mysql -uroot -e "CREATE USER  IF NOT EXISTS 'tarantool'@'%' IDENTIFIED BY '123456';"
  mysql --host mysql -uroot -e "GRANT REPLICATION CLIENT ON *.* TO 'tarantool'@'%';"
  mysql --host mysql -uroot -e "GRANT REPLICATION SLAVE ON *.* TO 'tarantool'@'%';"
  mysql --host mysql -uroot -e "GRANT SELECT ON *.* TO 'tarantool'@'%';"
  mysql --host mysql -uroot -e "FLUSH PRIVILEGES;"
fi
echo "script finished"
