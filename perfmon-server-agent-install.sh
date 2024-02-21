#!bin/bash

WORKDIR="/opt/perfmon-agent"

mkdir $WORKDIR
cd $WORKDIR

# install tools
apk install -y wget unzip openjdk11
wget "https://github.com/undera/perfmon-agent/releases/download/2.2.1/ServerAgent-2.2.1.zip"
unzip "ServerAgent-2.2.1.zip" -d .
rm "ServerAgent-2.2.1.zip"

# start perfmon server agent on port 4444
sh startAgent.sh
