# Load test — prerequisites and what the messages mean

A load test reaches across three machines: cm-ant orchestrates, a generator VM produces the
load, and the target VM receives it while reporting its own cpu and memory. Most failures are
about one of them not being reachable from another, so this page lists what has to be open
before a run, and what each message means when something is not.

## Before running

### Ports on the target

| Port | Needed for | If closed |
|---|---|---|
| The port under test (e.g. 80) | The load itself | The run stops during the precheck |
| **5555** | System metrics — cpu, memory, disk, network | The run stops during the precheck. The agent only needs the port open; the run installs it |
| 22 | Installing and starting the metric agent | The run stops during the precheck |

Both the load port and 5555 must allow inbound traffic in the target's security group.

**5555 is not optional when metrics are requested.** Every chart on the results screen other
than response time is drawn from what comes through it, so a run that cannot reach it stops
rather than spending several minutes to arrive at response times alone.

What the precheck settles about 5555 is whether the port is *open*, not whether the agent is
already running on it. The agent is installed by the run itself, so a port that answers with a
refusal — open, nothing listening yet — is the normal state of a target that has not been used
for a load test before, and the run carries on to install. A port that answers nothing at all
is the security group, and no later step can undo that, so the run stops there.

If system performance figures are not needed, clear **Collect Additional System Metrics** in
the load configuration. The run then skips the metric agent entirely and 5555 is not required.

### The target itself

- The node exists and is `Running`
- The service under test answers on the configured path — the precheck sends the same request
  the load will send, so a path that returns 404 is reported before the run starts
- Remote commands work, since everything cm-ant does to the target and the generator goes
  through cb-tumblebug's remote command

## The precheck

A run begins by checking the above. These answers arrive in seconds, so a wrong environment is
reported immediately instead of surfacing minutes into a run that was never going to work.

Each check is retried up to three times before it is called a failure. A healthy target answers
on the first attempt, so the retries cost nothing when things are fine, and a lost packet does
not get reported as a firewall problem. When a check needs more than one attempt the step says
so — `(attempt 2 of 3)` is worth noticing even when the check eventually passes.

## Messages

Each step shows a short line, with the explanation behind it on hover.

| Message | What happened | What to do |
|---|---|---|
| `Target node not found` | cb-tumblebug has no such node | Check the namespace, infra and node ids |
| `Target node is Suspended, not running` | The node exists but is not running | Start the node |
| `Target port 80 unreachable` | Nothing answered the configured request | Check, in order: the service is listening on the target, the security group allows inbound on that port, the node has a reachable address |
| `Target answered with status 404` | The target is alive but that path is not serving | Check the request path. The run continues — an endpoint that returns an error may be what you meant to test |
| `Metric port 5555 unreachable` | Nothing answered on 5555 at all — the traffic is being dropped | Add 5555 inbound to the security group and run again, or clear **Collect Additional System Metrics** to run without them. Installing the agent cannot help while the port is closed |
| `Metric agent could not be started` | Two install rounds ended without an agent process | The detail lists what each round saw. Check java on the target, that `/opt/perfmon-agent` holds the unpacked agent, and that nothing else is bound to 5555 |
| `Remote command failed on the target` | cb-tumblebug could not run a command on the node | If the node started recently, wait a few minutes and try again (see below). Otherwise check port 22 and the address cb-tumblebug holds for the node |
| `Metric agent not running` | The agent was installed but no process is up | Check java on the target and `/opt/perfmon-agent`. The install script starts the agent with nohup and reports success either way, so this means it did not stay up |
| `Metric agent silent on port 5555` | The process is up but nothing answers | The process being up rules out a failed start, so this is almost always the security group |
| `Results collected, without network metrics` | The load figures are in; one metric file never arrived | The charts for that metric will be empty. The run is otherwise complete |

## Installing the metric agent

The agent install fetches a JRE through apt and then unpacks the agent release, so on a target
that has never run a load test it takes a while — usually well under a minute, longer on a
small instance or a slow mirror. The step does not simply wait out a fixed time. While no agent
process is there, it asks the target whether an install is still running, and:

- **an install is running** — it keeps waiting, up to two minutes, reporting the elapsed time
- **nothing is running** — there is nothing to wait for, so it reinstalls immediately rather
  than spending the ceiling on a wait that was never going to end

An install that reaches the ceiling, or ends without leaving an agent behind, gets one more
round. Two rounds that both end without an agent are reported as `Metric agent could not be
started`, and the detail says what each round saw — an install still running after two minutes
is a different problem from an install that finished and left nothing.

## When a node has just started

Remote command failures right after a node starts are usually a matter of timing rather than
configuration. On AWS, ssh can take several minutes to accept connections after a boot, and
small instance types take longer. The check waits ten seconds per attempt rather than sitting
through it, so the answer comes back quickly and the run can be retried once the node has
settled.

Before running again, confirm you can reach port 22 yourself. A plain `ssh` from outside the
platform settles in one step what the message can only guess at — whether the security group
allows it, whether anything else in front of the node is filtering, and whether the ssh service
is up at all. A run started while that is still failing will fail the same way.

A stop and start also gives the node a new public address. cb-tumblebug picks the new one up
when the node is stopped and started through it, but the two can drift apart — if a run fails
to reach a node that looks healthy in the provider console, compare the address cb-tumblebug
holds against the one the provider shows.

## While a run is going

Steps are grouped by phase, and every step carries how long it has taken — the whole span once
finished, the time so far while running. A step that normally takes a second and now reads
`45s` is the one to look at.

Metric files are written during the run and routinely arrive after the main result, so
collection keeps going while files are still arriving and stops once a round brings nothing
new. Files that never came are named rather than silently missing.

A missing metric file does not fail the run. Only the main result does — the load figures are
what the run is for.
