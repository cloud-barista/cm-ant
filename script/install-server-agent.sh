#!/bin/bash


set -e


AGENT_WORK_DIR="${AGENT_WORK_DIR:=/opt/perfmon-agent}"
AGENT_VERSION="2.2.1"
AGENT_FILE_NAME="ServerAgent-$AGENT_VERSION.zip"
AGENT_DOWNLOAD_URL="https://github.com/undera/perfmon-agent/releases/download/$AGENT_VERSION/$AGENT_FILE_NAME"
TCP_PORT="${TCP_PORT:=5555}"
AUTO_SHUTDOWN="${AUTO_SHUTDOWN:=}"


terminate_existing_process() {
    local pid
    pid="$(lsof -t -i :$TCP_PORT)" || true
    if [[ -n "$pid" ]]; then
        kill -15 "$pid"
        echo "Terminated process using port $TCP_PORT"
    fi
}


install_dependencies() {
    sudo apt-get update -y
    sudo apt-get install -y wget unzip default-jre
}


install_agent() {
    if [[ -z "$(ls -A "$AGENT_WORK_DIR")" ]]; then
        echo "[CM-ANT] Installing perfmon server agent"
        sudo wget "$AGENT_DOWNLOAD_URL" -P "$AGENT_WORK_DIR"
        sudo unzip "$AGENT_WORK_DIR/$AGENT_FILE_NAME" -d "$AGENT_WORK_DIR"
        sudo rm "$AGENT_WORK_DIR/$AGENT_FILE_NAME"
        echo "[CM-ANT] Agent installed successfully"
    fi
}


start_agent() {
    if [[ -e "${AGENT_WORK_DIR}/startAgent.sh" ]]; then
        nohup "${AGENT_WORK_DIR}/startAgent.sh" --udp-port 0 --tcp-port "$TCP_PORT" "$AUTO_SHUTDOWN" > /dev/null 2>&1 &
        echo "Agent started successfully in the background."
    else
        echo "Failed to start the agent. startAgent.sh not found."
    fi
}


main() {
    terminate_existing_process
    sudo mkdir -p "$AGENT_WORK_DIR"
    install_dependencies
    install_agent
    start_agent
}


main
