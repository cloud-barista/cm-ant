#!/bin/bash
set -e

# Function to check and install a package
check_and_install() {
    PACKAGE_NAME=$1
    if command -v $PACKAGE_NAME &>/dev/null; then
        echo "$PACKAGE_NAME is already installed."
    else
        echo "Installing $PACKAGE_NAME..."
        apt-get update -y
        apt-get install -y $PACKAGE_NAME
    fi
}

check_and_install rsync
sleep 1
check_and_install ssh
