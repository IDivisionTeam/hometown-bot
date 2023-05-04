#!/usr/bin/env bash

set -Eeuo pipefail

systemctl stop hometown-bot
cp main $APP_DIR/

systemctl enable $APP_DIR/hometown-bot.service
systemctl restart hometown-bot
systemctl daemon-reload
