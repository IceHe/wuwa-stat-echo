package goapp

import "testing"

func TestComputeMaxGapFromEventsUsesPlayerInternalEchoSequence(t *testing.T) {
	state, total := computeMaxGapFromEvents([]maxGapEvent{
		{id: 10, userID: 1, hit: false},
		{id: 11, userID: 1, hit: true},
		{id: 12, userID: 1, hit: false},
		{id: 13, userID: 1, hit: false},
		{id: 14, userID: 1, hit: true},
		{id: 20, userID: 2, hit: false},
		{id: 21, userID: 2, hit: false},
		{id: 22, userID: 2, hit: false},
	})

	if total != 8 {
		t.Fatalf("expected total echo events 8, got %d", total)
	}
	if state.userID != 2 || state.maxGap != 3 || state.count != 0 {
		t.Fatalf("expected user 2 no-hit gap 3, got state=%+v", state)
	}
	if state.startID != -1 || state.endID != -1 {
		t.Fatalf("expected no-hit gap edges to be unset, got start=%d end=%d", state.startID, state.endID)
	}
}

func TestComputeMaxGapFromEventsTracksBetweenHitEdges(t *testing.T) {
	state, _ := computeMaxGapFromEvents([]maxGapEvent{
		{id: 30, userID: 3, hit: true},
		{id: 31, userID: 3, hit: false},
		{id: 32, userID: 3, hit: false},
		{id: 33, userID: 3, hit: true},
		{id: 34, userID: 3, hit: false},
	})

	if state.userID != 3 || state.maxGap != 2 || state.count != 2 {
		t.Fatalf("expected user 3 between-hit gap 2, got state=%+v", state)
	}
	if state.startID != 30 || state.endID != 33 {
		t.Fatalf("expected gap edge echo ids 30 -> 33, got %d -> %d", state.startID, state.endID)
	}
}
