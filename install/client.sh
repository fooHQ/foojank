#!/bin/bash

set -euo pipefail

DEVBOX_DOWNLOAD_URL="https://get.jetify.com/devbox"
INSTALL_PATH="/usr/local/bin"

get_os() {
        uname -s | tr '[:upper:]' '[:lower:]'
}

get_arch() {
        case "$(uname -m)" in
        x86_64)
            echo "amd64"
            ;;
        aarch64 | arm64)
            echo "arm64"
            ;;
        *)
            echo "Unsupported architecture"
            exit 1
            ;;
        esac
}

echo "[!] Foojank uses Devbox to manage its dependencies."

echo "[*] Downloading Devbox installer from the project's website ($DEVBOX_DOWNLOAD_URL)."
curl -fsSL "$DEVBOX_DOWNLOAD_URL" | FORCE=1 bash

devbox setup nix

echo "[*] Downloading Foojank ($(get_os)/$(get_arch))..."
tmp_file="$(mktemp)"
curl -fsSL -o "$tmp_file" "https://github.com/foohq/foojank/releases/latest/download/foojank-$(get_os)-$(get_arch)"

echo "[*] Installing Foojank client..."
$(command -v sudo || true) install "$tmp_file" "$INSTALL_PATH/foojank"

echo "[*] Foojank has been installed!"
echo

echo "[!] Run 'foojank' to use it."
echo "[!] Visit https://foojank.com to read the documentation."
