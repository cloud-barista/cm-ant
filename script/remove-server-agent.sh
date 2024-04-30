#!/bin/bash

set -e

AGENT_WORK_DIR="${AGENT_WORK_DIR:=/opt/perfmon-agent}"
TCP_PORT="${TCP_PORT:=4444}"

kill -15 "$(lsof -t -i :${TCP_PORT})"

sudo rm -rf $AGENT_WORK_DIR

echo "remove server agent completely!!"