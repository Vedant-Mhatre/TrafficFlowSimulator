# TrafficFlowSimulator

Go-based simulation and benchmarking tool for intersection traffic behavior.

## Start Here

Problem:
- simulation logic changes are easy to make but hard to evaluate reliably.
- ad-hoc runs do not clearly tell you if a change improved behavior or caused regressions.

Solution in this repo:
- run deterministic baseline vs candidate scenarios.
- compute a consistent scorecard (throughput, delay, collisions, TTC proxy, jerk, hard brakes).
- fail fast on regression thresholds with a non-zero exit code.

If you only run one command, run this:

```bash
go run ./cmd/trafficsim -benchmark configs/benchmark/intersection-regression.json
```

## Quick Commands

```bash
# run tests
go test ./...

# run deterministic regression benchmark
go run ./cmd/trafficsim -benchmark configs/benchmark/intersection-regression.json

# compare two configs quickly
go run ./cmd/trafficsim -compare configs/baseline.json,configs/improved.json

# interactive terminal render (ASCII grid)
go run ./cmd/trafficsim -config configs/baseline.json
```

Make shortcuts:

```bash
make test
make compare
make benchmark
make rush
```

## CLI Modes

- `-config <file>`: run one scenario and print metrics.
- `-compare a.json,b.json`: run multiple scenarios and print side-by-side summary.
- `-benchmark <spec.json>`: run deterministic baseline vs candidate plus pass/fail checks.

## What The Benchmark Reports

- `throughput_per_100_steps`
- `average_delay_steps`
- `potential_collisions`
- `min_ttc_steps` (discrete proxy)
- `mean_abs_jerk`
- `hard_brakes`

Checks are configured in `thresholds`. If any check fails, command exits non-zero.

## Example Benchmark Output

```text
Benchmark: intersection-rush-hour-regression
Case | Completed | Throughput/100 | Avg Delay | Collisions | Min TTC | Mean Abs Jerk | Hard Brakes
baseline(...) | 61 | 50.83 | 7.61 | 0 | 1.00 | 0.264 | 50
candidate(...) | 61 | 50.83 | 6.10 | 0 | 1.00 | 0.264 | 44
Overall: PASS
```

## Project Layout

- `cmd/trafficsim/main.go`: CLI.
- `internal/sim/*`: simulation engine, config, rendering, reports.
- `internal/benchmark/*`: deterministic benchmark runner and checks.
- `configs/baseline.json`: baseline scenario.
- `configs/improved.json`: alternate scenario.
- `configs/rush-hour.json`: profile-based demand scenario.
- `configs/rush-hour.csv`: demand profile.
- `configs/benchmark/intersection-regression.json`: benchmark spec.
- `configs/benchmark/intersection-baseline.json`: baseline benchmark scenario.
- `configs/benchmark/intersection-candidate.json`: candidate benchmark scenario.

## Visualization

- Terminal visualization already exists via `-config ...` without `-no-render`.
- It is currently ASCII-grid based and intentionally simple.
- Richer visualization (ncurses or web replay) should be a separate PR to keep benchmark logic changes isolated.

## Benchmark Spec Reference

Minimal benchmark spec:

```json
{
  "name": "intersection-rush-hour-regression",
  "baseline_config": "intersection-baseline.json",
  "candidate_config": "intersection-candidate.json",
  "thresholds": {
    "max_collision_increase": 0,
    "max_delay_increase": 0.2,
    "min_throughput_ratio": 0.95,
    "max_jerk_increase": 0.15,
    "max_min_ttc_drop": 0.5
  },
  "report_path": "../../reports/benchmark-intersection-scorecard.json"
}
```

Field summary:

- `baseline_config`, `candidate_config`: scenario config paths.
- `max_collision_increase`: allowed collision increase vs baseline.
- `max_delay_increase`: allowed average delay increase.
- `min_throughput_ratio`: required candidate/baseline throughput ratio.
- `max_jerk_increase`: allowed jerk increase.
- `max_min_ttc_drop`: allowed TTC proxy drop.
- `report_path`: optional JSON output path.

TTC note:
- TTC is a discrete proxy in this grid model, not continuous physics TTC.

## Scenario Config Notes

- Lanes: `up`, `down`, `left`, `right`.
- `step_interval: 0` disables periodic spawning.
- `max_vehicles: 0` means uncapped.
- `report_path` and `profile_csv` relative paths are resolved from config file directory.
- `up`/`down` must spawn on center vertical road.
- `left`/`right` must spawn on center horizontal road.

## Limits

- Single-intersection road topology.
- Discrete grid movement, not continuous vehicle dynamics.
- Conflict/TTC are proxy metrics.

## Demo

![Traffic Simulation Demo](./simulation.gif)
