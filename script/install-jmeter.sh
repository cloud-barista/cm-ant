#!/bin/bash
set -e

# Base setup
JMETER_WORK_DIR=${JMETER_WORK_DIR:="/opt/jmeter"}
JMETER_VERSION=${JMETER_VERSION:="5.3"}
JMETER_FOLDER="apache-jmeter-${JMETER_VERSION}"
JMETER_FULL_PATH="${JMETER_WORK_DIR}/${JMETER_FOLDER}"
JMETER_INSTALL_URL="https://archive.apache.org/dist/jmeter/binaries/$JMETER_FOLDER.tgz"

mkdir -p $JMETER_WORK_DIR
mkdir -p $JMETER_WORK_DIR/result
mkdir -p $JMETER_WORK_DIR/test_plan

cd $JMETER_WORK_DIR

# Installation
apk upgrade
apk add --update --no-cache openjdk11

unzip_jmeter() {
  tar xzvf "$JMETER_FULL_PATH.tgz" && rm "$JMETER_FULL_PATH.tgz"
  # delete unnecessary files
  rm -rf "$JMETER_FULL_PATH/docs" "$JMETER_FULL_PATH/printable_docs"
}

# install jmeter
if [ -d "$JMETER_FULL_PATH" ]; then
  # jmeter folder already has
  echo "[CB-ANT] Jmeter is already installed."
elif [ -f "$JMETER_FULL_PATH.tgz" ]; then
  # only jmeter tgz file has
  echo "[CB-ANT] Jmeter gzip file is installed on $JMETER_WORK_DIR. Let's do remaining installation."
  unzip_jmeter
else
  # jmeter is not installed
  echo "[CB-ANT] JMeter is installing on path $JMETER_WORK_DIR"
  wget $JMETER_INSTALL_URL
  unzip_jmeter
fi

sh "$JMETER_FULL_PATH/bin/jmeter.sh" --version
echo "[CB-ANT] Jmeter is completely installed!!"
