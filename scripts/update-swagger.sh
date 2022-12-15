#!/usr/bin/env bash

help() {
  cat << EOF
Update the vendored Swagger UI distribution to given version.

Usage:
  $(basename "$0") VERSION

Args:
  VERSION: Version of Swagger UI to download.

Examples:
  $(basename "$0") 4.15.5
  $(basename "$0") 4.15.2

Ref:
  https://github.com/swagger-api/swagger-ui
EOF
}

case $1 in -h | --help | help) help && exit ;; esac

SCRIPT_DIR=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_DIR" || exit 1

# ------------------------------------------------------------------------------

if [[ ! $1 ]]; then
  echo "Missing parameter: VERSION" && help && exit 1
fi

if printf "%s" "$1" | grep --quiet --invert-match '^[0-9.]*$'; then
  echo "Invalid argument: VERSION='$1'" && help && exit 1
fi

VERSION=$1

# ------------------------------------------------------------------------------

set -euo pipefail

mkdir -p tmp
rm -rf "tmp/swagger-ui-$VERSION.zip"
rm -rf "tmp/swagger-ui-$VERSION"

# Download archive.
OPTS=(--no-progress-meter --location --fail)
URL="https://github.com/swagger-api/swagger-ui/archive/refs/tags/v$VERSION.zip"
if ! curl "${OPTS[@]}" "$URL" > "tmp/swagger-ui-$VERSION.zip"; then
  echo "Download failed." && exit 1
fi

# Unzip archive.
unzip -q "tmp/swagger-ui-$VERSION.zip" -d tmp

# Override file.
cp -f assets/swagger-initializer.js "tmp/swagger-ui-$VERSION/dist/swagger-initializer.js"

# Add version indicator file.
printf '%s\n' "$VERSION" > "tmp/swagger-ui-$VERSION/dist/version.txt"

# Move archive dist into place.
rm -rf swagger-ui
cp -r "tmp/swagger-ui-$VERSION/dist" "swagger-ui"
