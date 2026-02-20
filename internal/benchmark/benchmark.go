package benchmark

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/Vedant-Mhatre/TrafficFlowSimulator/internal/sim"
)

const noClosingTTC = 1_000_000.0

type Spec struct {
	Name            string     `json:"name"`
	BaselineConfig  string     `json:"baseline_config"`
	CandidateConfig string     `json:"candidate_config"`
	Thresholds      Thresholds `json:"thresholds"`
	ReportPath      string     `json:"report_path"`
}

type Thresholds struct {
	MaxCollisionIncrease int     `json:"max_collision_increase"`
	MaxDelayIncrease     float64 `json:"max_delay_increase"`
	MinThroughputRatio   float64 `json:"min_throughput_ratio"`
	MaxJerkIncrease      float64 `json:"max_jerk_increase"`
	MaxMinTTCDrop        float64 `json:"max_min_ttc_drop"`
}

type Scorecard struct {
	ScenarioName        string  `json:"scenario_name"`
	VehiclesCompleted   int     `json:"vehicles_completed"`
	ThroughputPer100    float64 `json:"throughput_per_100_steps"`
	AverageDelay        float64 `json:"average_delay_steps"`
	PotentialCollisions int     `json:"potential_collisions"`
	MinTTCSteps         float64 `json:"min_ttc_steps"`
	MeanAbsJerk         float64 `json:"mean_abs_jerk"`
	HardBrakes          int     `json:"hard_brakes"`
}

type CheckResult struct {
	Name      string  `json:"name"`
	Rule      string  `json:"rule"`
	Baseline  float64 `json:"baseline"`
	Candidate float64 `json:"candidate"`
	Passed    bool    `json:"passed"`
}

type Result struct {
	Name      string        `json:"name"`
	Generated time.Time     `json:"generated"`
	Baseline  Scorecard     `json:"baseline"`
	Candidate Scorecard     `json:"candidate"`
	Checks    []CheckResult `json:"checks"`
	Passed    bool          `json:"passed"`
}

func LoadSpec(path string) (Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Spec{}, fmt.Errorf("read benchmark spec: %w", err)
	}

	var spec Spec
	if err := json.Unmarshal(data, &spec); err != nil {
		return Spec{}, fmt.Errorf("parse benchmark spec: %w", err)
	}

	applySpecDefaults(&spec)
	resolveSpecPaths(&spec, filepath.Dir(path))
	if err := validateSpec(spec); err != nil {
		return Spec{}, err
	}
	return spec, nil
}

func applySpecDefaults(spec *Spec) {
	if spec.Name == "" {
		spec.Name = "deterministic-benchmark"
	}
	if spec.Thresholds.MinThroughputRatio <= 0 {
		spec.Thresholds.MinThroughputRatio = 0.98
	}
	if spec.Thresholds.MaxDelayIncrease < 0 {
		spec.Thresholds.MaxDelayIncrease = 0
	}
	if spec.Thresholds.MaxCollisionIncrease < 0 {
		spec.Thresholds.MaxCollisionIncrease = 0
	}
	if spec.Thresholds.MaxJerkIncrease < 0 {
		spec.Thresholds.MaxJerkIncrease = 0
	}
	if spec.Thresholds.MaxMinTTCDrop < 0 {
		spec.Thresholds.MaxMinTTCDrop = 0
	}
}

func resolveSpecPaths(spec *Spec, baseDir string) {
	if baseDir == "" {
		return
	}
	if spec.BaselineConfig != "" && !filepath.IsAbs(spec.BaselineConfig) {
		spec.BaselineConfig = filepath.Join(baseDir, spec.BaselineConfig)
	}
	if spec.CandidateConfig != "" && !filepath.IsAbs(spec.CandidateConfig) {
		spec.CandidateConfig = filepath.Join(baseDir, spec.CandidateConfig)
	}
	if spec.ReportPath != "" && !filepath.IsAbs(spec.ReportPath) {
		spec.ReportPath = filepath.Join(baseDir, spec.ReportPath)
	}
}

func validateSpec(spec Spec) error {
	if spec.BaselineConfig == "" {
		return fmt.Errorf("baseline_config is required")
	}
	if spec.CandidateConfig == "" {
		return fmt.Errorf("candidate_config is required")
	}
	if spec.Thresholds.MinThroughputRatio <= 0 {
		return fmt.Errorf("threshold min_throughput_ratio must be > 0")
	}
	return nil
}

