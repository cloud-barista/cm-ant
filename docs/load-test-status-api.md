# Load test — reading a run's status and steps

A load test runs asynchronously: the request to start one returns immediately with a
`loadTestKey`, and the run then goes through pre-check, generator install, the load itself and
result collection on its own. This page describes how a client (the web console, a script)
reads where a run is, so a screen can show progress and explain a failure without guessing.

## The status endpoint

`GET /api/v1/load/tests/state/last` (operationId `GetLastLoadTestExecutionState`) returns the
last run for a node. The node is identified by `nsId` / `mciId` / `vmId` in the query.

The result is one `LoadTestExecutionState`:

| Field | Meaning |
|---|---|
| `executionStatus` | Where the run is, overall — see below |
| `startAt` / `finishAt` | When the run started, and finished (finish is null while running) |
| `expectedFinishAt`, `totalExpectedExecutionSecond` | The expected end, derived from duration + ramp-up. A hint for a progress bar, not a promise — the checks and installs before the load are not counted |
| `failureMessage` | A one-line reason when the run failed |
| `nodeUid` | The *uid* of the target node. Node ids are names and get reused, so a caller that keeps showing "the last run for this node" needs the uid to notice the answer belongs to a VM that has since been replaced |
| `steps` | The per-step progress, as a tree — see [Steps](#steps) |

### executionStatus

| Value | Meaning |
|---|---|
| `on_processing` | Running — checks, installs, or the load itself |
| `on_fetching` | The load has finished and results are being collected |
| `successed` | Finished; results are available |
| `test_failed` | Stopped before finishing; `failureMessage` says why |

## Steps

`steps` is the progress of the run as a **tree**: the top level is the *phases* a run goes
through, and each phase carries its *sub-steps* in `children`. A caller that only wants the
phases can ignore `children` and read the six phase rows.

Each node — a phase or a sub-step — is one `LoadTestExecutionStepResult`:

| Field | Meaning |
|---|---|
| `seq` | Order within the run |
| `name` | The step identifier. A phase is a bare name (`precheck`); a sub-step is `phase.sub` (`precheck.target_reachable`), the part before the dot naming its phase |
| `status` | `pending` · `running` · `ok` · `failed` · `skipped` |
| `attempt` | Retry count (0 = first try) |
| `startAt` / `finishAt` | When this step started and finished |
| `message` | A short, current-status line (e.g. `Checking the metric agent port`) |
| `detail` | The verbose diagnosis or error cause, shown when a step fails (may be multi-line) |
| `elapsedSec` | How long the step has taken: the whole span once done, or the time so far while running. A phase's figure is rolled up from its children |
| `children` | The sub-steps of a phase |

`message` moves as the work advances and is the line to show as "what is happening now";
`detail` is the paragraph to show when something has gone wrong. The phase itself carries a
static message (`precheck` → "Checking the environment"); the line that actually changes is on
the running sub-step, so a live status display should read the deepest running node.

### The phases and their sub-steps

All six phases are seeded when the run starts, so the whole outline is visible from the first
read. Sub-steps appear under the phases that report them; the other phases record at the phase
level only.

| Phase (`name`) | Sub-steps (`children[].name`) | Notes |
|---|---|---|
| `precheck` | `target_exists`, `target_running`, `target_reachable`, `metric_port_open`, `remote_command` | `metric_port_open` only when *Collect Additional System Metrics* is on, else `skipped` |
| `generator_install` | `reachable` | Sub-step only when reusing a remote generator; a fresh install records at the phase level |
| `agent_install` | `install`, `process_up`, `port_reachable` | The whole phase runs only when metrics are on, else `skipped` |
| `jmx_prepare` | — | Phase level only |
| `jmeter_run` | — | Phase level only. This is the timed load; use `elapsedSec` against `totalExpectedExecutionSecond` for its progress |
| `result_fetch` | `file_result`, `file_cpu`, `file_memory`, `file_disk`, `file_network` | The metric files only when metrics are on; `file_result` always. See [Result collection](#result-collection) |

A phase with no sub-steps shows its own status only. A run that has not reached a phase leaves
it `pending`.

## Result collection

Result collection used to be one opaque `result_fetch` step, so a run whose results took far
longer than the 30-second load gave no clue which file it was waiting on. Each result file is
now its own sub-step: it starts `running` ("Waiting for the … file"), turns `ok` when it lands
("… collected"), and turns `skipped` if it never arrives within the wait window. The load
result failing fails the step; a missing metric file only costs its chart, so it is skipped
rather than failed. Because the number of files is fixed, a client can show collection as a
percentage of files received.

The files are pulled with rsync while they are still being written on the generator, so the
metric files routinely land after the main result — hence the waiting and the retries. What a
long wait on a file usually means is that the file has not been produced yet, not that the
transfer is slow.

## Reading the last configuration (Re-run)

`GET /api/v1/load/tests/infos/{loadTestKey}` (operationId `GetLoadTestExecutionInfo`) returns
the parameters a run was started with — scenario name, virtual users, duration, ramp-up, and
the HTTP request. A client uses it to pre-fill the form for a re-run.

The configuration is recorded **before** the pre-check runs, so a run that fails in the
pre-check still has its parameters on record and can be re-run from any client. (Earlier the
record was written only after pre-check and generator install had both succeeded, so a
pre-check failure left nothing to read and this call answered 500.) The generator is linked to
the record after it is installed; until then the record simply has no generator.
