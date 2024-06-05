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
   echo "Starting telegram"
  ./telegram --config=$configPath 2>&1 | tee logs/telegram &
  echo "Starting statusserver"
  ./statusserver --config=$configPath 2>&1 | tee logs/status &
  echo "Starting storageserver"
  ./storageserver --config=$configPath 2>&1 | tee logs/storage &
  echo "Starting frontendserver"
  ./frontendserver --config=$configPath 2>&1 | tee logs/frontend &
  echo "Starting main"
  ./main --config=$configPath 2>&1 | tee logs/main 
else
  echo "Starting only $runcmd"
  ./$runcmd --config=$configPath 2>&1 | tee logs/$runcmd
fi
