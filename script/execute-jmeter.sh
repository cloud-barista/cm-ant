#!/bin/bash

set -e

JMETER_WORK_DIR=${JMETER_WORK_DIR:="/opt/jmeter"}
TEST_PLAN_PATH=${JMETER_WORK_DIR}/test_plan
RESULT_PATH=${JMETER_WORK_DIR}/result
JMETER_VERSION=${JMETER_VERSION:="5.3"}
JMETER_EXECUTION_PATH=${JMETER_WORK_DIR}/apache-jmeter-${JMETER_VERSION}/bin/jmeter

if [ ! -e ${TEST_PLAN_PATH} ]; then
  mkdir -p ${TEST_PLAN_PATH}
fi

if [ ! -e ${RESULT_PATH} ]; then
  mkdir -p ${RESULT_PATH}
fi

cd ${TEST_PLAN_PATH}

if [ -e "${TEST_PLAN_PATH}/test_plan_1.jmx" ]; then
  wget https://github.com/MZC-CSC/cm-ant/blob/feture_performa/test_plan/test_plan_1.jmx
fi

cd ${JMETER_WORK_DIR}

${JMETER_EXECUTION_PATH} -n -f \
  -Jthreads="${THREAD}" \
  -JrampTime="${RAMP_TIME}" \
  -JloopCount="${LOOP_COUNT}" \
  -Jprotocol="${PROTOCOL}" \
  -Jhostname="${HOST_NAME}" \
  -Jport="${PORT}" \
  -Jpath="${PATH}" \
  -JbodyData="${BODY_DATA}" \
  -t "${TEST_PLAN_PATH}/${TEST_PLAN_NAME}" \
  -l "${RESULT_PATH}/${RESULT_FILE_NAME}"
