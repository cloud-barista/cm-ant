#!bin/bash

WORKDIR="/opt/perfmon-agent"
AGENT_VERSION="2.2.1"
AGENT_FILE_NAME="ServerAgent-$AGENT_VERSION.zip"
AGENT_DOWNLOAD_URL="https://github.com/undera/perfmon-agent/releases/download/$AGENT_VERSION/$AGENT_FILE_NAME"

mkdir $WORKDIR
cd $WORKDIR

# install tools
apt-get install -y wget unzip default-jdk
wget $AGENT_DOWNLOAD_URL
unzip $AGENT_FILE_NAME -d .
rm $AGENT_FILE_NAME

# start perfmon server agent on port 4444
sh startAgent.sh
