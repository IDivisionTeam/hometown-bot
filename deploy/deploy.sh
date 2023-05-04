#!/usr/bin/env bash

set -Eeuo pipefail

echo $APP_DIR > test.txt
cp main $APP_DIR/

systemctl enable ./app.service
systemctl restart hometown-bot
systemctl daemon-reload
