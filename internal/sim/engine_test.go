package sim

import "testing"

func TestLightCycleRespectsConfiguredDurations(t *testing.T) {
	cfg := Config{
		Name:  "cycle-test",
		Steps: 6,
		Grid: GridConfig{
			Width:  20,
			Height: 10,
		},
		Signal: SignalConfig{
			VerticalGreenSteps:   2,
			HorizontalGreenSteps: 1,
		},
		Spawn: SpawnConfig{
			Lanes: map[Direction]LaneSpawnConfig{
				Up: {
					EntryX:       10,
					EntryY:       9,
					StepInterval: 0,
				},
			},
		},
	}

	engine, err := NewEngine(cfg)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	report := engine.Run(true, boolPtr(false))
	if len(report.Timeline) != cfg.Steps {
		t.Fatalf("timeline length = %d, want %d", len(report.Timeline), cfg.Steps)
	}

	got := make([]bool, 0, cfg.Steps)
	for _, snap := range report.Timeline {
		got = append(got, snap.LightGreen)
	}
	want := []bool{true, true, false, true, true, false}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("step %d: light vertical green = %v, want %v", i+1, got[i], want[i])
		}
	}
}

func TestRedLightCreatesWaitForHorizontalVehicle(t *testing.T) {
	cfg := Config{
		Name:  "wait-test",
		Steps: 20,
		Grid: GridConfig{
			Width:  20,
			Height: 10,
		},
		Signal: SignalConfig{
			VerticalGreenSteps:   5,
			HorizontalGreenSteps: 1,
		},
		Spawn: SpawnConfig{
			Lanes: map[Direction]LaneSpawnConfig{
				Right: {
					EntryX:       9,
					EntryY:       5,
					StepInterval: 1,
					MaxVehicles:  1,
				},
			},
		},
	}

	engine, err := NewEngine(cfg)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	report := engine.Run(false, boolPtr(false))
	if report.Metrics.BlockedBySignal == 0 {
		t.Fatalf("expected blocked-by-signal count > 0")
	}
	if report.Metrics.VehiclesCompleted == 0 {
		t.Fatalf("expected completed vehicles > 0")
	}
}

func TestTrafficConflictRaisesPotentialCollisionMetric(t *testing.T) {
	cfg := Config{
		Name:  "conflict-test",
		Steps: 1,
		Grid: GridConfig{
			Width:  20,
			Height: 10,
		},
		Signal: SignalConfig{VerticalGreenSteps: 10, HorizontalGreenSteps: 10},
		Spawn: SpawnConfig{
			Lanes: map[Direction]LaneSpawnConfig{
				Right: {
					EntryX:       9,
					EntryY:       4,
					StepInterval: 1,
					MaxVehicles:  1,
				},
				Left: {
					EntryX:       11,
					EntryY:       4,
					StepInterval: 1,
					MaxVehicles:  1,
				},
			},
		},
	}

	engine, err := NewEngine(cfg)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	report := engine.Run(false, boolPtr(false))
	if report.Metrics.PotentialCollisions == 0 {
		t.Fatalf("expected potential collision > 0")
	}
}

func TestPlatoonVehiclesCanAdvanceInSameStep(t *testing.T) {
	cfg := Config{
		Name:  "platoon-test",
		Steps: 1,
		Grid: GridConfig{
			Width:  20,
			Height: 10,
		},
		Signal: SignalConfig{
			VerticalGreenSteps:   5,
			HorizontalGreenSteps: 5,
		},
		Spawn: SpawnConfig{
			Lanes: map[Direction]LaneSpawnConfig{
				Right: {
					EntryX:       0,
					EntryY:       5,
					StepInterval: 0,
				},
			},
		},
	}

	engine, err := NewEngine(cfg)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	engine.vehicles = []Vehicle{
		{ID: 1, X: 1, Y: 5, Direction: Right, SpawnStep: 1},
		{ID: 2, X: 2, Y: 5, Direction: Right, SpawnStep: 1},
	}

	engine.moveVehicles(0)

	if len(engine.vehicles) != 2 {
		t.Fatalf("expected 2 active vehicles, got %d", len(engine.vehicles))
	}
	if engine.vehicles[0].X != 2 || engine.vehicles[1].X != 3 {
		t.Fatalf("expected vehicles to advance to x=2 and x=3, got x=%d and x=%d", engine.vehicles[0].X, engine.vehicles[1].X)
	}
	if engine.blockedTraffic != 0 {
		t.Fatalf("expected no traffic blocking, got %d", engine.blockedTraffic)
	}
}

func boolPtr(v bool) *bool {
	return &v
}
