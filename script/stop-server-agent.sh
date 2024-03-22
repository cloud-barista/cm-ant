#!bin/bash

AGENT_WORK_DIR="${AGENT_WORK_DIR:=/opt/perfmon-agent}"
TCP_PORT="${TCP_PORT:=4444}"
PROJECT_ROOT=$(pwd)


cd $AGENT_WORK_DIR
echo "moved to $AGENT_WORK_DIR directory"

kill -15 $(lsof -t -i :$TCP_PORT)
