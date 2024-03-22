#!/bin/bash

echo "[CM-ANT] JMeter Installation"
set -e

# Base setup
JMETER_WORK_DIR=${JMETER_WORK_DIR:="${HOME}/jmeter"}
JMETER_VERSION=${JMETER_VERSION:="5.3"}
JMETER_FOLDER="apache-jmeter-${JMETER_VERSION}"
JMETER_FULL_PATH="${JMETER_WORK_DIR}/${JMETER_FOLDER}"
JMETER_INSTALL_URL="https://archive.apache.org/dist/jmeter/binaries/${JMETER_FOLDER}.tgz"

echo "${JMETER_WORK_DIR}"
echo ${JMETER_FOLDER}

mkdir -p "${JMETER_WORK_DIR}"
mkdir -p "${JMETER_WORK_DIR}/result"
mkdir -p "${JMETER_WORK_DIR}/test_plan"

echo "[CM-ANT] [Step 1/3] Installing default required tools..."
sudo apt-get update -y | sudo apt-get install -y wget default-jre

unzip_jmeter() {
  sudo tar -xf "${JMETER_FULL_PATH}.tgz" -C "${JMETER_WORK_DIR}" && sudo rm "${JMETER_FULL_PATH}.tgz"
  sudo rm -rf "${JMETER_FULL_PATH}/docs" "${JMETER_FULL_PATH}/printable_docs"
}

echo "[CM-ANT] [Step 2/3] Downloading and Extracting Apache JMeter..."
if [ -d "${JMETER_FULL_PATH}" ]; then
  echo "[CM-ANT] Jmeter is already installed."
elif [ -f "${JMETER_FULL_PATH}.tgz" ]; then
  echo "[CM-ANT] Jmeter gzip file is installed on ${JMETER_WORK_DIR}. Let's do remaining installation."
  unzip_jmeter
else
  # jmeter is not installed
  echo "[CM-ANT] JMeter is installing on path ${JMETER_WORK_DIR}"
  sudo wget "${JMETER_INSTALL_URL}" -P "${JMETER_WORK_DIR}"
  unzip_jmeter
fi

echo "[CM-ANT] [Step 3/3] Configuring JMeter..."
sudo chmod +x "${JMETER_FULL_PATH}/bin/jmeter.sh"
"${JMETER_FULL_PATH}"/bin/jmeter.sh --version

echo "[CM-ANT] Jmeter is completely installed on ${JMETER_FULL_PATH}"
