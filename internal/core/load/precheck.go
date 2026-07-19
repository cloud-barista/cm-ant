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
	"github.com/rs/zerolog/log"
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
	precheckDialTimeout   = 5 * time.Second
	precheckHTTPTimeout   = 10 * time.Second
	precheckRemoteTimeout = 30 * time.Second
	metricAgentPort       = "5555"

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

// precheckOutcome carries what the run needs to know afterwards.
type precheckOutcome struct {
	// MetricsReachable is false when the metric agent port is closed. The run continues —
	// load figures are worth having without CPU and memory alongside them — but it is said
	// up front rather than discovered as missing files at the end.
	MetricsReachable bool
}

func (l *LoadService) runPrecheck(ctx context.Context, param RunLoadTestParam, rec *stepRecorder) (precheckOutcome, error) {
	out := precheckOutcome{MetricsReachable: true}
	rec.begin(constant.StepPrecheck, "Checking the environment")

	// ── the target exists and is running ────────────────────────────────────────────────
	rec.begin(constant.SubTargetExists, "Looking up the target node")
	vm, err := l.tumblebugClient.GetVmWithContext(ctx, param.NsId, param.InfraId, param.NodeId)
	if err != nil {
		msg := "Target node not found. Check the namespace, infra and node ids."
		rec.fail(constant.SubTargetExists, msg, err.Error())
		rec.fail(constant.StepPrecheck, msg, err.Error())
		return out, fmt.Errorf("target node not found: %w", err)
	}
	rec.ok(constant.SubTargetExists, "Target node found")

	rec.begin(constant.SubTargetRunning, "Checking the target state")
	if !strings.EqualFold(vm.Status, "Running") {
		msg := fmt.Sprintf("Target node is not running (currently %s).", vm.Status)
		rec.fail(constant.SubTargetRunning, msg, "")
		rec.fail(constant.StepPrecheck, msg, "")
		return out, fmt.Errorf("target node is not running: %s", vm.Status)
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
		msg := fmt.Sprintf("Cannot reach the target on port %s after %d attempts. Open the port in the security group, or check that the service is running.", portOf(param.HttpReqs), precheckAttempts)
		rec.fail(constant.SubTargetReachable, msg, err.Error())
		rec.fail(constant.StepPrecheck, msg, err.Error())
		return out, fmt.Errorf("target is not reachable: %w", err)
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
	if param.CollectAdditionalSystemMetrics {
		rec.begin(constant.SubMetricPortOpen, "Checking the metric agent port")
		if err := retry(rec, constant.SubMetricPortOpen, "Checking the metric agent port", func() error {
			return dial(vm.PublicIP, metricAgentPort)
		}); err != nil {
			out.MetricsReachable = false
			msg := fmt.Sprintf("Metric port %s is closed, so the run continues without CPU and memory figures. Add %s inbound to the security group to collect them.", metricAgentPort, metricAgentPort)
			// Not a failure: the load test itself is unaffected.
			rec.skip(constant.SubMetricPortOpen, msg)
			log.Warn().Msgf("metric agent port unreachable on %s: %v", vm.PublicIP, err)
		} else {
			rec.ok(constant.SubMetricPortOpen, "Metric agent port is reachable")
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
	if err := retry(rec, constant.SubRemoteCommand, "Running a test command on the target", func() error {
		return l.probeRemoteCommand(ctx, param.NsId, param.InfraId, param.NodeId)
	}); err != nil {
		msg := "Cannot run remote commands on the target node. Check SSH access and the node agent."
		rec.fail(constant.SubRemoteCommand, msg, err.Error())
		rec.fail(constant.StepPrecheck, msg, err.Error())
		return out, fmt.Errorf("remote command failed: %w", err)
	}
	rec.ok(constant.SubRemoteCommand, "Remote command works")

	rec.ok(constant.StepPrecheck, "Environment looks good")
	return out, nil
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

func portOf(reqs []RunLoadTestHttpParam) string {
	if len(reqs) == 0 {
		return "?"
	}
	return reqs[0].Port
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

// verifyMetricAgent waits for the agent to actually be up rather than trusting the install
// call, which returns as soon as the remote command does (BAR-1552).
//
// The agent is started with nohup, so there is a gap between the script returning and the
// port accepting connections. Short repeated checks cover that gap; giving up after them is
// what turns "never going to answer" into an answer, instead of a wait with no end.
func (l *LoadService) verifyMetricAgent(param RunLoadTestParam, rec *stepRecorder) error {
	ctx, cancel := context.WithTimeout(context.Background(), agentVerifyTimeout)
	defer cancel()

	rec.begin(constant.SubAgentProcess, "Checking the agent process")
	if err := retryN(rec, constant.SubAgentProcess, "Checking the agent process", agentVerifyAttempts, agentVerifyDelay, func() error {
		return l.probeAgentProcess(ctx, param.NsId, param.InfraId, param.NodeId)
	}); err != nil {
		rec.fail(constant.SubAgentProcess, "The metric agent is not running on the target", err.Error())
		return err
	}
	rec.ok(constant.SubAgentProcess, "Agent process is running")

	rec.begin(constant.SubAgentPort, "Waiting for the agent to answer")
	if err := retryN(rec, constant.SubAgentPort, "Waiting for the agent to answer", agentVerifyAttempts, agentVerifyDelay, func() error {
		return dial(param.AgentHostname, metricAgentPort)
	}); err != nil {
		rec.fail(constant.SubAgentPort, fmt.Sprintf("The metric agent is not answering on port %s", metricAgentPort), err.Error())
		return err
	}
	rec.ok(constant.SubAgentPort, "Agent is answering")
	return nil
}

const (
	agentVerifyAttempts = 5
	agentVerifyDelay    = 3 * time.Second
	agentVerifyTimeout  = 60 * time.Second
)

// probeAgentProcess asks the target whether the agent process exists.
func (l *LoadService) probeAgentProcess(ctx context.Context, nsId, infraId, nodeId string) error {
	res, err := l.tumblebugClient.CommandToVmWithContext(ctx, nsId, infraId, nodeId, tumblebug.SendCommandReq{
		Command:  []string{"pgrep -f ServerAgent >/dev/null && echo agent-up || echo agent-down"},
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
