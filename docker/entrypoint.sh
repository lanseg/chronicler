#!/bin/sh
set -euo pipefail

runcmd=""
case "$@" in
  "frontendserver"|"main"|"statusserver"|"storageserver"|"all")
    runcmd="$@"
    ;;
  *) echo "Unknown binary to run: \"$@\""; exit 1;;
esac

mkdir -p logs
if [[ "$runcmd" == "all" ]]
then
  ./statusserver --config=./config.json 2>&1 > logs/status &
  ./storageserver --config=./config.json 2>&1 > logs/storage &
  ./frontendserver --config=./config.json 2>&1 > logs/frontend &
  ./main --config=./config.json 2>&1 > logs/main &
else
  ./$runcmd --config=./config.json 2>&1 | tee logs/$runcmd
fi