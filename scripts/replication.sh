#!/bin/bash

docker-compose -f ./docker/docker-compose.yml --project-directory . down
#rm -rf ./master/data/*
sudo rm -rf ./mnt/mysql-slave/data/*
docker-compose -f ./docker/docker-compose.yml --project-directory . build
docker-compose -f ./docker/docker-compose.yml --project-directory . up -d

until docker-compose -f ./docker/docker-compose.yml --project-directory . exec mysql sh -c 'export MYSQL_PWD=123456; mysql -u root -e ";"'
do
    echo "Waiting for mysql_master database connection..."
    sleep 4
done

priv_stmt='drop user "repl_user"; create user "repl_user"@"%" identified with mysql_native_password by "repl_pwd"; GRANT REPLICATION SLAVE ON *.* TO repl_user@"%"; FLUSH PRIVILEGES;'
docker-compose -f ./docker/docker-compose.yml --project-directory . exec mysql sh -c "export MYSQL_PWD=123456; mysql -u root -e '$priv_stmt'"

until docker-compose -f ./docker/docker-compose.yml --project-directory . exec mysql-slave sh -c 'export MYSQL_PWD=123456; mysql -u root -e ";"'
do
    echo "Waiting for mysql_slave database connection..."
    sleep 4
done

docker-ip() {
    docker inspect --format '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "$@"
}

docker-compose -f ./docker/docker-compose.yml --project-directory . exec mysql sh -c "export MYSQL_PWD=123456; /usr/bin/mysqldump -u root --all-databases --triggers --routines --events" | gzip -6 -c > backup.sql.gz
zcat backup.sql.gz | docker exec -i otus-hiload_mysql-slave_1 sh -c "export MYSQL_PWD=123456; /usr/bin/mysql -u root"

MS_STATUS=`docker-compose -f ./docker/docker-compose.yml --project-directory . exec mysql sh -c 'export MYSQL_PWD=123456; mysql -sN -u root -e "SHOW MASTER STATUS"'`
CURRENT_LOG=`echo $MS_STATUS | awk '{print $1}'`
CURRENT_POS=`echo $MS_STATUS | awk '{print $2}'`
echo "MS_STATUS=$MS_STATUS"
echo "CURRENT_LOG=$CURRENT_LOG"
echo "CURRENT_POS=$CURRENT_POS"

#start_slave_stmt="CHANGE MASTER TO MASTER_HOST='$(docker-ip otus-hiload_mysql_1)',MASTER_USER='repl_user',MASTER_PASSWORD='repl_pwd',MASTER_LOG_FILE='$CURRENT_LOG',MASTER_LOG_POS=$CURRENT_POS; START SLAVE;"
start_slave_stmt="CHANGE MASTER TO MASTER_HOST='mysql',MASTER_PORT=3306,MASTER_USER='repl_user',MASTER_PASSWORD='repl_pwd',MASTER_LOG_FILE='$CURRENT_LOG',MASTER_LOG_POS=$CURRENT_POS,MASTER_CONNECT_RETRY=10; START SLAVE;"
start_slave_cmd='export MYSQL_PWD=123456; mysql -u root -e "'
start_slave_cmd+="$start_slave_stmt"
start_slave_cmd+='"'
echo $start_slave_cmd
docker-compose -f ./docker/docker-compose.yml --project-directory . exec mysql-slave sh -c "$start_slave_cmd"

docker-compose -f ./docker/docker-compose.yml --project-directory . exec mysql-slave sh -c "export MYSQL_PWD=123456; mysql -u root -e 'SHOW SLAVE STATUS \G'"
