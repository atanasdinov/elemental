#!/bin/bash
set -euo pipefail

MARKER_FILE="/var/lib/elemental/network-configuration-attempted"

if [ -f "$MARKER_FILE" ]; then
    echo "$(date): Marker file '$MARKER_FILE' found. Script already executed. Exiting." >&2
    exit 0
fi

./nmc apply --config-dir {{ .ConfigDir }} || {
echo "$(date): WARNING: Failed to apply static network configurations." >&2
}

touch "$MARKER_FILE"

echo "$(date): Network configuration attempt completed."
