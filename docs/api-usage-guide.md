# Simple & Sample API Usage Guide

> This document has been migrated from the project GitHub wiki to the repository `docs/`. From now on, documentation is maintained only in the repository `docs/`, not in the wiki.

## Cost Estimation API Call Flow

1. Query recommended models
2. Update or query cost estimation data
3. Migration process
4. Update estimated usage cost data
5. Query estimated usage cost data

---

## Cost Estimation API Specification

> The request/response examples for the cost estimation API below are values captured from actual calls against the dev cm-ant live environment (as of 2026-07-21). The response structure changed with a cb-Spider update, and those changes are reflected here.

### 1. Update or Query Cost Estimation Data API

#### Endpoint

`POST /api/v1/cost/estimate`

#### Description

Provides cost estimation data for the given providerName, regionName, instanceType, and image (optional) information. If the spec does not exist in Ant's database, or if it exists but the data is older than 7 days, the cost information is fetched from a sub-system, stored, and then returned in the response.

#### Request Body

```json
{
  "specs": [
    {
      "providerName": "aws",
      "regionName": "ap-northeast-2",
      "instanceType": "t3.small",
      "image": "ubuntu22.04"
    }
  ],
  "specsWithFormat": [
    {
      "commonSpec": "aws+ap-northeast-2+t3.small",
      "commonImage": "aws+ap-northeast-2+ubuntu22.04"
    }
  ]
}
```

- Only one of `specsWithFormat` or `specs` needs to be provided. (At least one is required.)
- The specs from all requests are aggregated and returned together in the result.
- The provided price may differ depending on the OS and license.

`specs`

- `providerName (required):` requested provider name (aws | azure | alibaba | tencent | gcp | ibm)

- `regionName (required)`: region name matching the CSP

- `instanceType (required)`: instance type matching the CSP

- `image(optional)`: OS image information used by the CSP.

`specsWithFormat`

- `commonSpec (required):` `providerName+region+instanceType` information joined by the `+` separator.

- `commonImage (Optional)`: `providerName+region+machineImage` information joined by the `+` separator.

#### Response

```json
{
  "successMessage": "Successfully update and get estimate cost info",
  "code": 200,
  "result": {
    "esimateCostSpecResults": [
      {
        "providerName": "aws",
        "regionName": "ap-northeast-2",
        "instanceType": "t3.small",
        "totalMinMonthlyPrice": 18.72,
        "totalMaxMonthlyPrice": 18.72,
        "estimateForecastCostSpecDetailResults": [
          {
            "id": 1,
            "vCpu": "2",
            "memory": "2048 GiB",
            "productDescription": "productFamily= Compute Instance, version= 20260721012550",
            "originalPricePolicy": "OnDemand",
            "pricePolicy": "OnDemand",
            "unit": "PerHour",
            "currency": "USD",
            "price": "0.0260000000",
            "calculatedMonthlyPrice": 18.72,
            "priceDescription": "$0.026 per On Demand Linux t3.small Instance Hour",
            "lastUpdatedAt": "2026-07-21T11:48:09.319204Z"
          }
        ]
      }
    ]
  }
}
```

#### Query Cost Estimation Data (Read-Only)

Cost estimation data that has already been stored can be checked via a separate query API. Since it queries only Ant's database, no sub-system call cost is incurred.

`GET /api/v1/cost/estimate`

```text
Query Param:
providerName(required), regionName(required), instanceType, vCpu, memory, osType, page, size
```

- `providerName (required)`: provider name to query
- `regionName (required)`: region name to query
- `instanceType`, `vCpu`, `memory`, `osType`: filter conditions (optional)
- `page`, `size`: pagination (optional)

Example: `GET /api/v1/cost/estimate?providerName=aws&regionName=ap-northeast-2&instanceType=t3.small`

```json
{
  "successMessage": "Successfully get price info.",
  "code": 200,
  "result": {
    "estimateCostInfoResult": [
      {
        "id": 1,
        "providerName": "aws",
        "regionName": "ap-northeast-2",
        "instanceType": "t3.small",
        "vCpu": "2",
        "memory": "2048 GiB",
        "productDescription": "productFamily= Compute Instance, version= 20260721012550",
        "originalPricePolicy": "OnDemand",
        "pricePolicy": "OnDemand",
        "unit": "PerHour",
        "currency": "USD",
        "price": "0.0260000000",
        "calculatedMonthlyPrice": 18.72,
        "priceDescription": "$0.026 per On Demand Linux t3.small Instance Hour",
        "lastUpdatedAt": "2026-07-21T11:48:09.319204Z",
        "imageName": ""
      }
    ],
    "resultCount": 1
  }
}
```

