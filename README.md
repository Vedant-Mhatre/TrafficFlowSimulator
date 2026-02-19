# TrafficFlowSimulator

`TrafficFlowSimulator` is a Go-based traffic simulation CLI for experimenting with intersection control, demand patterns, and performance metrics.

The original C++ prototype is still available in `traffic_sim.cpp`.

## Demo

![Traffic Simulation Demo](./simulation.gif)

## Why This Is Useful

- Signal timing experiments: compare baseline vs tuned timing plans quickly.
- Quantifiable outcomes: throughput, waiting, congestion, and conflict indicators.
- Interview readiness: clear simulator architecture plus reproducible scenario runs.
- Extendable base: can be evolved toward richer AV/traffic simulation components.

## Features

- Single-intersection grid simulation (`20x10` default, configurable).
- Multi-direction traffic demand (`up`, `down`, `left`, `right`).
- Fixed-time traffic lights with independent vertical/horizontal green windows.
- Arrival models:
  - Periodic arrivals with `step_interval`.
  - Optional CSV demand profiles (`step`, per-lane count column).
- Metrics and reporting:
  - Vehicles spawned/completed, throughput per 100 steps.
  - Average wait time and average trip duration.
  - Queue lengths (per lane and global max).
  - Blocking by signal vs blocking by traffic.
  - Potential collision/conflict count.
- Compare mode to evaluate multiple configs in one command.
- JSON report output for downstream analysis.

## Project Layout

- `cmd/trafficsim/main.go`: CLI entrypoint.
- `internal/sim/config.go`: config schema and validation.
- `internal/sim/engine.go`: simulation engine and metrics.
- `internal/sim/render.go`: terminal visualization.
- `internal/sim/report.go`: JSON report writer.
- `configs/baseline.json`: baseline scenario.
- `configs/improved.json`: tuned signal-timing scenario.
- `configs/rush-hour.json`: profile-based demand scenario.
- `configs/rush-hour.csv`: sample CSV demand profile.

## Prerequisites

- Go `1.22+`

## Quick Start

Run baseline scenario with terminal visualization:

```bash
go run ./cmd/trafficsim -config configs/baseline.json
```

Run without visualization (faster for analysis):

```bash
go run ./cmd/trafficsim -config configs/baseline.json -no-render
```

Run and include timeline snapshots in report:

```bash
go run ./cmd/trafficsim -config configs/baseline.json -no-render -timeline
```

Compare multiple scenarios:

```bash
go run ./cmd/trafficsim -compare configs/baseline.json,configs/improved.json
```

Run the CSV profile scenario:

```bash
go run ./cmd/trafficsim -config configs/rush-hour.json -no-render
```

Build a local binary:

```bash
go build -o trafficsim ./cmd/trafficsim
./trafficsim -config configs/baseline.json -no-render
```

Use Make targets (optional):

```bash
make test
make compare
make rush
```

## Config Reference

Top-level fields:

- `name`: scenario name used in CLI/report output.
- `steps`: total simulation steps.
- `grid.width`, `grid.height`: grid dimensions.
- `signal.vertical_green_steps`, `signal.horizontal_green_steps`: fixed-time signal plan.
- `spawn.lanes`: map keyed by lane direction (`up`, `down`, `left`, `right`).
- `render.enabled`, `render.delay_ms`: terminal rendering controls.
- `report_path`: JSON report path. Relative paths are resolved relative to the config file directory.

Lane fields:

- `entry_x`, `entry_y`: spawn coordinates for the lane.
- `step_interval`: spawn every N steps (`0` disables periodic spawning).
- `max_vehicles`: optional cap for total spawned vehicles (`0` means unlimited).
- `profile_csv`: optional CSV demand profile file.
- `profile_column`: CSV column to read counts from (defaults to lane name).

Lane validation rules:

- `up`/`down` lanes must use center vertical road (`entry_x == grid.width/2`).
- `left`/`right` lanes must use center horizontal road (`entry_y == grid.height/2`).

## CSV Demand Profile Format

Required:

- `step` column.
- A numeric lane column (for example `up`, `right`, or custom name referenced by `profile_column`).

Example:

```csv
step,up,right
1,2,0
2,3,1
3,5,2
```

## Example Output

```text
Scenario: baseline-fixed-time
Spawned: 98 | Completed: 83 | Active: 15
Avg speed: 0.532 | Avg wait: 10.61 | Avg trip: 24.47
Throughput/100 steps: 46.11 | Max queue: 5 | Potential collisions: 0
Blocked by signal: 154 | Blocked by traffic: 915
```

## Testing And Verification

Run unit tests:

```bash
go test ./...
```

Run additional static checks:

```bash
go vet ./...
```

## Known Limits

- Roads are represented as a single cross intersection (not a full road network graph).
- Vehicle motion is discrete grid-step movement (not continuous dynamics).
- Conflict counting is a simple indicator, not a full collision-physics model.

## Legacy C++ Prototype

If you want the original version:

```bash
g++ -std=c++11 -o traffic_sim traffic_sim.cpp
./traffic_sim
```

## Minimal Config Example

```json
{
  "name": "baseline-fixed-time",
  "steps": 180,
  "grid": { "width": 20, "height": 10 },
  "signal": { "vertical_green_steps": 5, "horizontal_green_steps": 5 },
  "spawn": {
    "lanes": {
      "up": { "entry_x": 10, "entry_y": 9, "step_interval": 3, "max_vehicles": 0 },
      "right": { "entry_x": 0, "entry_y": 5, "step_interval": 4, "max_vehicles": 0 }
    }
  },
  "render": { "enabled": false, "delay_ms": 0 },
  "report_path": "reports/baseline-report.json"
}
```

Notes:
- `max_vehicles: 0` means no hard cap.
- `step_interval: 0` disables periodic spawning for that lane.
- CSV demand profiles are supported per lane with `profile_csv` and `profile_column`.
