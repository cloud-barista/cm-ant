package load

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/cloud-barista/cm-ant/internal/core/common/constant"
	"github.com/cloud-barista/cm-ant/internal/infra/outbound/tumblebug"
)

// Checks that can be answered in seconds, run before anything is provisioned (BAR-1553).
//
// A run used to start by building a load generator, so a missing target or a closed port only
// surfaced after minutes of work — in the worst case measured, 27 minutes went by before it
// became clear the metric agent was unreachable. Everything below is cheap, so it goes first:
// if the environment is wrong the run stops here and says what to fix.
//
// Passing this does not mean the run will succeed. It means a later failure is about how the
// run behaved, not about the environment it was given.

const (
	precheckDialTimeout = 5 * time.Second
	precheckHTTPTimeout = 10 * time.Second
	// A working remote command answers in about a second, so waiting half a minute for one
	// buys nothing and delays the very finding this check exists to deliver quickly. Ten
	// seconds is ten times the normal round trip.
	//
	// Waiting longer would not help in the usual case either: after a node restarts, ssh often
	// takes minutes to accept connections - more so on small instance types - so this is a
	// "try again shortly" situation rather than something to sit through.
	precheckRemoteTimeout = 10 * time.Second

	// Attempts are spaced further apart than the other checks. Cutting a request short on our
	// side does not stop cb-tumblebug working on it, so retrying immediately would stack
	// requests against a service that is already struggling to answer.
	remoteRetryDelay = 5 * time.Second
	metricAgentPort  = "5555"

	// A single failed attempt is not enough to call something closed. A healthy target answers
	// immediately, so retrying costs almost nothing when things are fine, and it keeps a
	// momentary network hiccup from being reported as a firewall problem. The attempt count
	// goes into the step message, so a check that needed retries reads as a warning sign even
	// when it eventually passes.
	precheckAttempts   = 3
	precheckRetryDelay = 2 * time.Second
)

// retry runs check until it passes, reporting each attempt so the console can show that
// something needed more than one try. Only the latest state is kept, which is enough: seeing
// "attempt 2" means the first one failed.
func retry(rec *stepRecorder, step constant.ExecutionStep, what string, check func() error) error {
	var err error
	for attempt := 1; attempt <= precheckAttempts; attempt++ {
		if attempt > 1 {
			rec.progress(step, attempt, fmt.Sprintf("%s (attempt %d of %d)", what, attempt, precheckAttempts),
				fmt.Sprintf("previous attempt failed: %v", err))
			time.Sleep(precheckRetryDelay)
		}
		if err = check(); err == nil {
			return nil
		}
	}
	return err
}

