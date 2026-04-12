#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
FRONTEND_DIR="$ROOT_DIR/frontend"
TARGET_DIR="/var/www/wuwa-echo"

cd "$FRONTEND_DIR"
npm run build-only

install -d "$TARGET_DIR"
rsync -a --delete "$FRONTEND_DIR/dist/" "$TARGET_DIR/"
