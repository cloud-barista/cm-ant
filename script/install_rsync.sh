#!/bin/bash

set -e

# Check if rsync is already installed
if command -v rsync &>/dev/null; then
    echo "rsync is already installed."
    exit 0
fi

# Install rsync
echo "Installing rsync..."
apt-get update
apt-get install -y rsync