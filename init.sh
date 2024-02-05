#!/bin/bash
export JMETER_WORKING_DIR_PREFIX="./pkg/jmeter"
export JMETER_VERSION="apache-jmeter-5.6.3"
export JMETER_INSTALL_URL="https://downloads.apache.org//jmeter/binaries/$JMETER_VERSION.tgz"


if [ -d "$JMETER_WORKING_DIR_PREFIX/$JMETER_VERSION" ]; then
    echo "[CB-ANT] Jmeter is already completely installed!!"
        
elif [ -f "$JMETER_WORKING_DIR_PREFIX/$JMETER_VERSION.tgz" ]; then 
    echo "[CB-ANT] JMeter gzip file is installed on $JMETER_WORKING_DIR_PREFIX. Let's do remaining installation."

    tar xzvf "$JMETER_WORKING_DIR_PREFIX/$JMETER_VERSION.tgz" -C "$JMETER_WORKING_DIR_PREFIX/" && rm "$JMETER_WORKING_DIR_PREFIX/$JMETER_VERSION.tgz"
    $JMETER_WORKING_DIR_PREFIX/$JMETER_VERSION/bin/jmeter --version
    echo "[CB-ANT] Jmeter is completely installed!!"    

else
    echo "[CB-ANT] JMeter is installing on path $JMETER_WORKING_DIR_PREFIX"

    wget $JMETER_INSTALL_URL -P "$JMETER_WORKING_DIR_PREFIX/"
    tar xzvf "$JMETER_WORKING_DIR_PREFIX/$JMETER_VERSION.tgz" -C "$JMETER_WORKING_DIR_PREFIX/" && rm "$JMETER_WORKING_DIR_PREFIX/$JMETER_VERSION.tgz"
    $JMETER_WORKING_DIR_PREFIX/$JMETER_VERSION/bin/jmeter --version
    echo "[CB-ANT] Jmeter is completely installed!!"
fi

