package load

import "testing"

// These tests exist because cb-tumblebug node ids are names, not identities. A node id is
// built as "{group name}-{index within the group}", so deleting a VM and recreating it under
// the same name yields an identical ns/infra/node triple. Runs recorded against that triple
// alone cannot be attributed to a particular VM; only the uid, regenerated per VM, can.

// TestBelongsToDifferentNode covers the decision that guards stopping a load test.
//
// Getting this wrong is destructive in one direction and obstructive in the other: too
// lenient and a stop request kills a run belonging to the VM's predecessor, too strict and
// legitimate stops start failing for runs recorded before the uid existed.
func TestBelongsToDifferentNode(t *testing.T) {
	cases := []struct {
		name        string
		recordedUid string
		currentUid  string
		want        bool
		why         string
	}{
		{
			name:        "same vm",
			recordedUid: "tbaaa111",
			currentUid:  "tbaaa111",
			want:        false,
			why:         "the run belongs to the VM being asked about",
		},
		{
			name:        "recreated under the same name",
			recordedUid: "tbaaa111",
			currentUid:  "tbbbb222",
			want:        true,
			why:         "same names, different VM - this is the case worth refusing",
		},
		{
			name:        "run predates the uid column",
			recordedUid: "",
			currentUid:  "tbbbb222",
			want:        false,
			why:         "unknown is not a mismatch; refusing would block old but valid runs",
		},
		{
			name:        "cb-tumblebug could not be reached",
			recordedUid: "tbaaa111",
			currentUid:  "",
			want:        false,
			why:         "a failed lookup says nothing about which VM this is",
		},
		{
			name:        "neither side known",
			recordedUid: "",
			currentUid:  "",
			want:        false,
			why:         "no evidence either way",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := belongsToDifferentNode(c.recordedUid, c.currentUid); got != c.want {
				t.Errorf("belongsToDifferentNode(%q, %q) = %v, want %v — %s",
					c.recordedUid, c.currentUid, got, c.want, c.why)
			}
		})
	}
}

// TestMapLoadTestExecutionStateResult_ExposesNodeUid makes sure callers can tell which VM a
// run belongs to. Without the uid in the response the console cannot notice that "the last
// run for this node" describes a VM that has since been replaced.
func TestMapLoadTestExecutionStateResult_ExposesNodeUid(t *testing.T) {
	got := mapLoadTestExecutionStateResult(LoadTestExecutionState{
		LoadTestKey: "key-1",
		NodeUid:     "tbabc123def456ghi789",
	})

	if got.NodeUid != "tbabc123def456ghi789" {
		t.Errorf("expected the node uid to reach the response, got %q", got.NodeUid)
	}
}
