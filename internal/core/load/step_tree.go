package load

import (
	"sort"
	"time"

	"github.com/cloud-barista/cm-ant/internal/core/common/constant"
)

// buildStepTree turns the flat step rows into the phase/sub-step shape the console renders,
// and stamps every node with how long it has taken (BAR-1553).
//
// The flat list could not explain where a run's time went: a run stuck for 27 minutes unable
// to reach the metric agent and a run busy generating load both showed "jmeter_run: running".
// Sub-steps split that span, and the elapsed figure tells a slow step from a stuck one.
//
// A phase's own timing is derived from its children rather than recorded separately, so the
// two can never disagree.
func buildStepTree(steps []LoadTestExecutionStep, now time.Time) []LoadTestExecutionStepResult {
	phases := make(map[constant.ExecutionStep]*LoadTestExecutionStepResult)
	var order []constant.ExecutionStep

	phaseOf := func(name constant.ExecutionStep) *LoadTestExecutionStepResult {
		if p, ok := phases[name]; ok {
			return p
		}
		p := &LoadTestExecutionStepResult{Name: name, Status: constant.StepPending}
		phases[name] = p
		order = append(order, name)
		return p
	}

	for _, s := range steps {
		node := LoadTestExecutionStepResult{
			Seq:        s.Seq,
			Name:       s.Name,
			Status:     s.Status,
			Attempt:    s.Attempt,
			StartAt:    s.StartAt,
			FinishAt:   s.FinishAt,
			Message:    s.Message,
			Detail:     s.Detail,
			ElapsedSec: elapsedSec(s.StartAt, s.FinishAt, now),
		}

		if parent := s.Name.Parent(); parent != "" {
			p := phaseOf(parent)
			p.Children = append(p.Children, node)
			continue
		}

		p := phaseOf(s.Name)
		// Keep the children already collected; the row itself supplies the rest.
		children := p.Children
		*p = node
		p.Children = children
	}

	out := make([]LoadTestExecutionStepResult, 0, len(order))
	for _, name := range order {
		p := phases[name]
		sort.SliceStable(p.Children, func(i, j int) bool { return p.Children[i].Seq < p.Children[j].Seq })
		rollUp(p, now)
		out = append(out, *p)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Seq < out[j].Seq })
	return out
}

// rollUp gives a phase the span of its children. Without this a phase would report the moment
// its own row was written, which says nothing about how long the work underneath took.
func rollUp(p *LoadTestExecutionStepResult, now time.Time) {
	if len(p.Children) == 0 {
		return
	}

	start := p.StartAt
	var finish *time.Time
	allDone := true

	for i := range p.Children {
		c := p.Children[i]
		if c.StartAt != nil && (start == nil || c.StartAt.Before(*start)) {
			start = c.StartAt
		}
		if c.FinishAt != nil && (finish == nil || c.FinishAt.After(*finish)) {
			finish = c.FinishAt
		}
		if c.Status == constant.StepPending || c.Status == constant.StepRunning {
			allDone = false
		}
	}

	p.StartAt = start
	// A phase is only finished once nothing under it is still going.
	if allDone {
		p.FinishAt = finish
	} else {
		p.FinishAt = nil
	}
	p.ElapsedSec = elapsedSec(p.StartAt, p.FinishAt, now)
}

// elapsedSec reports the whole span for a finished step, and the time so far for one still
// running. A step that has not started yet reports nothing.
func elapsedSec(start, finish *time.Time, now time.Time) int {
	if start == nil {
		return 0
	}
	end := now
	if finish != nil {
		end = *finish
	}
	if end.Before(*start) {
		return 0
	}
	return int(end.Sub(*start).Seconds())
}
