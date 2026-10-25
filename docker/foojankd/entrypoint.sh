#!/bin/sh

set -eu

if [ -z "${FOOJANKD_ACCOUNT:-}" ]; then
    export FOOJANKD_ACCOUNT=main
fi

if ! /usr/local/bin/foojankd account describe "$FOOJANKD_ACCOUNT" >/dev/null 2>&1; then
    /usr/local/bin/foojankd account create --name "$FOOJANKD_ACCOUNT"
fi

# Import is not idempotent: an existing role makes the command fail.
# A restart must still reach user setup and foojankd start.
/usr/local/bin/foojankd role import --file /usr/share/foojank/roles.json || true

if [ -n "${FOOJANKD_ADMIN_NAME:-}" ] && [ -n "${FOOJANKD_ADMIN_PUBLIC_KEY:-}" ]; then
    if ! /usr/local/bin/foojankd user describe "$FOOJANKD_ADMIN_NAME" >/dev/null 2>&1; then
        /usr/local/bin/foojankd user create \
            --name "$FOOJANKD_ADMIN_NAME" \
            --public-key "$FOOJANKD_ADMIN_PUBLIC_KEY" \
            --role admin
    fi
fi

exec /usr/local/bin/foojankd start "$@"
