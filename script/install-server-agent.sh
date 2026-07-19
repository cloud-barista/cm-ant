#!/bin/bash


set -e


AGENT_WORK_DIR="${AGENT_WORK_DIR:=/opt/perfmon-agent}"
AGENT_VERSION="2.2.1"
AGENT_FILE_NAME="ServerAgent-$AGENT_VERSION.zip"
AGENT_DOWNLOAD_URL="https://github.com/undera/perfmon-agent/releases/download/$AGENT_VERSION/$AGENT_FILE_NAME"
TCP_PORT="${TCP_PORT:=5555}"
AUTO_SHUTDOWN="${AUTO_SHUTDOWN:=}"

# Where this script says what it is doing. cm-ant reads it while waiting for the agent, so it
# can tell an install still fetching packages from one that has already given up - the two look
# identical from outside, and only one of them is worth waiting for. It lives outside the agent
# directory so a failure before that directory exists is still recorded.
STATE_FILE="${STATE_FILE:=/var/tmp/cm-ant-agent-install.state}"
LOG_FILE="${LOG_FILE:=/var/tmp/cm-ant-agent-install.log}"


record() {
    local line
    line="$(date -u +%Y-%m-%dT%H:%M:%SZ)	$1	${2:-}"
    echo "$line" | sudo tee "$STATE_FILE" >/dev/null
    echo "$line" | sudo tee -a "$LOG_FILE" >/dev/null
}


on_failure() {
    local code=$?
    record failed "exit $code at line $1: $BASH_COMMAND"
    exit $code
}
trap 'on_failure $LINENO' ERR


terminate_existing_process() {
    local pid
    pid="$(lsof -t -i :$TCP_PORT)" || true
    if [[ -n "$pid" ]]; then
        kill -15 "$pid"
        echo "Terminated process using port $TCP_PORT"
    fi
}


install_dependencies() {
    record dependencies "installing wget, unzip and a jre"
    sudo apt-get update -y
    sudo apt-get install -y wget unzip default-jre
}


install_agent() {
    if [[ -z "$(ls -A "$AGENT_WORK_DIR")" ]]; then
        echo "[CM-ANT] Installing perfmon server agent"
        record downloading "$AGENT_DOWNLOAD_URL"
        sudo wget "$AGENT_DOWNLOAD_URL" -P "$AGENT_WORK_DIR"
        record unpacking "$AGENT_FILE_NAME"
        sudo unzip "$AGENT_WORK_DIR/$AGENT_FILE_NAME" -d "$AGENT_WORK_DIR"
        sudo rm "$AGENT_WORK_DIR/$AGENT_FILE_NAME"
        echo "[CM-ANT] Agent installed successfully"
    else
        record present "agent already unpacked in $AGENT_WORK_DIR"
    fi
}


start_agent() {
    if [[ -e "${AGENT_WORK_DIR}/startAgent.sh" ]]; then
        record starting "tcp port $TCP_PORT"
        nohup "${AGENT_WORK_DIR}/startAgent.sh" --udp-port 0 --tcp-port "$TCP_PORT" "$AUTO_SHUTDOWN" > /dev/null 2>&1 &
        echo "Agent started successfully in the background."
        record started "agent started in the background on port $TCP_PORT"
    else
        # Reached when the archive unpacked into an unexpected shape. The script has nothing
        # left to try, and saying so is what stops cm-ant from waiting out its ceiling.
        record failed "startAgent.sh not found under $AGENT_WORK_DIR"
        echo "Failed to start the agent. startAgent.sh not found."
        exit 1
    fi
}


main() {
    record preparing "terminating anything on port $TCP_PORT"
    terminate_existing_process
    sudo mkdir -p "$AGENT_WORK_DIR"
    install_dependencies
    install_agent
    start_agent
}


main
