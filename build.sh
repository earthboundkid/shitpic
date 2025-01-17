#!/bin/bash

set -eux -o pipefail

# Get the directory that this script file is in
THIS_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

cd "$THIS_DIR"

GOOS=js GOARCH=wasm go build -o static/main.wasm ./cmd/spew
