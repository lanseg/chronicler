#!/bin/sh
set -uo pipefail

BOT_DIR="${BOT_DIR:-/telegram/files}"
BOT_TEMP_DIR="${BOT_TEMP_DIR:-/tmp/telegram}"

echo "Api id: ${BOT_API_ID:-Undefined}"
echo "Api hash: ${BOT_API_HASH:-Undefined}"
echo "Dir: ${BOT_DIR:-Undefined}"
echo "Temp: ${BOT_TEMP_DIR:-Undefined}"

mkdir -p "$BOT_DIR"
mkdir -p "$BOT_TEMP_DIR"

echo "Starting telegram-bot-api"

./telegram-bot-api --api-id=$BOT_API_ID \
    --api-hash=$BOT_API_HASH \
    --dir=$BOT_DIR \
    --temp-dir=$BOT_TEMP_DIR \
    --local -l log