func (l *LoadService) runPrecheck(ctx context.Context, param RunLoadTestParam, rec *stepRecorder) error {
	rec.begin(constant.StepPrecheck, "Checking the environment")

	// ── the target exists and is running ────────────────────────────────────────────────
	rec.begin(constant.SubTargetExists, "Looking up the target node")
	vm, err := l.tumblebugClient.GetVmWithContext(ctx, param.NsId, param.InfraId, param.NodeId)
	if err != nil {
		msg := "Target node not found"
		rec.fail(constant.SubTargetExists, msg, fmt.Sprintf(
			"looked up ns=%s infra=%s node=%s in cb-tumblebug: %v",
			param.NsId, param.InfraId, param.NodeId, err))
		rec.fail(constant.StepPrecheck, msg, err.Error())
		return fmt.Errorf("target node not found: %w", err)
	}
	rec.ok(constant.SubTargetExists, "Target node found")

	rec.begin(constant.SubTargetRunning, "Checking the target state")
	if !strings.EqualFold(vm.Status, "Running") {
		msg := fmt.Sprintf("Target node is %s, not running", vm.Status)
		rec.fail(constant.SubTargetRunning, msg, fmt.Sprintf(
			"ns=%s infra=%s node=%s reported status %q; start the node before running a load test",
			param.NsId, param.InfraId, param.NodeId, vm.Status))
		rec.fail(constant.StepPrecheck, msg, "")
		return fmt.Errorf("target node is not running: %s", vm.Status)
	}
	rec.ok(constant.SubTargetRunning, "Target node is running")

	// ── the target answers the request the load will actually send ──────────────────────
	//
	// An open port only proves something is listening. It says nothing about the path the
	// load will hit, and a target that answers from here may still be unreachable from the
	// generator. So the configured request is sent once, and the status code is kept.
	rec.begin(constant.SubTargetReachable, "Sending a test request to the target")
	var code int
	err = retry(rec, constant.SubTargetReachable, "Sending a test request to the target", func() error {
		var e error
		code, e = probeTarget(ctx, param.HttpReqs)
		return e
	})
	if err != nil {
		msg := fmt.Sprintf("Target port %s unreachable", portOf(param.HttpReqs))
		rec.fail(constant.SubTargetReachable, msg, fmt.Sprintf(
			"sent %s %s from cm-ant %d times over %s, each with a %s timeout; last error: %v\n"+
				"check in this order: the service listens on the target, the security group allows inbound on that port, and the node has a reachable address",
			methodOf(param.HttpReqs), targetURL(param.HttpReqs), precheckAttempts,
			precheckRetryDelay*time.Duration(precheckAttempts-1), precheckHTTPTimeout, err))
		rec.fail(constant.StepPrecheck, msg, err.Error())
		return fmt.Errorf("target is not reachable: %w", err)
	} else if code >= 400 {
		// The target is alive but the path is not serving. Worth saying, not worth stopping
		// for — the user may be load testing an endpoint that returns an error on purpose.
		rec.progress(constant.SubTargetReachable, 0,
			fmt.Sprintf("Target answered with status %d - check the request path", code), "")
		rec.ok(constant.SubTargetReachable, fmt.Sprintf("Target answered with status %d", code))
	} else {
		rec.ok(constant.SubTargetReachable, fmt.Sprintf("Target answered with status %d", code))
	}

	// ── the metric agent port ───────────────────────────────────────────────────────────
	//
	// The charts are the point of a load test, and they are drawn from the metrics this port
	// carries. What this check settles is whether the port can carry them at all - that is,
	// whether the security group lets the traffic through. Something already listening is not
	// required, because installing the agent is the run's own next job.
	//
	// The distinction is the whole point here. A dial that times out means the packets are
	// being dropped, and no later step can undo that, so the run stops and asks for the port
	// to be opened rather than spending several minutes arriving at the same place with less
	// to show for it. A dial that is refused means the path is clear and nothing is listening
	// yet - precisely the state a target is in before the agent is installed, and the state a
	// target returns to when the agent dies. Stopping there would leave the run unable to do
	// the one thing that would fix it, so it goes on and verifyMetricAgent judges the result
	// once the install has actually had its turn.
	if param.CollectAdditionalSystemMetrics {
		rec.begin(constant.SubMetricPortOpen, "Checking the metric agent port")
		err := retry(rec, constant.SubMetricPortOpen, "Checking the metric agent port", func() error {
			return dial(vm.PublicIP, metricAgentPort)
		})
		switch {
		case err == nil:
			rec.ok(constant.SubMetricPortOpen, "Metric agent port is reachable")
		case isConnectionRefused(err):
			rec.ok(constant.SubMetricPortOpen, "Metric agent port is open, agent not up yet")
		default:
			msg := fmt.Sprintf("Metric port %s unreachable", metricAgentPort)
			detail := fmt.Sprintf(
				"Could not reach %s:%s, and nothing answered at all - the traffic is being dropped rather than refused.\n"+
					"Add %s inbound to the security group, then run the test again.\n"+
					"System metrics cannot be measured until this is resolved, and installing the agent will not help while the port is closed.\n"+
					"If system performance figures are not needed, clear 'Collect Additional System Metrics' in the load configuration and run again.\n"+
					"(tried %d times with a %s timeout; last error: %v)",
				vm.PublicIP, metricAgentPort, metricAgentPort, precheckAttempts, precheckDialTimeout, err)
			rec.fail(constant.SubMetricPortOpen, msg, detail)
			rec.fail(constant.StepPrecheck, msg, detail)
			return fmt.Errorf("metric agent port %s is not reachable: %w", metricAgentPort, err)
		}
	} else {
		rec.skip(constant.SubMetricPortOpen, "Metrics not requested")
	}

	// ── remote command actually round-trips ─────────────────────────────────────────────
	//
	// Everything the run does to the target and the generator goes through cb-tumblebug's
	// remote command. If that does not work, every later step fails in a way that is much
	// harder to read than this one line.
	rec.begin(constant.SubRemoteCommand, "Running a test command on the target")
	if err := retryN(rec, constant.SubRemoteCommand, "Running a test command on the target",
		precheckAttempts, remoteRetryDelay, func() error {
			return l.probeRemoteCommand(ctx, param.NsId, param.InfraId, param.NodeId)
		}); err != nil {
		msg := "Remote command failed on the target"
		rec.fail(constant.SubRemoteCommand, msg, fmt.Sprintf(
			"asked cb-tumblebug to run 'echo' on ns=%s infra=%s node=%s as cb-user, %d times; last error: %v\n"+
				"everything the run does to the target goes through this path, so it stops here\n"+
				"if the node was started recently, ssh can take several minutes to accept connections - more so on small instance types - so try again shortly\n"+
				"if it has been up a while, check that port 22 is open and that the address cb-tumblebug holds for the node still matches the one the provider shows (a stop and start changes it)\n"+
				"before running again, confirm you can reach port 22 yourself - a plain ssh from outside settles whether the security group, any other firewall in front of the node, and the ssh service itself are all in order",
			param.NsId, param.InfraId, param.NodeId, precheckAttempts, err))
		rec.fail(constant.StepPrecheck, msg, err.Error())
		return fmt.Errorf("remote command failed: %w", err)
	}
	rec.ok(constant.SubRemoteCommand, "Remote command works")

	rec.ok(constant.StepPrecheck, "Environment looks good")
	return nil
}

