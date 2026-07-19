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
)

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
	if code, err := probeTarget(ctx, param.HttpReqs); err != nil {
		msg := fmt.Sprintf("Cannot reach the target on port %s. Open the port in the security group, or check that the service is running.", portOf(param.HttpReqs))
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
		if err := dial(vm.PublicIP, metricAgentPort); err != nil {
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
	if err := l.probeRemoteCommand(ctx, param.NsId, param.InfraId, param.NodeId); err != nil {
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
