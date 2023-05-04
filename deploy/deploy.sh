#!/usr/bin/env bash

set -Eeuo pipefail

sudo systemctl stop hometown-bot

cp main $APP_DIR/

sudo systemctl daemon-reload
sudo systemctl restart hometown-bot