// probeTarget sends the configured request once and returns the status code.
func probeTarget(ctx context.Context, reqs []RunLoadTestHttpParam) (int, error) {
	if len(reqs) == 0 {
		return 0, fmt.Errorf("no http request configured")
	}
	r := reqs[0]

	scheme := strings.ToLower(strings.TrimSpace(r.Protocol))
	if scheme == "" {
		scheme = "http"
	}
	path := r.Path
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	url := fmt.Sprintf("%s://%s:%s%s", scheme, r.Hostname, r.Port, path)

	method := strings.ToUpper(strings.TrimSpace(r.Method))
	if method == "" {
		method = http.MethodGet
	}

	ctx, cancel := context.WithTimeout(ctx, precheckHTTPTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return 0, err
	}
	resp, err := (&http.Client{Timeout: precheckHTTPTimeout}).Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

func targetURL(reqs []RunLoadTestHttpParam) string {
	if len(reqs) == 0 {
		return "(no request configured)"
	}
	r := reqs[0]
	scheme := strings.ToLower(strings.TrimSpace(r.Protocol))
	if scheme == "" {
		scheme = "http"
	}
	path := r.Path
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return fmt.Sprintf("%s://%s:%s%s", scheme, r.Hostname, r.Port, path)
}

func methodOf(reqs []RunLoadTestHttpParam) string {
	if len(reqs) == 0 || strings.TrimSpace(reqs[0].Method) == "" {
		return "GET"
	}
	return strings.ToUpper(strings.TrimSpace(reqs[0].Method))
}

func portOf(reqs []RunLoadTestHttpParam) string {
	if len(reqs) == 0 {
		return "?"
	}
	return reqs[0].Port
}

// isConnectionRefused reports whether the host answered and refused, as opposed to never
// answering. The two mean different things: refused is a process that is not there, a timeout
// is traffic that is not getting through.
func isConnectionRefused(err error) bool {
	return err != nil && strings.Contains(err.Error(), "connection refused")
}

func dial(host, port string) error {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), precheckDialTimeout)
	if err != nil {
		return err
	}
	return conn.Close()
}

// probeRemoteCommand runs a trivial command on the target and checks the answer comes back.
func (l *LoadService) probeRemoteCommand(ctx context.Context, nsId, infraId, nodeId string) error {
	ctx, cancel := context.WithTimeout(ctx, precheckRemoteTimeout)
	defer cancel()

	const marker = "cm-ant-precheck-ok"
	res, err := l.tumblebugClient.CommandToVmWithContext(ctx, nsId, infraId, nodeId, tumblebug.SendCommandReq{
		Command:  []string{"echo " + marker},
		UserName: "cb-user",
	})
	if err != nil {
		return err
	}
	if !strings.Contains(res, marker) {
		// The call returned but the command did not run — the distinction the install path
		// misses today, which is how an agent that never started was reported as installed.
		return fmt.Errorf("remote command returned without executing")
	}
	return nil
}