---

### 2. Update Estimated Usage Cost Data API

#### Endpoint

`POST /api/v1/cost/estimate/forecast`

#### 

#### Description

Fetches operational cost data for migrated workloads into ant and stores it.
Only up to **14 days of cost data from the present can be stored, and each call incurs a Cost Explorer API usage charge of $0.01 billed to the connected aws account**.

It retrieves information for the resources (currently only instance ids can be obtained) corresponding to the nsId and infraId used in Tumblebug, and updates the matching price information.

The currently supported provider is aws, with more to be added in the future.

#### Request Parameters

```json
{
  "nsId": "mig01",
  "infraId": "infra101"
}
```

- `nsId`: Tumblebug namespace ID
- `infraId`: Tumblebug infra (formerly mci) ID

#### Response

```json
{
  "successMessage": "Successfully update estimate forecast cost info.",
  "code": 200,
  "result": {
    "fetchedDataCount": 0,
    "updatedDataCount": 0,
    "insertedDataCount": 0
  }
}
```

- `fetchedDataCount`: number of cost data records fetched from the sub-system
- `updatedDataCount`: number of existing records that were updated
- `insertedDataCount`: number of newly stored records

---

### 3. Query Estimated Usage Cost Data API

#### Endpoint

`GET /api/v1/cost/estimate/forecast`

#### Description

Queries the estimated usage cost data matching the user's filter conditions from ant's database. You must first run the estimated usage cost data update before querying.
Because it queries from ant's database, no separate API call cost is incurred.

```text
Query Param: startDate(format 2026-07-14), endDate(format 2026-07-21), costAggregationType, provider, nsIds, infraIds, resourceTypes, resourceIds, dateOrder, resourceTypeOrder
```

#### 

#### Request Parameters

- startDate: start date (inclusive) in the format 2026-07-14
- endDate: end date (inclusive) in the format 2026-07-21
- costAggregationType: aggregation interval for grouping cost data (*daily|weekly|monthly)
- provider: value to filter by provider (currently only aws is supported)
- nsIds: list of nsIds to query
- infraIds: list of infraIds to query
- resourceType: resource type to query (VM|VNet|DataDisk|Etc|)
- resourceIds: list of resource ids to query
- dateOrder: ascending or descending by date (asc|desc)
- resourceTypeOrder: ascending or descending by resourceType (asc|desc)

Example: `GET /api/v1/cost/estimate/forecast?startDate=2026-07-14&endDate=2026-07-21&costAggregationType=daily&nsIds=mig01&infraIds=infra101`

#### 

#### Response

The following is an actual response that is empty because no cost data has yet been billed for that infra. Once cost data accumulates through the estimated usage cost data update, the `getEstimateForecastCostInfoResults` list will be populated and `resultCount` will increase accordingly.

```json
{
  "successMessage": "Successfully get estimate forecast cost",
  "code": 200,
  "result": {
    "getEstimateForecastCostInfoResults": [],
    "resultCount": 0
  }
}
```

When the list contains data, each item has the fields `category` · `date` · `provider` · `resourceId` · `resourceType` · `totalCost` · `unit`.

---

### 5. Update Usage Cost Estimation RAW Data API

#### Description

Fetches operational cost data for migrated workloads into ant and stores it.
Only up to 14 days of cost data from the present can be stored, and each call incurs a Cost Explorer API usage charge of $0.01 billed to the connected aws account.

This allows storing cost data into ant using the ids of the desired resources, regardless of ns and infra.

```json
{
  "costResources": [
    {
      "resourceType": "VM",
      "resourceIds": [
        "i-02dxxxxxxdec4"
      ]
    },
    {
      "resourceType": "VNet",
      "resourceIds": [
        "eni-06cxxxxxx02e34"
      ]
    },
    {
      "resourceType": "DataDisk",
      "resourceIds": [
        "vol-0917xxxxxxe2f6"
      ]
    }
  ],
  "awsAdditionalInfo": {
    "ownerId": "xxxxxxxxxxxx",
    "regions": [
      "ap-northeast-2"
    ]
  }
}
```

#### Request Parameters

- costResources: the information required to query cost data
  - resourceType: defines which resource's cost data to query (VM|VNet|DataDisk)
  - resourceIds: list of resource ids for that type
- awsAdditionalInfo: additional data required for aws cost queries
  - ownerId: aws value consisting of 12 digits
  - regions: region name required when combining eni (used for VNet cost queries)