func Run(spec Spec) (Result, error) {
	baseScore, err := runScenario(spec.BaselineConfig)
	if err != nil {
		return Result{}, fmt.Errorf("run baseline scenario: %w", err)
	}
	candScore, err := runScenario(spec.CandidateConfig)
	if err != nil {
		return Result{}, fmt.Errorf("run candidate scenario: %w", err)
	}

	result := evaluate(spec, baseScore, candScore)
	if spec.ReportPath != "" {
		if err := writeResult(spec.ReportPath, result); err != nil {
			return Result{}, err
		}
	}
	return result, nil
}

func runScenario(configPath string) (Scorecard, error) {
	cfg, err := sim.LoadConfig(configPath)
	if err != nil {
		return Scorecard{}, fmt.Errorf("load config %q: %w", configPath, err)
	}

	engine, err := sim.NewEngine(cfg)
	if err != nil {
		return Scorecard{}, fmt.Errorf("create engine for %q: %w", cfg.Name, err)
	}

	render := false
	report := engine.Run(true, &render)
	if cfg.ReportPath != "" {
		if err := sim.WriteReport(cfg.ReportPath, report); err != nil {
			return Scorecard{}, fmt.Errorf("write scenario report %q: %w", cfg.ReportPath, err)
		}
	}

	minTTC, meanJerk, hardBrakes := analyzeTimeline(report.Timeline)
	return Scorecard{
		ScenarioName:        report.Metrics.ScenarioName,
		VehiclesCompleted:   report.Metrics.VehiclesCompleted,
		ThroughputPer100:    report.Metrics.ThroughputPer100Step,
		AverageDelay:        report.Metrics.AverageWaitPerTrip,
		PotentialCollisions: report.Metrics.PotentialCollisions,
		MinTTCSteps:         minTTC,
		MeanAbsJerk:         meanJerk,
		HardBrakes:          hardBrakes,
	}, nil
}

func evaluate(spec Spec, baseline Scorecard, candidate Scorecard) Result {
	checks := []CheckResult{
		{
			Name:      "throughput",
			Rule:      fmt.Sprintf("candidate throughput >= baseline * %.3f", spec.Thresholds.MinThroughputRatio),
			Baseline:  baseline.ThroughputPer100,
			Candidate: candidate.ThroughputPer100,
			Passed:    candidate.ThroughputPer100 >= baseline.ThroughputPer100*spec.Thresholds.MinThroughputRatio,
		},
		{
			Name:      "average_delay",
			Rule:      fmt.Sprintf("candidate delay <= baseline + %.3f", spec.Thresholds.MaxDelayIncrease),
			Baseline:  baseline.AverageDelay,
			Candidate: candidate.AverageDelay,
			Passed:    candidate.AverageDelay <= baseline.AverageDelay+spec.Thresholds.MaxDelayIncrease,
		},
		{
			Name:      "potential_collisions",
			Rule:      fmt.Sprintf("candidate collisions <= baseline + %d", spec.Thresholds.MaxCollisionIncrease),
			Baseline:  float64(baseline.PotentialCollisions),
			Candidate: float64(candidate.PotentialCollisions),
			Passed:    candidate.PotentialCollisions <= baseline.PotentialCollisions+spec.Thresholds.MaxCollisionIncrease,
		},
		{
			Name:      "mean_abs_jerk",
			Rule:      fmt.Sprintf("candidate mean abs jerk <= baseline + %.3f", spec.Thresholds.MaxJerkIncrease),
			Baseline:  baseline.MeanAbsJerk,
			Candidate: candidate.MeanAbsJerk,
			Passed:    candidate.MeanAbsJerk <= baseline.MeanAbsJerk+spec.Thresholds.MaxJerkIncrease,
		},
		{
			Name:      "min_ttc",
			Rule:      fmt.Sprintf("candidate min TTC >= baseline - %.3f", spec.Thresholds.MaxMinTTCDrop),
			Baseline:  baseline.MinTTCSteps,
			Candidate: candidate.MinTTCSteps,
			Passed:    candidate.MinTTCSteps >= baseline.MinTTCSteps-spec.Thresholds.MaxMinTTCDrop,
		},
	}

	passed := true
	for _, check := range checks {
		if !check.Passed {
			passed = false
			break
		}
	}

	return Result{
		Name:      spec.Name,
		Generated: time.Now().UTC(),
		Baseline:  baseline,
		Candidate: candidate,
		Checks:    checks,
		Passed:    passed,
	}
}

