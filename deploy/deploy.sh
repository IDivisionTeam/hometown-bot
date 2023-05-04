#!/usr/bin/env bash

set -Eeuo pipefail

cp main $APP_DIR/

systemctl enable $APP_DIR/app.service
systemctl restart hometown-bot
systemctl daemon-reload
