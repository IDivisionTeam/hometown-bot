#!/usr/bin/env bash

set -Eeuo pipefail

cp main $APP_DIR/

sudo systemctl enable $APP_DIR/hometown-bot.service
sudo systemctl restart hometown-bot
sudo systemctl daemon-reload
