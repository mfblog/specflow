// Package next provides the deterministic directive for the current governance step.
package next

import (
	"strings"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/statusfile"
)

// UnitState classifies a unit's current lifecycle state.
type UnitState string

const (
	StateStableIdle       UnitState = "stable_idle"
	StateStableVerify     UnitState = "stable_verify"
	StateCandidateCheck   UnitState = "candidate_check"
	StateCandidatePending UnitState = "candidate_pending_impl"
	StateCandidateVerify  UnitState = "candidate_verify"
	StateCandidatePromote UnitState = "candidate_promote"
	StateUnregistered     UnitState = "unregistered"
)

// ClassifyUnitState reads a single unit status row and returns the matching UnitState.
func ClassifyUnitState(s statusfile.ObjectStatus) UnitState {
	if s.Object == "" {
		return StateUnregistered
	}
	stable := strings.TrimSpace(strings.ToLower(s.Stable))
	candidate := strings.TrimSpace(strings.ToLower(s.Candidate))
	active := strings.TrimSpace(strings.ToLower(s.ActiveLayer))
	next := strings.TrimSpace(strings.ToLower(s.NextCommand))
	notes := strings.TrimSpace(strings.ToLower(s.Notes))

	switch {
	case stable == "yes" && candidate == "no" && active == "stable" && next == "unit_fork":
		return StateStableIdle
	case stable == "yes" && candidate == "no" && active == "stable" && next == "unit_stable_verify":
		return StateStableVerify
	case next == "unit_init":
		return StateUnregistered
	case next == "unit_new":
		return StateCandidateCheck
	case candidate == "yes" && next == "unit_check":
		return StateCandidateCheck
	case candidate == "yes" && next == "unit_verify" && strings.Contains(notes, "pending_impl"):
		return StateCandidatePending
	case candidate == "yes" && next == "unit_verify":
		return StateCandidateVerify
	case candidate == "yes" && next == "unit_promote":
		return StateCandidatePromote
	default:
		return StateUnregistered
	}
}
