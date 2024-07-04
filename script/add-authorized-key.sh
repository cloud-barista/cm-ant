#!/bin/bash

PUBLIC_KEY=""
AUTHORIZED_KEYS_FILE="$HOME/.ssh/authorized_keys"
SSH_DIR="$HOME/.ssh"

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

# Check if the key already exists in authorized_keys
key_exists=false
while IFS= read -r line; do
    if [ "$line" = "$PUBLIC_KEY" ]; then
        key_exists=true
        break
    fi
done < "$AUTHORIZED_KEYS_FILE"

if [ "$key_exists" = true ]; then
    echo "Key already exists in authorized_keys file"
    exit 0
fi

# add public key to authorized_keys
echo "$PUBLIC_KEY" >> "$AUTHORIZED_KEYS_FILE"
echo "Key added to authorized_keys."
exit 0 