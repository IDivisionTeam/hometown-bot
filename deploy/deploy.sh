#!/usr/bin/env bash

set -Eeuo pipefail

sudo systemctl stop hometown-bot

cp hometown-bot $APP_DIR/

sudo systemctl daemon-reload
sudo systemctl restart hometown-bot
