#!/usr/bin/env bash
# Script to generate OpenAPI-based stubs for Wili services.
#
# Usage:
#   ./scripts/generate.sh <service-name> <client|server>
#
# Examples:
#   ./scripts/generate.sh user server        # generate Go chi server stubs (oapi-codegen)
#   ./scripts/generate.sh wishlist client    # generate Go client (oapi-codegen)
#
# Notes:
#   • Service specs live at backend/services/<service-name>/openapi.yaml
#   • Generated Go server code is written to backend/services/<service-name>/gen
#   • Generated Go clients are written to backend/services/<service-name>/client
#
set -euo pipefail

if [[ $# -ne 2 ]]; then
  echo "Usage: $0 <service-name> <client|server>" >&2
  exit 1
fi

SERVICE_NAME="$1"
TYPE="$2"

# Resolve repository root (handles script called from anywhere)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

SPEC_FILE="$ROOT_DIR/services/$SERVICE_NAME/openapi.yaml"
if [[ ! -f "$SPEC_FILE" ]]; then
  echo "❌ Spec file not found at $SPEC_FILE" >&2
  exit 1
fi

case "$TYPE" in
  server)
    # Determine how to invoke oapi-codegen (binary or via "go run")
    if command -v oapi-codegen >/dev/null 2>&1; then
      OAPI="oapi-codegen"
    else
      echo "ℹ️  oapi-codegen binary not found – falling back to \"go run\" (module tool directive)."
      OAPI="go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest"
    fi

    OUT_DIR="$ROOT_DIR/services/$SERVICE_NAME/gen"
    PACKAGE_NAME="${SERVICE_NAME//-/_}_gen"

    echo "⚙️  Generating Go chi server stubs for $SERVICE_NAME via oapi-codegen …"
    rm -rf "$OUT_DIR"
    mkdir -p "$OUT_DIR"

    # shellcheck disable=SC2086
    $OAPI -o "$OUT_DIR/api.gen.go" -config "$ROOT_DIR/services/$SERVICE_NAME/cfg-gen-server.yaml" "$SPEC_FILE"
    echo "✅ Generated code written to $OUT_DIR"
    ;;

  client)
    # Determine how to invoke oapi-codegen (binary or via "go run")
    if command -v oapi-codegen >/dev/null 2>&1; then
      OAPI="oapi-codegen"
    else
      echo "ℹ️  oapi-codegen binary not found – falling back to \"go run\" (module tool directive)."
      OAPI="go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest"
    fi

    OUT_DIR="$ROOT_DIR/services/$SERVICE_NAME/client"
    PACKAGE_NAME="${SERVICE_NAME//-/_}_client"

    echo "⚙️  Generating Go client for $SERVICE_NAME via oapi-codegen …"
    rm -rf "$OUT_DIR"
    mkdir -p "$OUT_DIR"

    # shellcheck disable=SC2086
    $OAPI -o "$OUT_DIR/api.gen.go" -config "$ROOT_DIR/services/$SERVICE_NAME/cfg-gen-client.yaml" "$SPEC_FILE"

    echo "✅ Generated code written to $OUT_DIR"
    ;;

  *)
    echo "❌ Unknown generation type '$TYPE'. Use 'client' or 'server'." >&2
    exit 1
    ;;
esac
