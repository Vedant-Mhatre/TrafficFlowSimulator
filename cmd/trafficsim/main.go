package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/Vedant-Mhatre/TrafficFlowSimulator/internal/benchmark"
	"github.com/Vedant-Mhatre/TrafficFlowSimulator/internal/sim"
)

func main() {
	configPath := flag.String("config", "configs/baseline.json", "Path to a simulation config JSON")
	compare := flag.String("compare", "", "Comma-separated config paths to run and compare")
	benchmarkPath := flag.String("benchmark", "", "Path to deterministic benchmark spec JSON")
	noRender := flag.Bool("no-render", false, "Disable terminal rendering")
	captureTimeline := flag.Bool("timeline", false, "Include per-step timeline in report JSON")
	out := flag.String("out", "", "Optional report output path override for single config mode")
	flag.Parse()

	if *benchmarkPath != "" && *compare != "" {
		exitErr(errors.New("benchmark mode cannot be used with compare mode"))
	}
	if *benchmarkPath != "" {
		if err := runBenchmark(*benchmarkPath); err != nil {
			exitErr(err)
		}
		return
	}

	if *compare != "" {
		paths := splitAndTrim(*compare)
		if len(paths) < 2 {
			exitErr(errors.New("compare mode requires at least two config paths"))
		}
		if err := runCompare(paths); err != nil {
			exitErr(err)
		}
		return
	}

	cfg, err := sim.LoadConfig(*configPath)
	if err != nil {
		exitErr(err)
	}

	engine, err := sim.NewEngine(cfg)
	if err != nil {
		exitErr(err)
	}

	var override *bool
	if *noRender {
		render := false
		override = &render
	}
	report := engine.Run(*captureTimeline, override)
	printReport(report)

	reportPath := cfg.ReportPath
	if *out != "" {
		reportPath = *out
	}
	if reportPath != "" {
		if err := sim.WriteReport(reportPath, report); err != nil {
			exitErr(err)
		}
		fmt.Printf("\nReport written to %s\n", reportPath)
	}
}

func runBenchmark(path string) error {
	spec, err := benchmark.LoadSpec(path)
	if err != nil {
		return err
	}

	result, err := benchmark.Run(spec)
	if err != nil {
		return err
	}
	printBenchmark(result)

	if spec.ReportPath != "" {
		fmt.Printf("\nBenchmark report written to %s\n", spec.ReportPath)
	}
	if !result.Passed {
		return errors.New("benchmark failed regression checks")
	}
	return nil
}

func runCompare(paths []string) error {
	reports := make([]sim.Report, 0, len(paths))
	for _, path := range paths {
		cfg, err := sim.LoadConfig(path)
		if err != nil {
			return fmt.Errorf("load %q: %w", path, err)
		}
		engine, err := sim.NewEngine(cfg)
		if err != nil {
			return fmt.Errorf("build engine for %q: %w", path, err)
		}
		render := false
		report := engine.Run(false, &render)
		reports = append(reports, report)
		if cfg.ReportPath != "" {
			if err := sim.WriteReport(cfg.ReportPath, report); err != nil {
				return fmt.Errorf("write report %q: %w", cfg.ReportPath, err)
			}
		}
	}
	printComparison(reports)
	return nil
}

func printBenchmark(result benchmark.Result) {
	fmt.Printf("Benchmark: %s\n", result.Name)
	fmt.Println("Scorecard:")
	fmt.Println("Case | Completed | Throughput/100 | Avg Delay | Collisions | Min TTC | Mean Abs Jerk | Hard Brakes")
	fmt.Printf("baseline(%s) | %d | %.2f | %.2f | %d | %.2f | %.3f | %d\n",
		result.Baseline.ScenarioName,
		result.Baseline.VehiclesCompleted,
		result.Baseline.ThroughputPer100,
		result.Baseline.AverageDelay,
		result.Baseline.PotentialCollisions,
		result.Baseline.MinTTCSteps,
		result.Baseline.MeanAbsJerk,
		result.Baseline.HardBrakes,
	)
	fmt.Printf("candidate(%s) | %d | %.2f | %.2f | %d | %.2f | %.3f | %d\n",
		result.Candidate.ScenarioName,
		result.Candidate.VehiclesCompleted,
		result.Candidate.ThroughputPer100,
		result.Candidate.AverageDelay,
		result.Candidate.PotentialCollisions,
		result.Candidate.MinTTCSteps,
		result.Candidate.MeanAbsJerk,
		result.Candidate.HardBrakes,
	)

	fmt.Println("\nChecks:")
	for _, check := range result.Checks {
		status := "PASS"
		if !check.Passed {
			status = "FAIL"
		}
		fmt.Printf("- %s: %s (baseline=%.3f candidate=%.3f) -> %s\n",
			check.Name, check.Rule, check.Baseline, check.Candidate, status)
	}
	final := "PASS"
	if !result.Passed {
		final = "FAIL"
	}
	fmt.Printf("\nOverall: %s\n", final)
}

func printReport(report sim.Report) {
	m := report.Metrics
	fmt.Printf("Scenario: %s\n", m.ScenarioName)
	fmt.Printf("Spawned: %d | Completed: %d | Active: %d\n", m.VehiclesSpawned, m.VehiclesCompleted, m.ActiveVehicles)
	fmt.Printf("Avg speed: %.3f | Avg wait: %.2f | Avg trip: %.2f\n", m.AverageNetworkSpeed, m.AverageWaitPerTrip, m.AverageTripDuration)
	fmt.Printf("Throughput/100 steps: %.2f | Max queue: %d | Potential collisions: %d\n", m.ThroughputPer100Step, m.MaxQueueOverall, m.PotentialCollisions)
	fmt.Printf("Blocked by signal: %d | Blocked by traffic: %d\n", m.BlockedBySignal, m.BlockedByTraffic)

	dirs := make([]sim.Direction, 0, len(m.DirectionStats))
	for dir := range m.DirectionStats {
		dirs = append(dirs, dir)
	}
	sort.Slice(dirs, func(i, j int) bool { return dirs[i] < dirs[j] })
	for _, dir := range dirs {
		s := m.DirectionStats[dir]
		fmt.Printf("  %s -> spawned=%d completed=%d avg_wait=%.2f avg_trip=%.2f max_queue=%d\n",
			dir, s.Spawned, s.Completed, s.AverageWait, s.AverageDuration, s.MaxQueue)
	}
}

func printComparison(reports []sim.Report) {
	fmt.Println("Comparison:")
	fmt.Println("Scenario | Completed | Throughput/100 | Avg Wait | Avg Trip | Collisions")
	for _, report := range reports {
		m := report.Metrics
		fmt.Printf("%s | %d | %.2f | %.2f | %.2f | %d\n",
			m.ScenarioName,
			m.VehiclesCompleted,
			m.ThroughputPer100Step,
			m.AverageWaitPerTrip,
			m.AverageTripDuration,
			m.PotentialCollisions,
		)
	}
}

func splitAndTrim(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func exitErr(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}
