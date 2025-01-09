#!/bin/sh
set -euo pipefail

mkdir -p /chronicler/logs/

./server --FrontendPort $FRONTEND_PORT \
         --StatusServer $STATUSSERVER \
         --StorageServer $STORAGESERVER \
         --StaticRoot=/chronicler/frontend/static 
