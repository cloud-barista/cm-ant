#!/bin/bash
set -e

AGENT_WORK_DIR="${AGENT_WORK_DIR:="/opt/perfmon-agent"}"
AGENT_VERSION="2.2.1"
AGENT_FILE_NAME="ServerAgent-$AGENT_VERSION.zip"
AGENT_DOWNLOAD_URL="https://github.com/undera/perfmon-agent/releases/download/$AGENT_VERSION/$AGENT_FILE_NAME"

TCP_PORT="${TCP_PORT:=4444}"
AUTO_SHUTDOWN="${AUTO_SHUTDOWN:=--auto-shutdown}"

sudo mkdir -p "$AGENT_WORK_DIR"
sudo apt-get update -y
sudo apt-get install -y wget unzip default-jre

if [ -z "$(ls -A "$AGENT_WORK_DIR")" ]; then
    echo "[CM-ANT] perfmon server agent is installing"
    sudo wget "${AGENT_DOWNLOAD_URL}" -P "${AGENT_WORK_DIR}"
    sudo unzip "${AGENT_WORK_DIR}/${AGENT_FILE_NAME}" -d "${AGENT_WORK_DIR}"
    sudo rm "${AGENT_WORK_DIR}/${AGENT_FILE_NAME}"
    echo "[CM-ANT] agent installed successfully!!!!"
fi

if [ $? -eq 0 ] && [ -e "${AGENT_WORK_DIR}/startAgent.sh" ]; then
   nohup "${AGENT_WORK_DIR}/startAgent.sh" --udp-port 0 --tcp-port $TCP_PORT $AUTO_SHUTDOWN &
else
    echo "something went wrong.........."
fi


