#!bin/bash

AGENT_WORK_DIR="${AGENT_WORK_DIR:=/opt/perfmon-agent}"
TCP_PORT="${TCP_PORT:=4444}"
AUTO_SHUTDOWN="${AUTO_SHUTDOWN:=--auto-shutdown}"
PROJECT_ROOT=$(pwd)

if [ -z "$(ls -A $AGENT_WORK_DIR)" ]; then
    echo "[CM-ANT] please install server agent first!"
    exit
fi

"$AGENT_WORK_DIR/startAgent.sh" --udp-port 0 --tcp-port $TCP_PORT $AUTO_SHUTDOWN