#!/bin/bash

set -e

JMETER_WORK_DIR=${JMETER_WORK_DIR:="${HOME}/jmeter"}
TEST_PLAN_PATH=${JMETER_WORK_DIR}/test_plan
RESULT_PATH=${JMETER_WORK_DIR}/result
JMETER_VERSION=${JMETER_VERSION:="5.3"}
JMETER_EXECUTION_PATH=${JMETER_WORK_DIR}/apache-jmeter-${JMETER_VERSION}/bin/jmeter.sh

echo "[Step 1] Test plan path folder check"
if [ ! -e "${TEST_PLAN_PATH}" ]; then
  mkdir -p "${TEST_PLAN_PATH}"
  echo "test plan path folder created"
fi

echo "[Step 2] Result path folder check"
if [ ! -e "${RESULT_PATH}" ]; then
  mkdir -p "${RESULT_PATH}"
  echo "result path folder created"
fi

TEST_PLAN_NAME=${TEST_PLAN_NAME}

echo "[Step 3] Test plan '${TEST_PLAN_NAME}' check in ${TEST_PLAN_PATH}"
if [ ! -e "${TEST_PLAN_PATH}/${TEST_PLAN_NAME}" ]; then
  # wget "https://github.com/MZC-CSC/cm-ant/raw/feature-performance-test-sehyeong/test_plan/${TEST_PLAN_NAME}" -P "${TEST_PLAN_PATH}"
  # echo "test plan downloaded to ${TEST_PLAN_PATH}/${TEST_PLAN_NAME}"
  echo "@@@@@@@@@@@@@@ test plan file does not exist"
  exit 1
fi
