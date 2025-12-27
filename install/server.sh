#!/bin/bash

INSTALL_PATH="/usr/sbin"
NATS_DOWNLOAD_URL="https://binaries.nats.dev/nats-io/nats-server/v2@latest"
NSC_DOWNLOAD_URL="https://binaries.nats.dev/nats-io/nsc/v2@latest"
SYSTEMD_UNIT_DOWNLOAD_URL="https://raw.githubusercontent.com/nats-io/nats-server/refs/heads/main/util/nats-server.service"
NATS_CONFIG_PATH="/etc/nats-server.conf"
NATS_SYSTEMD_PATH="/etc/systemd/system/nats-server.service"
NATS_OPERATOR_NAME="nats-prod"
NATS_PATH="/opt/nats"
NATS_DEFAULT_CONFIG=$(cat <<EOF
# The default configuration prefers WebSocket over NATS' TCP-based protocol.
# If you want the clients to use TCP-based protocol, rebind it to a non-local address.
host: 0.0.0.0
port: 4222

# Uncomment and change according to your configuration.
# For more information visit: https://docs.nats.io/running-a-nats-service/configuration/securing_nats/tls
#tls {
    # Uncomment and change according to your configuration.
    # cert_file: "/etc/letsencrypt/live/example.com/fullchain.pem"
    # key_file: "/etc/letsencrypt/live/example.com/privkey.pem"
    # handshake_first: true
#}

# For more information visit: https://docs.nats.io/running-a-nats-service/configuration/websocket/websocket_conf
#websocket {
#    host: 0.0.0.0
#    port: 8443
#    compression: true

    #tls {
        # Uncomment and change according to your configuration.
        # cert_file: "/etc/letsencrypt/live/example.com/fullchain.pem"
        # key_file: "/etc/letsencrypt/live/example.com/privkey.pem"
    #}
#}

jetstream {
    store_dir: $NATS_PATH
}

EOF
)

install_nats() {
    echo "[+] Downloading NATS server installer ($NATS_DOWNLOAD_URL)."
    curl -fsSL "$NATS_DOWNLOAD_URL" | PREFIX="$INSTALL_PATH" sh
}

install_nsc() {
    echo "[+] Downloading nsc installer ($NSC_DOWNLOAD_URL)."
    curl -sf "$NSC_DOWNLOAD_URL" | PREFIX="$INSTALL_PATH" sh
}

install_nats_systemd() {
    echo "[+] Downloading systemd unit file ($SYSTEMD_UNIT_DOWNLOAD_URL)."
    if [ -f "$NATS_SYSTEMD_PATH" ]; then
        echo "[!] Unit file already exists in '$NATS_SYSTEMD_PATH'."
        return
    fi
    tmp_file="$(mktemp)"
    curl -fsSL -o "$tmp_file" "$SYSTEMD_UNIT_DOWNLOAD_URL"
    $(command -v sudo || true) mv "$tmp_file" "$NATS_SYSTEMD_PATH"

    echo "[+] Reloading systemd daemon."
    $(command -v sudo || true) systemctl daemon-reload

    echo "[+] Enabling NATS server."
    $(command -v sudo || true) systemctl enable nats-server.service
}

install_nats_group() {
    echo "[+] Adding nats group."
    $(command -v sudo || true) groupadd --force --system nats
}

install_nats_user() {
    echo "[+] Adding nats user."
    id -u nats > /dev/null 2>&1
    ret=$?
    if [ $ret -eq 0 ]; then
        echo "[!] User 'nats' already exists."
        return
    fi
    $(command -v sudo || true) useradd --system -g nats -s /usr/sbin/nologin -c "NATS service user" nats
}

configure_nats_operator() {
    echo "[+] Creating NATS Operator."
    nsc describe operator --name "$NATS_OPERATOR_NAME" >/dev/null 2>&1
    ret=$?
    if [ $ret -eq 0 ]; then
        echo "[!] NATS Operator '$NATS_OPERATOR_NAME' already exists."
        return
    fi
    $(command -v sudo || true) nsc add operator --sys --name "$NATS_OPERATOR_NAME"
}

configure_nats() {
    echo "[+] Configuring NATS server."
    if [ -f "$NATS_CONFIG_PATH" ]; then
        echo "[!] NATS configuration file already exists in '$NATS_CONFIG_PATH'."
        return
    fi
    echo "$NATS_DEFAULT_CONFIG" | $(command -v sudo || true) tee "$NATS_CONFIG_PATH"
    $(command -v sudo || true) nsc generate config --nats-resolver | $(command -v sudo || true) tee -a "$NATS_CONFIG_PATH"
    # Replace resolver's default JWT directory path.
    $(command -v sudo || true) sed -i "s|dir: './jwt'|dir: '$NATS_PATH/jwt'|" "$NATS_CONFIG_PATH"
    $(command -v sudo || true) mkdir -p "$NATS_PATH"
    $(command -v sudo || true) chown nats:nats "$NATS_PATH"
}

install_nats
install_nsc
install_nats_systemd
install_nats_group
install_nats_user
configure_nats_operator
configure_nats

echo
echo "[*] Installation was successful!"
echo "[*] Default configuration has been created in $NATS_CONFIG_PATH. You must now generate a TLS certificate for the server and enable it in $NATS_CONFIG_PATH/server.conf."
echo "[*] For more information please visit NATS official documentation https://docs.nats.io/running-a-nats-service/configuration."
echo "[*] To start NATS server run: systemctl start nats-server.service"
