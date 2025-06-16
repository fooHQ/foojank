#!/bin/bash

set -euo pipefail

DEVBOX_DOWNLOAD_URL="https://get.jetify.com/devbox"
FOOJANK_INSTALL_PATH="/usr/local/bin"

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


echo "[*] Foojank uses devbox to manage its dependencies."
echo "[*] I will now download and run Devbox installer from the project's website ($DEVBOX_DOWNLOAD_URL)."

curl -fsSL "$DEVBOX_DOWNLOAD_URL" | bash

devbox setup nix

echo "[*] Downloading Foojank ($(get_os)/$(get_arch))..."

tmp_file="$(mktemp)"
curl -fsSL "https://github.com/foohq/foojank/releases/latest/download/foojank-$(get_os)-$(get_arch)" > "$tmp_file"

echo "[*] Installing Foojank..."

$(command -v sudo || true) mv "$tmp_file" "$FOOJANK_INSTALL_PATH/foojank" && chmod +x "$FOOJANK_INSTALL_PATH/foojank"

echo "[*] Foojank has been installed!"
echo
echo "[!] Run 'foojank' to use it."
echo "[!] To read the documentation please visit https://foojank.com."
