#!bin/bash

AGENT_WORK_DIR
TCP_PORT
PROJECT_ROOT=$(pwd)


cd $AGENT_WORK_DIR
echo "moved to $AGENT_WORK_DIR directory"

kill -9 $(lsof -t -i :$TCP_PORT)