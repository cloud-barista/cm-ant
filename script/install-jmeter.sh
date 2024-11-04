#!/bin/bash

set -e
echo "[CM-ANT] JMeter Installation"

# Base setup
JMETER_WORK_DIR=${JMETER_WORK_DIR:="/opt/ant/jmeter"}
JMETER_VERSION=${JMETER_VERSION:="5.6"}
JMETER_FOLDER="apache-jmeter-${JMETER_VERSION}"
JMETER_FULL_PATH="${JMETER_WORK_DIR}/${JMETER_FOLDER}"
JMETER_INSTALL_URL="https://archive.apache.org/dist/jmeter/binaries/${JMETER_FOLDER}.tgz"

echo "This is jmeter working directory >>> ${JMETER_WORK_DIR}"
echo "This is jmeter folder name >>> ${JMETER_FOLDER}"

# Create directories if they don't exist
create_directory() {
  local dir=$1
  if [ ! -e "$dir" ]; then
    sudo mkdir -p "$dir"
    echo "Created directory: $dir"
  fi
}

create_directory "${JMETER_WORK_DIR}"
create_directory "${JMETER_WORK_DIR}/result"
create_directory "${JMETER_WORK_DIR}/test_plan"

echo "[CM-ANT] [Step 1/6] Installing required tools..."
sudo apt install -y software-properties-common
sudo add-apt-repository universe -y
sudo apt-get update -y
sudo apt-get install -y wget openjdk-17-jre

# Function to extract JMeter
unzip_jmeter() {
  sudo tar -xf "${JMETER_WORK_DIR}/${JMETER_FOLDER}.tgz" -C "${JMETER_WORK_DIR}" &&
    sudo rm "${JMETER_WORK_DIR}/${JMETER_FOLDER}.tgz"
  sudo rm -rf "${JMETER_FULL_PATH}/docs" "${JMETER_FULL_PATH}/printable_docs"
}

echo "[CM-ANT] [Step 2/6] Downloading and Extracting Apache JMeter..."
if [ -d "${JMETER_FULL_PATH}" ]; then
  echo "[CM-ANT] JMeter is already installed."
else
  if [ ! -f "${JMETER_WORK_DIR}/${JMETER_FOLDER}.tgz" ]; then
    echo "[CM-ANT] Downloading JMeter..."
    sudo wget "${JMETER_INSTALL_URL}" -P "${JMETER_WORK_DIR}"
  fi
  echo "[CM-ANT] Extracting JMeter..."
  unzip_jmeter
fi

sudo chmod -R 777 "${JMETER_WORK_DIR}"

# Install CMD Runner
install_cmd_runner() {
  local version="2.2.1"
  local jar="cmdrunner-${version}.jar"
  if [ ! -f "${JMETER_FULL_PATH}/lib/${jar}" ]; then
    echo "[CM-ANT] [Step 3/6] Installing CMD Runner..."
    wget "https://repo1.maven.org/maven2/kg/apc/cmdrunner/${version}/${jar}" -P "${JMETER_WORK_DIR}" &&
      sudo chmod +x "${JMETER_WORK_DIR}/${jar}" &&
      sudo mv "${JMETER_WORK_DIR}/${jar}" "${JMETER_FULL_PATH}/lib/"
  fi
}

# Install Plugin Manager
install_plugin_manager() {
  local version="1.6"
  local jar="jmeter-plugins-manager-${version}.jar"
  if [ ! -f "${JMETER_FULL_PATH}/lib/ext/${jar}" ]; then
    echo "[CM-ANT] [Step 4/6] Installing Plugin Manager..."
    wget "https://repo1.maven.org/maven2/kg/apc/jmeter-plugins-manager/${version}/${jar}" -P "${JMETER_WORK_DIR}" &&
      sudo chmod +x "${JMETER_WORK_DIR}/${jar}" &&
      sudo mv "${JMETER_WORK_DIR}/${jar}" "${JMETER_FULL_PATH}/lib/ext/"
  fi
}

# Install required plugins
install_required_plugins() {
  echo "[CM-ANT] [Step 5/6] Installing required plugins for load testing..."
  java -jar "${JMETER_FULL_PATH}/lib/cmdrunner-2.2.1.jar" --tool org.jmeterplugins.repository.PluginManagerCMD install jpgc-perfmon,jpgc-casutg
}

# Configure JMeter
configure_jmeter() {
  echo "[CM-ANT] [Step 6/6] Configuring JMeter..."
  sudo chmod +x "${JMETER_FULL_PATH}/bin/jmeter.sh"
  "${JMETER_FULL_PATH}/bin/jmeter.sh" --version
}

# Execute steps
install_cmd_runner
install_plugin_manager
install_required_plugins
configure_jmeter

echo "[CM-ANT] JMeter installation completed successfully at ${JMETER_FULL_PATH}"
