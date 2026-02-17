#!/usr/bin/env bash
# shellcheck shell=bash
#
# scripts/lib/paths.sh

set -euo pipefail

source "$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)/lib.sh"

CMD_DIR="${CMD_DIR:-$ROOT_DIR/cmd/site}"
BINARY_NAME="${BINARY_NAME:-$(basename "$CMD_DIR")}"

# Guardrails
[[ "$BINARY_NAME" != "." && "$BINARY_NAME" != "/" && -n "$BINARY_NAME" ]] || {
  die "Refusing: invalid BINARY_NAME derived from CMD_DIR: '$CMD_DIR'"
}

TMP_BASE="${TMPDIR:-/tmp}"
TMP_BASE="${TMP_BASE%/}"

BIN_DIR="${BIN_DIR:-$TMP_BASE/$BINARY_NAME/bin}"
BIN_PATH="${BIN_PATH:-$BIN_DIR/$BINARY_NAME}"
