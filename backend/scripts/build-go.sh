#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."
mkdir -p bin
mkdir -p .gocache
env GOPROXY=off GOSUMDB=off GOCACHE="$(pwd)/.gocache" go build -o ./bin/wuwa-echo-backend ./cmd/wuwa-echo-backend
