package load

import (
	"testing"
	"time"

	"github.com/cloud-barista/cm-ant/internal/core/common/constant"
)

func at(base time.Time, sec int) *time.Time {
	t := base.Add(time.Duration(sec) * time.Second)
	return &t
}

// A phase must report the span of the work underneath it. Recording the phase's own timing
// separately is what let a 27 minute wait hide behind a step that claimed to be running load.
func TestPhaseSpansItsChildren(t *testing.T) {
	base := time.Date(2026, 7, 19, 9, 0, 0, 0, time.UTC)
	now := base.Add(100 * time.Second)

	tree := buildStepTree([]LoadTestExecutionStep{
		{Seq: 3, Name: constant.StepAgentInstall, Status: constant.StepRunning},
		{Seq: 1, Name: constant.SubAgentInstall, Status: constant.StepOk, StartAt: at(base, 0), FinishAt: at(base, 10)},
		{Seq: 2, Name: constant.SubAgentProcess, Status: constant.StepOk, StartAt: at(base, 10), FinishAt: at(base, 12)},
		{Seq: 3, Name: constant.SubAgentPort, Status: constant.StepRunning, StartAt: at(base, 12)},
	}, now)

	if len(tree) != 1 {
		t.Fatalf("expected one phase, got %d", len(tree))
	}
	phase := tree[0]
	if len(phase.Children) != 3 {
		t.Fatalf("expected three sub-steps, got %d", len(phase.Children))
	}
	if phase.StartAt == nil || !phase.StartAt.Equal(base) {
		t.Errorf("phase should start when its first sub-step did, got %v", phase.StartAt)
	}
	if phase.FinishAt != nil {
		t.Errorf("phase is not finished while a sub-step still runs, got %v", phase.FinishAt)
	}
	if phase.ElapsedSec != 100 {
		t.Errorf("running phase should report time so far (100s), got %d", phase.ElapsedSec)
	}
	if got := phase.Children[2].ElapsedSec; got != 88 {
		t.Errorf("running sub-step should report time so far (88s), got %d", got)
	}
}

func TestFinishedPhaseReportsWholeSpan(t *testing.T) {
	base := time.Date(2026, 7, 19, 9, 0, 0, 0, time.UTC)
	now := base.Add(time.Hour) // long after the fact — must not affect a finished step

	tree := buildStepTree([]LoadTestExecutionStep{
		{Seq: 1, Name: constant.StepPrecheck, Status: constant.StepOk},
		{Seq: 1, Name: constant.SubTargetExists, Status: constant.StepOk, StartAt: at(base, 0), FinishAt: at(base, 1)},
		{Seq: 2, Name: constant.SubTargetReachable, Status: constant.StepOk, StartAt: at(base, 1), FinishAt: at(base, 4)},
	}, now)

	phase := tree[0]
	if phase.FinishAt == nil || !phase.FinishAt.Equal(base.Add(4*time.Second)) {
		t.Fatalf("phase should finish with its last sub-step, got %v", phase.FinishAt)
	}
	if phase.ElapsedSec != 4 {
		t.Errorf("finished phase should report its own span (4s), got %d", phase.ElapsedSec)
	}
}

// A phase with no sub-steps recorded still has to render — older rows look like this.
func TestPhaseWithoutChildrenKeepsOwnTiming(t *testing.T) {
	base := time.Date(2026, 7, 19, 9, 0, 0, 0, time.UTC)

	tree := buildStepTree([]LoadTestExecutionStep{
		{Seq: 4, Name: constant.StepJmeterRun, Status: constant.StepOk, StartAt: at(base, 0), FinishAt: at(base, 35)},
	}, base.Add(time.Minute))

	if len(tree) != 1 || len(tree[0].Children) != 0 {
		t.Fatalf("expected a childless phase, got %+v", tree)
	}
	if tree[0].ElapsedSec != 35 {
		t.Errorf("expected 35s, got %d", tree[0].ElapsedSec)
	}
}

// Phases come back in the order a run goes through them, whatever order the rows arrive in.
func TestPhasesAreOrdered(t *testing.T) {
	tree := buildStepTree([]LoadTestExecutionStep{
		{Seq: 5, Name: constant.StepJmeterRun, Status: constant.StepPending},
		{Seq: 1, Name: constant.StepPrecheck, Status: constant.StepPending},
		{Seq: 2, Name: constant.StepGeneratorInstall, Status: constant.StepPending},
	}, time.Now())

	want := []constant.ExecutionStep{constant.StepPrecheck, constant.StepGeneratorInstall, constant.StepJmeterRun}
	for i, w := range want {
		if tree[i].Name != w {
			t.Errorf("position %d: expected %s, got %s", i, w, tree[i].Name)
		}
	}
}

func TestParent(t *testing.T) {
	if got := constant.SubAgentPort.Parent(); got != constant.StepAgentInstall {
		t.Errorf("expected %s, got %s", constant.StepAgentInstall, got)
	}
	if got := constant.StepAgentInstall.Parent(); got != "" {
		t.Errorf("a phase has no parent, got %s", got)
	}
}

// A step that never started reports nothing rather than a span measured from the epoch.
func TestNotStartedReportsNothing(t *testing.T) {
	if got := elapsedSec(nil, nil, time.Now()); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}