// verifyMetricAgent decides whether the agent is really up, and reinstalls it if it is not,
// rather than trusting the install call - which returns as soon as the remote command does,
// and reports success whether or not anything is left running (BAR-1552).
//
// Waiting alone is not enough, because two very different things look the same from outside:
// an install still working through apt and the agent zip, and an install that ended without
// leaving an agent behind. The first only needs more time; the second will never resolve on
// its own. So while the agent is missing this asks the target what is going on - if the
// install is still running, it keeps waiting up to the ceiling; if nothing is running, there
// is nothing to wait for and it goes straight to reinstalling.
//
// A reinstall gets one more round. Two rounds that both end without an agent are reported as
// a failure with what was seen each time, since a third would only spend more of the run's
// time arriving at the same answer.
func (l *LoadService) verifyMetricAgent(param RunLoadTestParam, rec *stepRecorder) error {
	rec.begin(constant.SubAgentProcess, "Checking the agent process")

	var lastErr error
	var attempts []string

	for round := 1; round <= agentInstallRounds; round++ {
		if round > 1 {
			// The install script kills whatever holds the port before it starts, so running it
			// again is the kill-and-reinstall - no separate teardown needed.
			rec.progress(constant.SubAgentProcess, round,
				fmt.Sprintf("Agent did not come up, reinstalling (round %d of %d)", round, agentInstallRounds),
				fmt.Sprintf("previous round: %s", attempts[len(attempts)-1]))

			arg := MonitoringAgentInstallationParams{NsId: param.NsId, InfraId: param.InfraId, NodeIds: []string{param.NodeId}}
			if _, err := l.InstallMonitoringAgent(arg); err != nil {
				lastErr = err
				attempts = append(attempts, fmt.Sprintf("round %d: reinstall itself failed: %v", round, err))
				continue
			}
		}

		outcome, err := l.waitForAgentProcess(param, rec, round)
		lastErr = err
		attempts = append(attempts, fmt.Sprintf("round %d: %s", round, outcome))
		if err != nil {
			continue
		}

		rec.ok(constant.SubAgentProcess, "Agent process is running")

		rec.begin(constant.SubAgentPort, "Waiting for the agent to answer")
		if err := retryN(rec, constant.SubAgentPort, "Waiting for the agent to answer", agentVerifyAttempts, agentVerifyDelay, func() error {
			return dial(param.AgentHostname, metricAgentPort)
		}); err != nil {
			lastErr = err
			attempts = append(attempts, fmt.Sprintf("round %d: process up but nothing answered on %s: %v", round, metricAgentPort, err))
			// The precheck already established that the port is not being filtered, so silence
			// with a process up means the agent did not bind - which a reinstall can fix.
			continue
		}
		rec.ok(constant.SubAgentPort, "Agent is answering")
		return nil
	}

	detail := fmt.Sprintf(
		"the metric agent on ns=%s infra=%s node=%s did not come up after %d install rounds:\n  %s\n"+
			"each round waited up to %s for the agent, checking every %s, and kept waiting only while the install was still running.\n"+
			"the install script starts the agent with nohup and reports success either way, so reaching here means it did not stay up.\n"+
			"check on the target that java is present, that %s holds the unpacked agent, and that nothing else is bound to port %s.\n"+
			"the install script's own account of each round is in %s, with the phase it reached in %s.\n"+
			"last error: %v",
		param.NsId, param.InfraId, param.NodeId, agentInstallRounds,
		strings.Join(attempts, "\n  "), agentUpCeiling, agentPollDelay, agentWorkDir, metricAgentPort,
		agentInstallLogFile, agentInstallStateFile, lastErr)
	rec.fail(constant.SubAgentProcess, "Metric agent could not be started", detail)
	if lastErr == nil {
		lastErr = fmt.Errorf("metric agent did not come up")
	}
	return lastErr
}