---

## Performance Evaluation API Call Flow

1. Select the migrated workload
   - nsId, mciId, vmId
2. Request metrics agent installation
3. Enter the user-defined scenario for the performance evaluation, then run the evaluation
4. Query the status of the performance evaluation run against the workload
   - So that the status can be displayed on the workload
5. Query the results of the performance evaluation run against the workload

### Notes

Currently there is no relationship defined in the framework between a migrated workload and a performance evaluation run. <br>
In terms of the performance evaluation flow, a link between the workload and the performance evaluation appears necessary, and as mentioned previously we plan to add this linkage.

---

## Performance Evaluation API Specification

### 1. Install the Performance Metrics Collection Agent (optional)

#### Endpoint

`POST /api/v1/load/monitoring/agent/install`

#### Description

This is the step to install the metric agent used to additionally collect metric information during a performance evaluation. Through the metric agent, infrastructure metrics such as CPU, memory, network I/O, and disk I/O are collected when load is generated.

If you run the performance evaluation without installing it separately, only metrics for HTTP requests such as latency and requests per second are collected.

When you run a performance evaluation, you can choose the option to collect metrics with the agent.

```json
{
  "nsId": "nsId",
  "mcisId": "mcisId",
  "vmId": "vmId"  # optional 
}
```

#### Request Properties

- nsId: namespace ID
- mciId: MCI ID
- vmId: VM ID

-----

### 2. Install the Load Generator for Performance Evaluation (optional)

#### Endpoint

`POST /api/v1/load/generators`

#### Description

This is the step to install the Load Generator that generates load for running a performance evaluation.
It can be installed in a Local or Remote environment.

If you run the performance evaluation without installing it separately, the `installLocation` property of the load generation run request is checked to proceed with the load generator installation.

```json
{
  "installLocation": "local"  // local | remote
}
```

#### Properties

- installLocation: installation location (local or remote)
  
  - local: installed in the environment where ANT currently runs (Ubuntu-based)
  
  - remote: provisions a VM via Tumblebug's recommendVM and then installs on the provisioned VM
    
    | t2.medium | 2   | 4.0 |
    | --------- | --- | --- |

-----

### 3. Start Load Generation for Performance Evaluation

#### Endpoint

`POST /api/v1/load/tests/run`

#### Description

This is the step to run the load test for a performance evaluation.

This API includes the preceding steps of installing the performance metrics collection agent and installing the performance evaluation load generator.

```json
{
        "agentHostname": "",
        "collectAdditionalSystemMetrics": true,
        "httpReqs": [
            {
                "bodyData": "",
                "hostname": "xx.xxx.xx.xxx",
                "method": "get",
                "path": "",
                "port": "80",
                "protocol": "http"
            }
        ],
        "installLoadGenerator": {
            "installLocation": "remote"
        },
        "nsId": "mig01",
        "mciId": "mmci01",
        "vmId": "vm01-1",
        "testName": "first test gogo",
        "virtualUsers": "10",
        "duration": "20",
        "rampUpTime": "20",
        "rampUpSteps": "3"
    }
```

The load generation run request returns immediately, and the response contains a `loadTestKey`. The actual load test (precheck → load generator installation → load generation → result collection) runs asynchronously, so the progress status is queried via **4. Check Load Generation Status** below.

#### Properties

- installLoadGenerator: Load Generator installation location (skipped if already installed earlier)
  - installLocation: installation location (remote or local)
- testName: test name
- virtualUsers: number of virtual users
- duration: test run duration (in seconds)
- rampUpTime: load ramp-up time (in seconds)
- rampUpSteps: number of load ramp-up steps
- hostname: target host name for the test
- port: port number
- agentHostname: host name where the agent is installed (Optional)
- collectAdditionalSystemMetrics: whether to collect metrics via the agent
- httpReqs: list of HTTP requests
  - method: HTTP method (GET, POST)
  - protocol: communication protocol (http, https)
  - hostname: host name to request
  - port: port number
  - path: request path
  - bodyData: request body data (Optional)

-----

### 4. Check Load Generation Status

#### Endpoint

The load generation status can be queried in two ways.

- `GET /api/v1/load/tests/state/{loadTestKey}` — query the run status for a specific test key
- `GET /api/v1/load/tests/state/last` (operationId `GetLastLoadTestExecutionState`) — query the **most recently performed** run status for a specific workload (node). The node is specified via the query parameters `nsId` / `mciId` / `vmId`.

