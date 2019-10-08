#!/bin/sh

log () {
  echo "$0: $1"
}

die() {
  log "$1"
  exit 1
}

parse_dsn () {
  local dsn=$1
  local proto="$(echo ${dsn} | grep :// | sed -e's,^\(.*://\).*,\1,g')"
  local url=$(echo $1 | sed -e s,$proto,,g)
  local userpass="$(echo ${url} | grep @ | cut -d@ -f1)"
  local pass="$(echo ${userpass} | grep : | cut -d: -f2)"
  
  if [ -n "${pass}" ]; then
    local user="$(echo ${userpass} | grep : | cut -d: -f1)"
  else
    local user=${userpass}
  fi

  # There variables are used outside of the function
  host=$(echo $url | sed -e s,${userpass}@,,g | cut -d/ -f1 | tr -d tcp\(|tr -d \))
  port="$(echo ${host} | grep : | cut -d: -f2 | sed -e 's,[^0-9],,g')"
  
  if [ -z ${port} ]; then
    case "${proto}" in
      "postgres://")
      port="5432"
      ;;
      "mysql://")
      port="3306"
      ;;
      *)
      die "no port is set and no default"
      ;;
    esac
  else
    host=$(echo ${host} | sed -e s,:${port},,g | cut -d/ -f1)
  fi
  
  local path="/$(echo ${url} | grep / | cut -d/ -f2-)"

  usertrunc=$(echo $user | awk '{ string=substr($0, 1, 3); print string; }' )
  log "proto: $proto, user: $usertrunc***, pass: ***, host: $host, port: $port, path: $path"
}

wait_for_it () {
  local host="$1"
  local port="$2"
  local msg="$3"

  res="1"
  
  until [ "0" = ${res} ]; do
    log "$msg"
    nc -w 5 -z "$host" "$port" 2>/dev/null
    res=$?
    if [ "0" = ${res} ]; then
      log "connect successfully"
    else
      sleep 1
    fi
  done
}

parse_dsn $1
wait_for_it ${host} ${port} "trying to connect to $host:$port"
