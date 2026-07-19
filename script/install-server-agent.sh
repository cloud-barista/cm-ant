#!/bin/bash


set -e


AGENT_WORK_DIR="${AGENT_WORK_DIR:=/opt/perfmon-agent}"
AGENT_VERSION="2.2.1"
AGENT_FILE_NAME="ServerAgent-$AGENT_VERSION.zip"
AGENT_DOWNLOAD_URL="https://github.com/undera/perfmon-agent/releases/download/$AGENT_VERSION/$AGENT_FILE_NAME"
TCP_PORT="${TCP_PORT:=5555}"
AUTO_SHUTDOWN="${AUTO_SHUTDOWN:=}"

# Where this script says what it is doing. cm-ant reads it while waiting for the agent, so it
# can tell an install still working from one that has already given up - the two look identical
# from outside, and only one of them is worth waiting for. It lives outside the agent directory
# so a failure before that directory exists is still recorded.
STATE_FILE="${STATE_FILE:=/var/tmp/cm-ant-agent-install.state}"
LOG_FILE="${LOG_FILE:=/var/tmp/cm-ant-agent-install.log}"
WORK_OUT="${WORK_OUT:=/var/tmp/cm-ant-agent-install.out}"

# How often the long stages report where they have got to. The detail has to keep changing
# while real work is happening, because that is the only thing that separates a slow download
# from a dead one - a phase name alone says the same thing in both cases.
HEARTBEAT_INTERVAL="${HEARTBEAT_INTERVAL:=5}"


record() {
    local line
    line="$(date -u +%Y-%m-%dT%H:%M:%SZ)	$1	${2:-}"
    echo "$line" > "$STATE_FILE"
    echo "$line" >> "$LOG_FILE"
}


# heartbeat_start runs a reporter alongside a long stage. The command it is given must print
# something that advances with the work - apt's own output, the size of the file being
# downloaded - so that a value which stops changing means the work has stopped, not merely that
# the stage is long.
HB_PID=""
heartbeat_start() {
    local phase="$1" progress_cmd="$2"
    (
        while :; do
            sleep "$HEARTBEAT_INTERVAL"
            record "$phase" "$(eval "$progress_cmd" 2>/dev/null | tr -d '\r' | tr '\t\n' '  ' | tail -c 160)"
        done
    ) &
    HB_PID=$!
}

heartbeat_stop() {
    if [[ -n "$HB_PID" ]]; then
        kill "$HB_PID" 2>/dev/null || true
        wait "$HB_PID" 2>/dev/null || true
        HB_PID=""
    fi
}


on_failure() {
    local code=$?
    # Stop the reporter first, or it overwrites the failure with another progress line and the
    # install looks like it is still going.
    heartbeat_stop
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
    : > "$WORK_OUT"
    # apt's last line is the progress: fetching a package, unpacking, setting up. It stops
    # moving exactly when apt does.
    heartbeat_start dependencies "tail -n 1 '$WORK_OUT'"
    sudo apt-get update -y >> "$WORK_OUT" 2>&1
    sudo apt-get install -y wget unzip default-jre >> "$WORK_OUT" 2>&1
    heartbeat_stop
}


install_agent() {
    if [[ -z "$(ls -A "$AGENT_WORK_DIR")" ]]; then
        echo "[CM-ANT] Installing perfmon server agent"

        record downloading "$AGENT_DOWNLOAD_URL"
        # The size of the partial file is the honest measure of a download. A slow link keeps
        # raising it; a stalled one leaves it where it is, which is what cm-ant watches for.
        heartbeat_start downloading \
            "echo \"\$(stat -c%s '$AGENT_WORK_DIR/$AGENT_FILE_NAME' 2>/dev/null || echo 0) bytes of $AGENT_FILE_NAME\""
        sudo wget "$AGENT_DOWNLOAD_URL" -P "$AGENT_WORK_DIR" >> "$WORK_OUT" 2>&1
        heartbeat_stop

        record unpacking "$AGENT_FILE_NAME"
        sudo unzip "$AGENT_WORK_DIR/$AGENT_FILE_NAME" -d "$AGENT_WORK_DIR" >> "$WORK_OUT" 2>&1
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