> Since a node id is a name and can be reused, a screen that keeps showing "the last run of this workload" must use the response's `nodeUid` to distinguish whether that run belongs to a VM that was subsequently replaced.

#### Description

Checks the load generation status. Returns the state and information of the ongoing load generation.

```text
Path Param
loadTestKey: the test key returned when the performance evaluation was run
```

The response includes fields that express the run's progress in detail.

- `executionStatus`: overall run status. Values are `on_processing` (in progress: precheck, installation, load generation, etc.) · `on_fetching` (load is done and results are being collected) · `successed` (completed, results can be queried) · `test_failed` (aborted before completion; the cause is in `failureMessage`)
- `startAt` / `finishAt`: run start/end time (if in progress, `finishAt` is null)
- `expectedFinishAt`, `totalExpectedExecutionSecond`: the expected finish computed from duration + ramp-up. Since it does not include the precheck/installation before load generation, use it only as a hint for the progress bar.
- `failureMessage`: a one-line reason on failure (empty on success)
- `nodeUid`: the uid of the target node. Since a node id is reused, it is used to identify which VM the run belongs to.
- `steps`: expresses the step-by-step progress of the run as a **tree**. The top level is the *phase* the run goes through, and each phase's sub-steps are contained in `children`. If you only need the phases, you can ignore `children` and read only the phase rows.

Fields of each node (phase or sub-step) in `steps[]`:

- `seq`: order within the run
- `name`: step identifier. A phase has a single name (`precheck`); a sub-step has the form `phase.sub` (`precheck.target_reachable`)
- `status`: `pending` · `running` · `ok` · `failed` · `skipped`
- `message`: a short one-line description of the current status (used to show what is being done right now)
- `detail`: detailed diagnostics/error cause to show on failure (may be multiple lines)
- `elapsedSec`: time taken by that step. For a completed step it is the entire duration; if in progress, the time so far. A phase's value is summed from its children.
- `children`: list of sub-steps of the phase

#### Response Example

```json
{
        "code": 200,
        "result": {
            "compileDuration": "436.008µs",
            "createdAt": "2024-11-05T06:22:08.832848Z",
            "executionDuration": "1m4.032562237s",
            "executionStatus": "on_processing",
            "id": 1,
            "loadGeneratorInstallInfo": {
                "createdAt": "2024-11-05T06:19:27.85666Z",
                "id": 1,
                "installLocation": "local",
                "installPath": "/opt/ant/jmeter",
                "installType": "jmeter",
                "installVersion": "5.6",
                "status": "installed",
                "updatedAt": "2024-11-05T06:22:08.824086Z"
            },
            "loadTestKey": "1730787567844-loadtest-key-example",
            "nodeUid": "vm-9f2c1a7b-3e40-4d21-9c88-0a1b2c3d4e5f",
            "startAt": "2024-11-05T06:22:08.828219Z",
            "finishAt": null,
            "expectedFinishAt": "2024-11-05T06:23:08.828219Z",
            "totalExpectedExecutionSecond": 60,
            "failureMessage": "",
            "updatedAt": "2024-11-05T06:23:12.869756Z",
            "steps": [
                {
                    "seq": 1,
                    "name": "precheck",
                    "status": "ok",
                    "message": "Checking the environment",
                    "elapsedSec": 5,
                    "children": [
                        {
                            "seq": 1,
                            "name": "precheck.target_exists",
                            "status": "ok",
                            "elapsedSec": 1
                        },
                        {
                            "seq": 3,
                            "name": "precheck.target_reachable",
                            "status": "ok",
                            "message": "Target port 80 reachable",
                            "elapsedSec": 3
                        },
                        {
                            "seq": 4,
                            "name": "precheck.metric_port_open",
                            "status": "skipped",
                            "message": "Metrics not requested"
                        }
                    ]
                },
                {
                    "seq": 2,
                    "name": "generator_install",
                    "status": "ok",
                    "elapsedSec": 12
                },
                {
                    "seq": 3,
                    "name": "agent_install",
                    "status": "skipped",
                    "message": "Metrics not requested"
                },
                {
                    "seq": 4,
                    "name": "jmx_prepare",
                    "status": "ok",
                    "elapsedSec": 1
                },
                {
                    "seq": 5,
                    "name": "jmeter_run",
                    "status": "running",
                    "message": "Generating load",
                    "elapsedSec": 22
                },
                {
                    "seq": 6,
                    "name": "result_fetch",
                    "status": "pending",
                    "children": [
                        {
                            "seq": 1,
                            "name": "result_fetch.file_result",
                            "status": "pending"
                        },
                        {
                            "seq": 2,
                            "name": "result_fetch.file_cpu",
                            "status": "pending"
                        }
                    ]
                }
            ]
        },
        "successMessage": "Successfully retrieved load test execution state information"
    }
```

