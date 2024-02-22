#!bin/bash

AGENT_WORK_DIR
TCP_PORT
AUTO_SHUTDOWN
PROJECT_ROOT=$(pwd)

if [ -z "$(ls -A $AGENT_WORK_DIR)" ]; then
    echo "[CM-ANT] please install server agent first!"
    exit
fi

"$AGENT_WORK_DIR/startAgent.sh" --udp-port 0 --tcp-port $TCP_PORT $AUTO_SHUTDOWN