// waitForAgentProcess waits for the agent process to appear, giving up early when there is
// nothing left to wait for. It returns a sentence describing what it saw, for the failure
// detail - a wait that ended because the install was still going says something different
// from one that ended because nothing was running at all.
func (l *LoadService) waitForAgentProcess(param RunLoadTestParam, rec *stepRecorder, round int) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), agentUpCeiling+agentProbeGrace)
	defer cancel()

	deadline := time.Now().Add(agentUpCeiling)
	var lastErr error
	sawInstaller := false

	started := time.Now()

	for attempt := 1; ; attempt++ {
		if err := l.probeAgentProcess(ctx, param.NsId, param.InfraId, param.NodeId); err == nil {
			return fmt.Sprintf("agent came up after %d checks", attempt), nil
		} else {
			lastErr = err
		}

		installing, what := l.installProgress(ctx, param)
		if !installing {
			if sawInstaller {
				return fmt.Sprintf("the install stopped at '%s' and left no agent process", what),
					fmt.Errorf("install stopped at %s without leaving an agent: %w", what, lastErr)
			}
			return fmt.Sprintf("no agent and no install running (%s)", what),
				fmt.Errorf("no agent process and no install in progress: %w", lastErr)
		}
		sawInstaller = true

		if time.Now().After(deadline) {
			return fmt.Sprintf("still at '%s' after %s", what, agentUpCeiling),
				fmt.Errorf("install still at %s after %s without producing an agent: %w", what, agentUpCeiling, lastErr)
		}

		// The phase is the point of this message. "45s elapsed" only says time is passing;
		// "downloading the agent, 45s" says whether that is reasonable.
		rec.progress(constant.SubAgentProcess, round,
			fmt.Sprintf("Installing the agent - %s (%ds)", what, int(time.Since(started).Seconds())),
			fmt.Sprintf("the install is still running on the target; waiting up to %s", agentUpCeiling))
		time.Sleep(agentPollDelay)
	}
}

// installProgress answers whether an install is still going, and what it is doing.
//
// The script's own marker is the first source, because it is the only one that can report a
// failure: an install that died leaves no process, which is indistinguishable from one that
// never started. Targets whose agent predates the marker have no file, and there the process
// list is all there is - so it stays as a fallback rather than the primary answer.
//
// A target that cannot be asked at all is treated as still installing. A lost packet is not
// evidence that the install stopped, and reinstalling on one would interrupt work in progress.
func (l *LoadService) installProgress(ctx context.Context, param RunLoadTestParam) (bool, string) {
	st, err := l.probeAgentInstallState(ctx, param.NsId, param.InfraId, param.NodeId)
	if err != nil {
		return true, "could not read the install state"
	}
	if st.Phase != "" {
		if st.Phase == "failed" && st.Detail != "" {
			return false, fmt.Sprintf("install failed: %s", st.Detail)
		}
		return st.Running(), st.Describe()
	}

	installing, err := l.probeAgentInstalling(ctx, param.NsId, param.InfraId, param.NodeId)
	if err != nil {
		return true, "could not read the process list"
	}
	if installing {
		return true, "install running (no state recorded)"
	}
	return false, "no install recorded and nothing running"
}

const (
	agentVerifyAttempts = 5
	agentVerifyDelay    = 3 * time.Second

	// A healthy install lands well inside this; the ceiling is here so an install that never
	// finishes becomes an answer instead of an open-ended wait.
	agentUpCeiling  = 2 * time.Minute
	agentPollDelay  = 5 * time.Second
	agentProbeGrace = 30 * time.Second

	agentInstallRounds = 2

	agentWorkDir = "/opt/perfmon-agent"

	// Written by script/install-server-agent.sh as it goes. Kept outside the agent directory
	// so a failure before that directory exists is still recorded.
	agentInstallStateFile = "/var/tmp/cm-ant-agent-install.state"
	agentInstallLogFile   = "/var/tmp/cm-ant-agent-install.log"
)

// probeAgentProcess asks the target whether the agent process exists.
//
// The pattern matches how the agent actually appears in the process list, which is not what
// its name suggests: the release is distributed as ServerAgent-x.y.z.zip but runs as
//
//	java -jar /opt/perfmon-agent/CMDRunner.jar --tool PerfMonAgent --udp-port 0 --tcp-port 5555
//
// Looking for "ServerAgent" finds nothing on a target where the agent is running perfectly
// well - a check that is wrong in the direction that matters, since it reports a healthy
// agent as missing. The bracket keeps pgrep from matching the command carrying the pattern.
func (l *LoadService) probeAgentProcess(ctx context.Context, nsId, infraId, nodeId string) error {
	res, err := l.tumblebugClient.CommandToVmWithContext(ctx, nsId, infraId, nodeId, tumblebug.SendCommandReq{
		Command:  []string{"pgrep -f '[P]erfMonAgent' >/dev/null && echo agent-up || echo agent-down"},
		UserName: "cb-user",
	})
	if err != nil {
		return err
	}
	if !strings.Contains(res, "agent-up") {
		return fmt.Errorf("agent process not found on the target")
	}
	return nil
}

