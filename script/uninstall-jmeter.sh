#!/bin/bash

echo "[CM-ANT] JMeter Uninstallation"
set -e

# Base setup
JMETER_WORK_DIR=${JMETER_WORK_DIR:="/opt/ant/jmeter"}
JMETER_VERSION=${JMETER_VERSION:="5.6"}
JMETER_FOLDER="apache-jmeter-${JMETER_VERSION}"
JMETER_FULL_PATH="${JMETER_WORK_DIR}/${JMETER_FOLDER}"


sudo rm -rf "${JMETER_FULL_PATH}"
echo "[CM-ANT] Jmeter is completely uninstalled on system."