> For details such as the full structure of the `steps` tree (phase ↔ sub-step), the sub-steps of each phase, and how `result_fetch` collects result files, see [docs/load-test-status-api.md](load-test-status-api.md). For the meaning of each step's message on failure, see [docs/load-test-troubleshooting.md](load-test-troubleshooting.md).

------

### 5. Check Performance Evaluation Results

#### Endpoint

`GET /api/v1/load/tests/result`

`GET /api/v1/load/tests/result/metrics`

#### Description

This is the step to query the performance evaluation results after load generation is complete.

```text
Query Param

loadTestKey: the test key returned when the performance evaluation was run
format: result format (normal | aggregate)
```

If a monitoring agent is installed, the cpu, memory, disk, network, and other data collected via the metric agent can be checked by calling the metrics result query API.

```text
Query Param:
loadTestKey: the test key returned when the performance evaluation was run
```

#### Performance Evaluation Result Query Response

For `aggregate`, the values are aggregated over the results. It provides aggregated values over all data without any data sampling.

For `normal`, the response values are expressed as time-series data. The result data is provided by computing the average per 100ms interval, finding the closest value, and extracting the sampled result value.

```json
 "[aggregate]": {
    "code": 0,
    "errorMessage": "string",
    "result": [
      {
        "average": 0,  # average latency (ms)
        "errorPercent": 0,  # error rate
        "label": "string", # test name
        "maxTime": 0,  # maximum latency (ms)
        "median": 0,  # median latency (ms)
        "minTime": 0,   # minimum latency (ms)
        "ninetyFive": 0,  # 95th percentile latency (ms)
        "ninetyNine": 0,  # 99th percentile latency (ms)
        "ninetyPercent": 0,  # 90th percentile latency (ms)
        "receivedKB": 0,  # received KB / sec
        "requestCount": 0,  # total number of requests generated
        "sentKB": 0,  # sent KB / sec
        "throughput": 0  # throughput for requests / sec
      }
    ],
    "successMessage": "string"
  },
  "[normal]": {
    "code": 0,
    "errorMessage": "string",
    "result": [
      {
        "label": "string",
        "results": [  # list of individual result values for individual requests
          {
            "bytes": 0,
            "connection": 0,
            "elapsed": 0,  # total elapsed time of an individual request (ms)
            "idleTime": 0,
            "isError": true,  # whether it is an error
            "latency": 0,  # latency of an individual request (ms)
            "no": 0,
            "sentBytes": 0,
            "timestamp": "string",  # time an individual request occurred
            "url": "string"
          }
        ]
      }
    ],
    "successMessage": "string"
  }
}
```

#### Performance Evaluation Metrics Query Response

```json
{
  "code": 0,
  "errorMessage": "string",
  "result": [
    {
      "label": "string",  # can be checked in the list of collected metrics below
      "metrics": [  
        {
          "isError": true,  # whether the request is an error
          "timestamp": "string",  # time an individual request occurred
          "unit": "string",  # can be checked in the list of collected metrics below
          "value": "string"  # value for the collected metric
        }
      ]
    }
  ],
  "successMessage": "string"
}
```

##### List of Collected Performance Evaluation Metrics

```json
"cpu_all_combined": {  # overall cpu usage
        Unit:     "%",
    },
    "cpu_all_idle": {  # idle cpu ratio
        Unit:     "%",
    },
    "memory_all_used": {  # memory usage
        Unit:     "%",
    },
    "memory_all_free": {  # idle memory ratio
        Unit:     "%",
    },
    "memory_all_used_kb": {  # memory used kb
        Unit:     "mb",
    },
    "memory_all_free_kb": {  # idle memory kb
        Unit:     "mb",
    },
    "disk_read_kb": {  # disk read kb
        Unit:     "kb",
    },
    "disk_write_kb": {   # disk write kb
        Unit:     "kb",
    },
    "disk_use": {  # disk usage
        Unit:     "%",
    },
    "disk_total": {  # total disk capacity mb
        Unit:     "mb",
    },
    "network_recv_kb": {  # network received kb
        Unit:     "kb",
    },
    "network_sent_kb": {  # network sent kb
        Unit:     "kb",
    },
```
