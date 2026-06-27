#!/bin/bash

echo "[CM-ANT] Uninstalling JMeter..."
set -e

# Default paths
JMETER_WORK_DIR="${JMETER_WORK_DIR:=/opt/ant/jmeter}"
JMETER_VERSION="${JMETER_VERSION:=5.6}"
JMETER_PATH="${JMETER_WORK_DIR}/apache-jmeter-${JMETER_VERSION}"

# Remove JMeter directory
sudo rm -rf "$JMETER_PATH"

# Clean up the Temurin JRE tarball install and the system-wide JAVA_HOME profile
# entry placed by install-jmeter.sh. Both are safe to remove when absent.
sudo rm -rf "${JMETER_WORK_DIR}/jdk"
sudo rm -f /etc/profile.d/cm-ant-jmeter.sh

echo "[CM-ANT] JMeter uninstalled."
