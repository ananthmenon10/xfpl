// Copyright 2026 ananth-menon. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"math"
	"testing"
)

func approx(a, b, eps float64) bool {
	return math.Abs(a-b) < eps
}

func TestScoreCaptain_FDRMonotonic(t *testing.T) {
	// Holding everything else constant, an easier fixture (low FDR) must score
	// higher than a harder one.
	easy := scoreCaptain(7.0, 2.0, 0.5, 2000, 0, true)
	hard := scoreCaptain(7.0, 5.0, 0.5, 2000, 0, true)
	if !(easy > hard) {
		t.Fatalf("expected easy(%v) > hard(%v)", easy, hard)
	}
}

func TestScoreCaptain_HomeBoost(t *testing.T) {
	home := scoreCaptain(7.0, 3.0, 0.5, 2000, 0, true)
	away := scoreCaptain(7.0, 3.0, 0.5, 2000, 0, false)
	if !(home > away) {
		t.Fatalf("expected home(%v) > away(%v)", home, away)
	}
}

func TestScoreCaptain_XGIBoost(t *testing.T) {
	withXG := scoreCaptain(7.0, 3.0, 1.0, 2000, 0, true)
	noXG := scoreCaptain(7.0, 3.0, 0.0, 2000, 0, true)
	if !(withXG > noXG) {
		t.Fatalf("expected withXG(%v) > noXG(%v)", withXG, noXG)
	}
}

func TestScoreCaptain_MinutesRisk(t *testing.T) {
	regular := scoreCaptain(7.0, 3.0, 0.5, 2000, 0, true)
	rotation := scoreCaptain(7.0, 3.0, 0.5, 400, 0, true)
	if !(regular > rotation) {
		t.Fatalf("expected regular(%v) > rotation(%v)", regular, rotation)
	}
	if !approx(captainMinRisk(2000), 1.0, 0.01) {
		t.Fatalf("expected min risk = 1.0 for full-season starter, got %v", captainMinRisk(2000))
	}
	if !approx(captainMinRisk(100), 0.5, 0.01) {
		t.Fatalf("expected min risk = 0.5 floor, got %v", captainMinRisk(100))
	}
}

func TestScoreCaptain_DGW(t *testing.T) {
	single := scoreCaptain(7.0, 3.0, 0.5, 2000, 0, true)
	dgw := scoreCaptain(7.0, 3.0, 0.5, 2000, 1, true)
	if !(dgw > single*1.5) {
		t.Fatalf("expected DGW(%v) to be > 1.5x single(%v)", dgw, single)
	}
}

func TestScoreCaptain_ZeroEP(t *testing.T) {
	got := scoreCaptain(0, 3.0, 0.5, 2000, 0, true)
	if got != 0 {
		t.Fatalf("expected 0 when ep_next is 0, got %v", got)
	}
}

func TestResolveNextGW(t *testing.T) {
	events := []captainEvent{
		{ID: 36, Finished: true},
		{ID: 37, IsCurrent: true, Finished: false},
		{ID: 38, IsNext: true},
	}
	gw, err := resolveNextGW(events, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gw != 38 {
		t.Fatalf("expected next GW 38, got %d", gw)
	}

	gw, err = resolveNextGW(events, 30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gw != 30 {
		t.Fatalf("override should win, got %d", gw)
	}

	gw, err = resolveNextGW([]captainEvent{}, 0)
	if err == nil {
		t.Fatalf("expected error on empty events, got %d", gw)
	}
}
