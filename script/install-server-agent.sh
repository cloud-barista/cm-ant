#!bin/bash

AGENT_WORK_DIR="${AGENT_WORK_DIR:=/opt/perfmon-agent}"
AGENT_VERSION="2.2.1"
AGENT_FILE_NAME="ServerAgent-$AGENT_VERSION.zip"
AGENT_DOWNLOAD_URL="https://github.com/undera/perfmon-agent/releases/download/$AGENT_VERSION/$AGENT_FILE_NAME"

sudo mkdir -p "$AGENT_WORK_DIR"
cd "$AGENT_WORK_DIR"
# install tools
apt-get install -y wget unzip default-jre

if [ -z "$(ls -A "$AGENT_WORK_DIR")" ]; then
    echo "[CM-ANT] perfmon server agent is installing"
    sudo wget "$AGENT_DOWNLOAD_URL"
    sudo unzip "$AGENT_FILE_NAME" -d . && sudo rm "$AGENT_FILE_NAME"
    echo "[CM-ANT] agent installed successfully!!!!"
fi


