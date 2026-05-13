#!/bin/sh
set -e

REPO="varmiguemunoz/sprintos"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="sprintos"

get_latest_version() {
  curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' \
    | sed -E 's/.*"tag_name": "([^"]+)".*/\1/'
}

get_arch() {
  ARCH=$(uname -m)
  case $ARCH in
    x86_64)  echo "amd64" ;;
    aarch64) echo "arm64" ;;
    arm64)   echo "arm64" ;;
    *)
      echo "Unsupported architecture: $ARCH" >&2
      exit 1
      ;;
  esac
}

VERSION=$(get_latest_version)
ARCH=$(get_arch)
OS="linux"
BINARY="sprintos-${OS}-${ARCH}"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY}"

echo "Installing SprintOS ${VERSION} (${OS}/${ARCH})..."

curl -fsSL "$URL" -o "/tmp/${BINARY_NAME}"
chmod +x "/tmp/${BINARY_NAME}"

if [ -w "$INSTALL_DIR" ]; then
  mv "/tmp/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
else
  sudo mv "/tmp/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
fi

echo "SprintOS installed successfully."
echo "Run: sprintos start"
