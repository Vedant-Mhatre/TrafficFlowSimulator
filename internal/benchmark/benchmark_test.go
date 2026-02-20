package benchmark

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Vedant-Mhatre/TrafficFlowSimulator/internal/sim"
)

func TestAnalyzeTimelineComputesJerkAndHardBrake(t *testing.T) {
	timeline := []sim.StepSnapshot{
		{
			Step: 1,
			Vehicles: []sim.Vehicle{
				{ID: 1, X: 0, Y: 5, Direction: sim.Right},
			},
		},
		{
			Step: 2,
			Vehicles: []sim.Vehicle{
				{ID: 1, X: 1, Y: 5, Direction: sim.Right},
			},
		},
		{
			Step: 3,
			Vehicles: []sim.Vehicle{
				{ID: 1, X: 1, Y: 5, Direction: sim.Right},
			},
		},
	}

	minTTC, meanJerk, hardBrakes := analyzeTimeline(timeline)
	if minTTC != noClosingTTC {
		t.Fatalf("minTTC = %.2f, want %.2f when no closing pairs", minTTC, noClosingTTC)
	}
	if hardBrakes != 1 {
		t.Fatalf("hardBrakes = %d, want 1", hardBrakes)
	}
	if meanJerk <= 0 {
		t.Fatalf("meanJerk = %.2f, want > 0", meanJerk)
	}
}

func TestMinTTCStepDetectsClosingPair(t *testing.T) {
	vehicles := []sim.Vehicle{
		{ID: 1, X: 3, Y: 5, Direction: sim.Right}, // leader
		{ID: 2, X: 1, Y: 5, Direction: sim.Right}, // follower
	}
	speeds := map[int]float64{
		1: 0,
		2: 1,
	}

	ttc := minTTCStep(vehicles, speeds)
	if ttc != 2 {
		t.Fatalf("ttc = %.2f, want 2.00", ttc)
	}
}

func TestEvaluateChecksFailOnRegression(t *testing.T) {
	spec := Spec{
		Name: "regression-check",
		Thresholds: Thresholds{
			MaxCollisionIncrease: 0,
			MaxDelayIncrease:     0.1,
			MinThroughputRatio:   1.0,
			MaxJerkIncrease:      0.1,
			MaxMinTTCDrop:        0.1,
		},
	}

	base := Scorecard{
		ThroughputPer100:    50,
		AverageDelay:        5,
		PotentialCollisions: 0,
		MeanAbsJerk:         1,
		MinTTCSteps:         4,
	}
	candidate := Scorecard{
		ThroughputPer100:    45,
		AverageDelay:        7,
		PotentialCollisions: 2,
		MeanAbsJerk:         2,
		MinTTCSteps:         2,
	}

	result := evaluate(spec, base, candidate)
	if result.Passed {
		t.Fatalf("expected benchmark to fail on regression")
	}
}

func TestLoadSpecResolvesRelativePaths(t *testing.T) {
	temp := t.TempDir()
	specPath := filepath.Join(temp, "bench.json")
	content := `{
		"name": "path-test",
		"baseline_config": "baseline.json",
		"candidate_config": "candidate.json",
		"report_path": "reports/out.json",
		"thresholds": {
			"min_throughput_ratio": 0.95
		}
	}`
	if err := os.WriteFile(specPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	spec, err := LoadSpec(specPath)
	if err != nil {
		t.Fatalf("load spec: %v", err)
	}

	if spec.BaselineConfig != filepath.Join(temp, "baseline.json") {
		t.Fatalf("unexpected baseline path: %s", spec.BaselineConfig)
	}
	if spec.CandidateConfig != filepath.Join(temp, "candidate.json") {
		t.Fatalf("unexpected candidate path: %s", spec.CandidateConfig)
	}
	if spec.ReportPath != filepath.Join(temp, "reports", "out.json") {
		t.Fatalf("unexpected report path: %s", spec.ReportPath)
	}
}
