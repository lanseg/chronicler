#!/bin/sh
set -euo pipefail

runcmd="$1"
configPath="$2"
case "$runcmd" in
  "frontendserver"|"main"|"statusserver"|"storageserver"|"telegram"|"all")
    echo "Command: $runcmd"
    ;;
  *) echo "Unknown binary to run: \"$@\""; exit 1;;
esac

mkdir -p logs
if [[ "$runcmd" == "all" ]]
then
  ./telegram --config=$configPath 2>&1 > logs/telegram &
  ./statusserver --config=$configPath 2>&1 > logs/status &
  ./storageserver --config=$configPath 2>&1 > logs/storage &
  ./frontendserver --config=$configPath 2>&1 > logs/frontend &
  ./main --config=$configPath 2>&1 > logs/main &
else
  ./$runcmd --config=$configPath 2>&1 | tee logs/$runcmd
fi