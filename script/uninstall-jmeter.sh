#!/bin/bash

echo "[CM-ANT] Uninstalling JMeter..."
set -e

# Default paths
JMETER_WORK_DIR="${JMETER_WORK_DIR:=/opt/ant/jmeter}"
JMETER_VERSION="${JMETER_VERSION:=5.6}"
JMETER_PATH="${JMETER_WORK_DIR}/apache-jmeter-${JMETER_VERSION}"

# Remove JMeter directory
sudo rm -rf "$JMETER_PATH"
echo "[CM-ANT] JMeter uninstalled."