// agentInstallState is what the install script last said about itself.
type agentInstallState struct {
	Phase  string // preparing, dependencies, downloading, unpacking, present, starting, started, failed
	Detail string
	Raw    string
}

// Running reports whether the install still has work in front of it. "started" is not running:
// the script has done all it can and the agent either came up or did not.
func (s agentInstallState) Running() bool {
	switch s.Phase {
	case "preparing", "dependencies", "downloading", "unpacking", "present", "starting":
		return true
	default:
		return false
	}
}

// Describe renders the phase for a step message - what the target is actually doing, rather
// than a spinner that says only that time is passing.
func (s agentInstallState) Describe() string {
	switch s.Phase {
	case "preparing":
		return "preparing the target"
	case "dependencies":
		return "installing java and tools"
	case "downloading":
		return "downloading the agent"
	case "unpacking":
		return "unpacking the agent"
	case "present":
		return "agent already present"
	case "starting":
		return "starting the agent"
	case "started":
		return "agent start reported"
	case "failed":
		return "install failed"
	case "":
		return "no install recorded"
	default:
		return s.Phase
	}
}

// probeAgentInstallState reads the marker the install script writes as it goes.
//
// This is what separates "give it more time" from "it is never going to finish", and the
// script saying so directly beats inferring it from the process list: an apt running on the
// target is not necessarily ours, and a failed install leaves no process at all - it looks
// exactly like one that never started. The script records a phase before each stage and
// records "failed" from an ERR trap, so both of those become answers instead of guesses.
//
// A target whose agent was installed before this marker existed has no file. That is reported
// as an empty phase rather than an error, and the caller falls back to the process list.
func (l *LoadService) probeAgentInstallState(ctx context.Context, nsId, infraId, nodeId string) (agentInstallState, error) {
	res, err := l.tumblebugClient.CommandToVmWithContext(ctx, nsId, infraId, nodeId, tumblebug.SendCommandReq{
		Command:  []string{fmt.Sprintf("cat %s 2>/dev/null || true", agentInstallStateFile)},
		UserName: "cb-user",
	})
	if err != nil {
		return agentInstallState{}, err
	}

	line := lastNonEmptyLine(res)
	if line == "" {
		return agentInstallState{}, nil
	}

	// timestamp \t phase \t detail
	parts := strings.SplitN(line, "\t", 3)
	st := agentInstallState{Raw: line}
	if len(parts) > 1 {
		st.Phase = strings.TrimSpace(parts[1])
	}
	if len(parts) > 2 {
		st.Detail = strings.TrimSpace(parts[2])
	}
	return st, nil
}

// probeAgentInstalling falls back to the process list for targets whose install predates the
// state marker. It looks for what the install actually spends its time on: apt fetching a JRE,
// then the agent download and unpack, then the start script. The dpkg lock counts too, since
// an install blocked behind another package operation is still an install in progress.
func (l *LoadService) probeAgentInstalling(ctx context.Context, nsId, infraId, nodeId string) (bool, error) {
	res, err := l.tumblebugClient.CommandToVmWithContext(ctx, nsId, infraId, nodeId, tumblebug.SendCommandReq{
		Command: []string{
			"if pgrep -f '[a]pt-get' >/dev/null || pgrep -f '[d]pkg' >/dev/null || " +
				"pgrep -f '[w]get.*perfmon-agent' >/dev/null || pgrep -f '[u]nzip.*ServerAgent' >/dev/null || " +
				"pgrep -f '[s]tartAgent.sh' >/dev/null || fuser /var/lib/dpkg/lock-frontend >/dev/null 2>&1; " +
				"then echo install-running; else echo install-idle; fi",
		},
		UserName: "cb-user",
	})
	if err != nil {
		return false, err
	}
	return strings.Contains(res, "install-running"), nil
}

func lastNonEmptyLine(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if t := strings.TrimSpace(lines[i]); t != "" {
			return t
		}
	}
	return ""
}

// retryN is retry with the attempt count and delay supplied by the caller.
func retryN(rec *stepRecorder, step constant.ExecutionStep, what string, attempts int, delay time.Duration, check func() error) error {
	var err error
	for attempt := 1; attempt <= attempts; attempt++ {
		if attempt > 1 {
			rec.progress(step, attempt, fmt.Sprintf("%s (attempt %d of %d)", what, attempt, attempts),
				fmt.Sprintf("previous attempt failed: %v", err))
			time.Sleep(delay)
		}
		if err = check(); err == nil {
			return nil
		}
	}
	return err
}
