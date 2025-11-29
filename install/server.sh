#!/bin/bash

set -euo pipefail

INSTALL_PATH="/usr/local/bin"
NATS_DOWNLOAD_URL="https://binaries.nats.dev/nats-io/nats-server/v2@latest"
NSC_DOWNLOAD_URL="https://binaries.nats.dev/nats-io/nsc/v2@latest"
SYSTEMD_UNIT_DOWNLOAD_URL="https://raw.githubusercontent.com/nats-io/nats-server/refs/heads/main/util/nats-server.service"
NATS_CONFIG_PATH="/etc/nats-server"
NATS_DEFAULT_CONFIG=<<EOF
# The default configuration prefers WebSocket over NATS' TCP-based protocol and does not expose it by default.
# If you want the clients to use TCP-based protocol, you must rebind it to a non-local address.
host: 127.0.0.1
port: 4222

# Uncomment and change according to your configuration.
# For more information visit: https://docs.nats.io/running-a-nats-service/configuration/securing_nats/tls
# cert_file: "/etc/letsencrypt/live/example.com/fullchain.pem"
# key_file: "/etc/letsencrypt/live/example.com/privkey.pem"
# handshake_first: true

websocket {
    # For more information visit: https://docs.nats.io/running-a-nats-service/configuration/websocket/websocket_conf
    host: 0.0.0.0
    port: 443
    compression: true

    tls {
        # Uncomment and change according to your configuration.
        # cert_file: "/etc/letsencrypt/live/example.com/fullchain.pem"
        # key_file: "/etc/letsencrypt/live/example.com/privkey.pem"
    }
}

jetstream: true

include ./auth.conf
EOF

echo "[*] Downloading NATS server installer from the project's website ($NATS_DOWNLOAD_URL)."
curl -fsSL "$NATS_DOWNLOAD_URL" | PREFIX="$INSTALL_PATH" sh

echo "[*] Downloading nsc installer from the project's website ($NSC_DOWNLOAD_URL)."
curl -sf "$NSC_DOWNLOAD_URL" | PREFIX="$INSTALL_PATH" sh

echo "[*] Downloading systemd unit file ($SYSTEMD_UNIT_DOWNLOAD_URL)."
tmp_file="$(mktemp)"
curl -fsSL -o "$tmp_file" "$SYSTEMD_UNIT_DOWNLOAD_URL"
$(command -v sudo || true) mv "$tmp_file" "/etc/systemd/system/nats-server.service"

echo "[*] Adding nats group."
$(command -v sudo || true) groupadd --system nats

echo "[*] Adding nats user."
$(command -v sudo || true) useradd --system -g nats -s /usr/sbin/nologin -c "NATS service user" nats

echo "[*] Configuring NATS server."
$(command -v sudo || true) mkdir "$NATS_CONFIG_PATH"
echo "$NATS_DEFAULT_CONFIG" | $(command -v sudo || true) tee "$NATS_CONFIG_PATH/server.conf"
$(command -v sudo || true) nsc add operator --sys --name "nats-prod"
$(command -v sudo || true) nsc edit operator --account-jwt-server-url "nats://127.0.0.1"
$(command -v sudo || true) nsc generate config --nats-resolver --config-file "$NATS_CONFIG_PATH/auth.conf"

echo "[*] Reloading systemd daemon."
$(command -v sudo || true) systemctl daemon-reload

echo "[*] Enabling NATS server."
$(command -v sudo || true) systemctl enable nats-server.service

echo "[*] NATS server has been installed!"
echo

echo "[!] Default configuration has been created in $NATS_CONFIG_PATH. You must now generate a TLS certificate for the server and enable it in $NATS_CONFIG_PATH/server.conf."
echo "[!] For more information please visit NATS official documentation https://docs.nats.io/running-a-nats-service/configuration."
echo "[!] To start NATS server run: systemctl start nats-server.service"
