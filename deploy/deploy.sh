#!/usr/bin/env bash

set -Eeuo pipefail

chmod +x main
cp main $APP_DIR/

systemctl enable ./app.service
systemctl restart hometown-bot
systemctl daemon-reload
