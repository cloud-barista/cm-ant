#!/bin/bash

set -e

# Set default values for variables if not provided
AGENT_WORK_DIR="${AGENT_WORK_DIR:=/opt/perfmon-agent}"
TCP_PORT="${TCP_PORT:=5555}"

# Check if a process is using the specified TCP port and terminate it if found
if lsof -t -i :"${TCP_PORT}" &>/dev/null; then
    echo "Terminating process on port ${TCP_PORT}..."
    kill -15 "$(lsof -t -i :"${TCP_PORT}")"
else
    echo "No process found on port ${TCP_PORT}."
fi

# Remove the agent work directory with sudo privileges
echo "Removing agent work directory: ${AGENT_WORK_DIR}"
sudo rm -rf "${AGENT_WORK_DIR}"

echo "Server agent removed completely!"