func writeResult(path string, result Result) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create benchmark report dir: %w", err)
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal benchmark report: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write benchmark report: %w", err)
	}
	return nil
}

func analyzeTimeline(timeline []sim.StepSnapshot) (float64, float64, int) {
	if len(timeline) == 0 {
		return noClosingTTC, 0, 0
	}

	type state struct {
		x, y     int
		speed    float64
		accel    float64
		hasPos   bool
		hasAccel bool
	}

	states := map[int]state{}
	minTTC := noClosingTTC
	jerkSum := 0.0
	jerkSamples := 0
	hardBrakes := 0

	for _, snap := range timeline {
		currentSpeeds := map[int]float64{}
		for _, v := range snap.Vehicles {
			prev := states[v.ID]
			speed := 0.0
			if prev.hasPos {
				speed = float64(abs(v.X-prev.x) + abs(v.Y-prev.y))
			}
			currentSpeeds[v.ID] = speed

			accel := speed - prev.speed
			if prev.hasPos && accel <= -1 {
				hardBrakes++
			}
			if prev.hasAccel {
				jerkSum += math.Abs(accel - prev.accel)
				jerkSamples++
			}
			states[v.ID] = state{
				x:        v.X,
				y:        v.Y,
				speed:    speed,
				accel:    accel,
				hasPos:   true,
				hasAccel: true,
			}
		}

		if ttc := minTTCStep(snap.Vehicles, currentSpeeds); ttc < minTTC {
			minTTC = ttc
		}
	}

	meanJerk := 0.0
	if jerkSamples > 0 {
		meanJerk = jerkSum / float64(jerkSamples)
	}
	return minTTC, meanJerk, hardBrakes
}

func minTTCStep(vehicles []sim.Vehicle, speeds map[int]float64) float64 {
	type laneKey struct {
		dir sim.Direction
		key int
	}
	lanes := map[laneKey][]sim.Vehicle{}
	for _, v := range vehicles {
		key := v.Y
		if v.Direction == sim.Up || v.Direction == sim.Down {
			key = v.X
		}
		lanes[laneKey{dir: v.Direction, key: key}] = append(lanes[laneKey{dir: v.Direction, key: key}], v)
	}

	minTTC := noClosingTTC
	minHeadwayProxy := noClosingTTC
	for k, laneVehicles := range lanes {
		if len(laneVehicles) < 2 {
			continue
		}

		sort.Slice(laneVehicles, func(i, j int) bool {
			switch k.dir {
			case sim.Up:
				return laneVehicles[i].Y < laneVehicles[j].Y
			case sim.Down:
				return laneVehicles[i].Y > laneVehicles[j].Y
			case sim.Left:
				return laneVehicles[i].X < laneVehicles[j].X
			case sim.Right:
				return laneVehicles[i].X > laneVehicles[j].X
			default:
				return laneVehicles[i].ID < laneVehicles[j].ID
			}
		})

		for i := 1; i < len(laneVehicles); i++ {
			leader := laneVehicles[i-1]
			follower := laneVehicles[i]

			gap := 0
			switch k.dir {
			case sim.Up:
				gap = follower.Y - leader.Y - 1
			case sim.Down:
				gap = leader.Y - follower.Y - 1
			case sim.Left:
				gap = follower.X - leader.X - 1
			case sim.Right:
				gap = leader.X - follower.X - 1
			}
			if gap < 0 {
				gap = 0
			}
			headwayProxy := float64(gap + 1)
			if headwayProxy < minHeadwayProxy {
				minHeadwayProxy = headwayProxy
			}

			relativeSpeed := speeds[follower.ID] - speeds[leader.ID]
			if relativeSpeed <= 0 {
				continue
			}

			ttc := float64(gap+1) / relativeSpeed
			if ttc < minTTC {
				minTTC = ttc
			}
		}
	}
	if minTTC == noClosingTTC {
		return minHeadwayProxy
	}
	return minTTC
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
