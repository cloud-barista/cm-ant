#!/bin/bash
export JMETER_WORKING_DIR_PREFIX="./third_party/jmeter"
export JMETER_VERSION="apache-jmeter-5.5"
export JMETER_INSTALL_URL="https://downloads.apache.org//jmeter/binaries/$JMETER_VERSION.tgz"
export JMETER_FULL_PATH="$JMETER_WORKING_DIR_PREFIX/$JMETER_VERSION"
export JMETER_BIN="$JMETER_FULL_PATH/bin"


if [ -d "$JMETER_FULL_PATH" ]; then
    echo "[CB-ANT] Jmeter is already completely installed!!"
        
elif [ -f "$JMETER_FULL_PATH.tgz" ]; then 
    echo "[CB-ANT] JMeter gzip file is installed on $JMETER_WORKING_DIR_PREFIX. Let's do remaining installation."

    tar xzvf "$JMETER_FULL_PATH.tgz" -C "$JMETER_WORKING_DIR_PREFIX/" && rm "$JMETER_WORKING_DIR_PREFIX/$JMETER_VERSION.tgz"
    $JMETER_FULL_PATH/bin/jmeter --version
    echo "[CB-ANT] Jmeter is completely installed!!"    

else
    echo "[CB-ANT] JMeter is installing on path $JMETER_WORKING_DIR_PREFIX"

    wget $JMETER_INSTALL_URL -P "$JMETER_WORKING_DIR_PREFIX/"
    tar xzvf "$JMETER_FULL_PATH.tgz" -C "$JMETER_WORKING_DIR_PREFIX/" && rm "$JMETER_FULL_PATH.tgz"
    $JMETER_FULL_PATH/bin/jmeter --version
    echo "[CB-ANT] Jmeter is completely installed!!"
fi

export PATH="$PATH:$JMETER_BIN"
alias jmeter="$JMETER_BIN/jmeter"
