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

echo "[CM-ANT] [Step 1/6] Installing Java (Eclipse Temurin JRE 21 LTS, no apt dependency)..."

# Tunable via env (cm-ant can override per-install). Defaults: Java 21 LTS, linux x64.
JAVA_VERSION="${JAVA_VERSION:-21.0.5+11}"
JAVA_ARCH="${JAVA_ARCH:-x64}"
JAVA_OS="${JAVA_OS:-linux}"
JAVA_HOME_PATH="${JMETER_WORK_DIR}/jdk"

# Adoptium URL pattern. Build path uses %2B for + (URL-encoded) and _ in filename.
JAVA_VERSION_URL="${JAVA_VERSION//+/%2B}"
JAVA_VERSION_FILE="${JAVA_VERSION//+/_}"
JAVA_TGZ="OpenJDK21U-jre_${JAVA_ARCH}_${JAVA_OS}_hotspot_${JAVA_VERSION_FILE}.tar.gz"
JAVA_URL="${JAVA_URL:-https://github.com/adoptium/temurin21-binaries/releases/download/jdk-${JAVA_VERSION_URL}/${JAVA_TGZ}}"

# Ensure wget or curl is available. Minimal stock AMIs may need a one-shot apt for wget.
if ! command -v wget >/dev/null 2>&1; then
  sudo apt-get install -y wget 2>/dev/null || true
fi
if ! command -v wget >/dev/null 2>&1 && ! command -v curl >/dev/null 2>&1; then
  echo "[CM-ANT][ERROR] Neither wget nor curl is available" >&2
  exit 1
fi

# Download and extract Temurin JRE.
sudo mkdir -p "${JAVA_HOME_PATH}"
if command -v wget >/dev/null 2>&1; then
  sudo wget -qO /tmp/jre.tgz "${JAVA_URL}"
else
  sudo curl -fsSL -o /tmp/jre.tgz "${JAVA_URL}"
fi
sudo tar -xzf /tmp/jre.tgz -C "${JAVA_HOME_PATH}" --strip-components=1
sudo rm -f /tmp/jre.tgz

# Safety net (1) — system-wide login shells pick up JAVA_HOME automatically.
sudo tee /etc/profile.d/cm-ant-jmeter.sh > /dev/null <<EOF
export JAVA_HOME="${JAVA_HOME_PATH}"
export PATH="\${JAVA_HOME}/bin:\${PATH}"
EOF
sudo chmod 644 /etc/profile.d/cm-ant-jmeter.sh

# Safety net (2) — current install session uses Temurin for Steps 5 and 6.
export JAVA_HOME="${JAVA_HOME_PATH}"
export PATH="${JAVA_HOME}/bin:${PATH}"

"${JAVA_HOME}/bin/java" -version 2>&1 || { echo "[CM-ANT][ERROR] java installation failed"; exit 1; }

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

  # Safety net (3) — non-interactive ssh (cb-tumblebug POST /cmd/infra) does not
  # source /etc/profile.d/*.sh, so jmeter.sh would not find java otherwise.
  # Inject JAVA_HOME export right after the shebang; guarded by a marker so the
  # injection is idempotent across re-installs and JMeter upgrades.
  if ! sudo grep -q "# CM-ANT: JAVA_HOME injected" "${JMETER_FULL_PATH}/bin/jmeter.sh"; then
    sudo sed -i "2i # CM-ANT: JAVA_HOME injected\\nexport JAVA_HOME=\"${JAVA_HOME_PATH}\"\\nexport PATH=\"\${JAVA_HOME}/bin:\${PATH}\"" "${JMETER_FULL_PATH}/bin/jmeter.sh"
  fi

  JAVA_HOME="${JAVA_HOME_PATH}" "${JMETER_FULL_PATH}/bin/jmeter.sh" --version
}

# Execute steps
install_cmd_runner
install_plugin_manager
install_required_plugins
configure_jmeter

echo "[CM-ANT] JMeter installation completed successfully at ${JMETER_FULL_PATH}"
