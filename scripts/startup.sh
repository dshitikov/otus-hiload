#!/bin/sh

# exit when any command fails
set -e

./scripts/wait-for-it.sh ${DB_URI}

exec ./bin/build
