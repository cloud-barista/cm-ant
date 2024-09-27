#!/bin/bash

set -e

PUBLIC_KEY=""
AUTHORIZED_KEYS_FILE="$HOME/.ssh/authorized_keys"
SSH_DIR="$HOME/.ssh"


if [ -z "$PUBLIC_KEY" ]; then
    echo "No public key provided. Usage: $0 \"<public_key>\""
    exit 1
fi

# create .ssh directory if does not exist
if [ ! -d "$SSH_DIR" ]; then
    mkdir "$SSH_DIR"
    chmod 700 "$SSH_DIR"
fi

# create authorized_keys file if does not exist
if [ ! -f "$AUTHORIZED_KEYS_FILE" ]; then
    touch "$AUTHORIZED_KEYS_FILE"
    chmod 600 "$AUTHORIZED_KEYS_FILE"
fi

# Temporary file to store the public key
TEMP_KEY_FILE=$(mktemp)
echo "$PUBLIC_KEY" > "$TEMP_KEY_FILE"

# Generate the fingerprint of the provided public key
PUBLIC_KEY_FINGERPRINT=$(ssh-keygen -lf "$TEMP_KEY_FILE" | awk '{print $2}')

# Check if the key already exists in authorized_keys
key_exists=false
while IFS= read -r line; do
    echo "$line" > "$TEMP_KEY_FILE"
    EXISTING_KEY_FINGERPRINT=$(ssh-keygen -lf "$TEMP_KEY_FILE" | awk '{print $2}')
    if [ "$EXISTING_KEY_FINGERPRINT" = "$PUBLIC_KEY_FINGERPRINT" ]; then
        key_exists=true
        break
    fi
done < "$AUTHORIZED_KEYS_FILE"

# Clean up temporary file
rm "$TEMP_KEY_FILE"

if [ "$key_exists" = true ]; then
    echo "Key already exists in authorized_keys file"
    exit 0
fi

# add public key to authorized_keys
echo "$PUBLIC_KEY" >> "$AUTHORIZED_KEYS_FILE"
echo "Key added to authorized_keys."
exit 0 