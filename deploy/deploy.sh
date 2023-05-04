#!/usr/bin/env bash

set -Eeuo pipefail

cp main $APP_DIR/

systemctl daemon-reload
systemctl restart hometown-bot
