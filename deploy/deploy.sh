#!/usr/bin/env bash

set -Eeuo pipefail

cp main $APP_DIR/

sudo systemctl daemon-reload
sudo systemctl restart hometown-bot
