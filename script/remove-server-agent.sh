#!bin/bash

AGENT_WORK_DIR="${AGENT_WORK_DIR:=/opt/perfmon-agent}"
PROJECT_ROOT=$(pwd)
sudo rm -rf $AGENT_WORK_DIR && cd $PROJECT_ROOT
echo "remove server agent completely!!"