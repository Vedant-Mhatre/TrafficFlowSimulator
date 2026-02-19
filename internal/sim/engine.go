package sim

import (
	"fmt"
	"sort"
	"time"
)

type Vehicle struct {
	ID          int       `json:"id"`
	X           int       `json:"x"`
	Y           int       `json:"y"`
	Direction   Direction `json:"direction"`
	SpawnStep   int       `json:"spawn_step"`
	WaitSteps   int       `json:"wait_steps"`
	MovedSteps  int       `json:"moved_steps"`
	BlockedStep int       `json:"blocked_step"`
}

type TrafficLight struct {
	VerticalGreen bool `json:"vertical_green"`
	Timer         int  `json:"timer"`
}

type LaneState struct {
	Direction        Direction `json:"direction"`
	EntryX           int       `json:"entry_x"`
	EntryY           int       `json:"entry_y"`
	Interval         int       `json:"interval"`
	MaxVehicles      int       `json:"max_vehicles"`
	Spawned          int       `json:"spawned"`
	Queued           int       `json:"queued"`
	MaxQueueObserved int       `json:"max_queue_observed"`
	Profile          DemandProfile
}

type Metrics struct {
	ScenarioName         string                 `json:"scenario_name"`
	Steps                int                    `json:"steps"`
	VehiclesSpawned      int                    `json:"vehicles_spawned"`
	VehiclesCompleted    int                    `json:"vehicles_completed"`
	ActiveVehicles       int                    `json:"active_vehicles"`
	BlockedBySignal      int                    `json:"blocked_by_signal"`
	BlockedByTraffic     int                    `json:"blocked_by_traffic"`
	PotentialCollisions  int                    `json:"potential_collisions"`
	TotalDistance        int                    `json:"total_distance"`
	AverageNetworkSpeed  float64                `json:"average_network_speed"`
	AverageWaitPerTrip   float64                `json:"average_wait_per_trip"`
	AverageTripDuration  float64                `json:"average_trip_duration"`
	ThroughputPer100Step float64                `json:"throughput_per_100_steps"`
	MaxQueueOverall      int                    `json:"max_queue_overall"`
	DirectionStats       map[Direction]DirStats `json:"direction_stats"`
}

type DirStats struct {
	Spawned         int     `json:"spawned"`
	Completed       int     `json:"completed"`
	AverageWait     float64 `json:"average_wait"`
	AverageDuration float64 `json:"average_duration"`
	MaxQueue        int     `json:"max_queue"`
}

type StepSnapshot struct {
	Step       int       `json:"step"`
	LightGreen bool      `json:"light_green_vertical"`
	Vehicles   []Vehicle `json:"vehicles"`
}

type Report struct {
	ConfigName string         `json:"config_name"`
	Generated  time.Time      `json:"generated"`
	Metrics    Metrics        `json:"metrics"`
	Timeline   []StepSnapshot `json:"timeline,omitempty"`
}

type Engine struct {
	cfg              Config
	vehicles         []Vehicle
	light            TrafficLight
	laneStates       map[Direction]*LaneState
	nextVehicleID    int
	intersectionX    int
	intersectionY    int
	totalVehicleStep int
	totalWaitEnded   int
	totalTripEnded   int
	dirWaitEnded     map[Direction]int
	dirTripEnded     map[Direction]int
	dirDone          map[Direction]int
	dirSpawn         map[Direction]int
	blockedSignal    int
	blockedTraffic   int
	potentialCrash   int
	totalDistance    int
	maxQueueOverall  int
	timeline         []StepSnapshot
}

func NewEngine(cfg Config) (*Engine, error) {
	laneStates := make(map[Direction]*LaneState, len(cfg.Spawn.Lanes))
	for dir, lane := range cfg.Spawn.Lanes {
		state := &LaneState{
			Direction:   dir,
			EntryX:      lane.EntryX,
			EntryY:      lane.EntryY,
			Interval:    lane.StepInterval,
			MaxVehicles: lane.MaxVehicles,
			Profile:     DemandProfile{},
		}
		if lane.ProfileCSV != "" {
			column := lane.ProfileColumn
			if column == "" {
				column = string(dir)
			}
			profile, err := LoadDemandProfile(lane.ProfileCSV, column)
			if err != nil {
				return nil, fmt.Errorf("load demand profile for lane %q: %w", dir, err)
			}
			state.Profile = profile
		}
		laneStates[dir] = state
	}

	return &Engine{
		cfg:           cfg,
		light:         TrafficLight{VerticalGreen: true},
		laneStates:    laneStates,
		intersectionX: cfg.Grid.Width / 2,
		intersectionY: cfg.Grid.Height / 2,
		dirWaitEnded:  map[Direction]int{},
		dirTripEnded:  map[Direction]int{},
		dirDone:       map[Direction]int{},
		dirSpawn:      map[Direction]int{},
	}, nil
}

func (e *Engine) Run(captureTimeline bool, renderOverride *bool) Report {
	shouldRender := e.cfg.Render.Enabled
	if renderOverride != nil {
		shouldRender = *renderOverride
	}

	if captureTimeline {
		e.timeline = make([]StepSnapshot, 0, e.cfg.Steps)
	}

	for step := 0; step < e.cfg.Steps; step++ {
		e.spawnVehicles(step)
		e.moveVehicles(step)
		e.updateLight()

		if captureTimeline {
			e.timeline = append(e.timeline, e.snapshot(step))
		}

		if shouldRender {
			RenderGrid(e.cfg, e.vehicles, e.light, step)
			if e.cfg.Render.DelayMS > 0 {
				time.Sleep(time.Duration(e.cfg.Render.DelayMS) * time.Millisecond)
			}
		}
	}

	return Report{
		ConfigName: e.cfg.Name,
		Generated:  time.Now().UTC(),
		Metrics:    e.metrics(),
		Timeline:   e.timeline,
	}
}

func (e *Engine) snapshot(step int) StepSnapshot {
	copyVehicles := make([]Vehicle, len(e.vehicles))
	copy(copyVehicles, e.vehicles)
	return StepSnapshot{
		Step:       step + 1,
		LightGreen: e.light.VerticalGreen,
		Vehicles:   copyVehicles,
	}
}

func (e *Engine) spawnVehicles(step int) {
	directions := make([]Direction, 0, len(e.laneStates))
	for dir := range e.laneStates {
		directions = append(directions, dir)
	}
	sort.Slice(directions, func(i, j int) bool { return directions[i] < directions[j] })

	for _, dir := range directions {
		lane := e.laneStates[dir]
		newArrivals := e.arrivalsForStep(lane, step)
		if newArrivals == 0 {
			continue
		}
		lane.Queued += newArrivals
		if lane.Queued > lane.MaxQueueObserved {
			lane.MaxQueueObserved = lane.Queued
		}
		if lane.Queued > e.maxQueueOverall {
			e.maxQueueOverall = lane.Queued
		}
	}

	for _, dir := range directions {
		lane := e.laneStates[dir]
		for lane.Queued > 0 {
			if lane.MaxVehicles > 0 && lane.Spawned >= lane.MaxVehicles {
				lane.Queued = 0
				break
			}
			if e.occupied(lane.EntryX, lane.EntryY) {
				break
			}
			e.nextVehicleID++
			e.vehicles = append(e.vehicles, Vehicle{
				ID:        e.nextVehicleID,
				X:         lane.EntryX,
				Y:         lane.EntryY,
				Direction: dir,
				SpawnStep: step + 1,
			})
			lane.Queued--
			lane.Spawned++
			e.dirSpawn[dir]++
		}
	}
}

func (e *Engine) arrivalsForStep(lane *LaneState, step int) int {
	if len(lane.Profile) > 0 {
		return lane.Profile[step+1]
	}
	if lane.Interval <= 0 {
		return 0
	}
	if (step+1)%lane.Interval == 0 {
		return 1
	}
	return 0
}

func (e *Engine) occupied(x, y int) bool {
	for i := range e.vehicles {
		if e.vehicles[i].X == x && e.vehicles[i].Y == y {
			return true
		}
	}
	return false
}

func (e *Engine) moveVehicles(step int) {
	type movePlan struct {
		canMove   bool
		exitsGrid bool
		blockedBy string
		nextX     int
		nextY     int
	}
	type cell struct {
		x int
		y int
	}

	plans := make([]movePlan, len(e.vehicles))
	occupancy := map[cell]bool{}
	currentPos := make([]cell, len(e.vehicles))
	positionToVehicle := map[cell]int{}
	for i := range e.vehicles {
		pos := cell{x: e.vehicles[i].X, y: e.vehicles[i].Y}
		currentPos[i] = pos
		occupancy[pos] = true
		positionToVehicle[pos] = i
	}

	targets := map[cell][]int{}
	for i := range e.vehicles {
		v := e.vehicles[i]
		nextX, nextY := nextCell(v)
		plan := movePlan{nextX: nextX, nextY: nextY}

		if nextX < 0 || nextX >= e.cfg.Grid.Width || nextY < 0 || nextY >= e.cfg.Grid.Height {
			plan.canMove = true
			plan.exitsGrid = true
			plans[i] = plan
			continue
		}

		if nextX == e.intersectionX && nextY == e.intersectionY {
			if (v.Direction == Up || v.Direction == Down) && !e.light.VerticalGreen {
				plan.blockedBy = "signal"
				plans[i] = plan
				continue
			}
			if (v.Direction == Left || v.Direction == Right) && e.light.VerticalGreen {
				plan.blockedBy = "signal"
				plans[i] = plan
				continue
			}
		}

		targets[cell{x: nextX, y: nextY}] = append(targets[cell{x: nextX, y: nextY}], i)
		plan.canMove = true
		plans[i] = plan
	}

	for _, indices := range targets {
		if len(indices) <= 1 {
			continue
		}
		e.potentialCrash++
		for _, idx := range indices {
			plans[idx].canMove = false
			plans[idx].blockedBy = "traffic"
		}
	}

	// Resolve dependencies against vehicles occupying target cells. This allows
	// platoons to move forward in the same step when the lead vehicle vacates.
	changed := true
	for changed {
		changed = false
		for i := range e.vehicles {
			plan := plans[i]
			if !plan.canMove || plan.exitsGrid {
				continue
			}

			target := cell{x: plan.nextX, y: plan.nextY}
			occIdx, occupied := positionToVehicle[target]
			if !occupied {
				continue
			}
			if occIdx == i {
				continue
			}

			occPlan := plans[occIdx]
			occupantLeaves := occPlan.canMove && (occPlan.exitsGrid || occPlan.nextX != target.x || occPlan.nextY != target.y)
			swapsPositions := occPlan.canMove && !occPlan.exitsGrid &&
				occPlan.nextX == currentPos[i].x && occPlan.nextY == currentPos[i].y

			if occupantLeaves && !swapsPositions {
				continue
			}

			plans[i].canMove = false
			plans[i].blockedBy = "traffic"
			changed = true
		}
	}

	nextVehicles := make([]Vehicle, 0, len(e.vehicles))
	for i := range e.vehicles {
		v := e.vehicles[i]
		e.totalVehicleStep++
		plan := plans[i]

		if plan.canMove {
			v.MovedSteps++
			e.totalDistance++
			if plan.exitsGrid {
				tripDuration := (step + 1) - v.SpawnStep + 1
				e.totalTripEnded += tripDuration
				e.totalWaitEnded += v.WaitSteps
				e.dirWaitEnded[v.Direction] += v.WaitSteps
				e.dirTripEnded[v.Direction] += tripDuration
				e.dirDone[v.Direction]++
				continue
			}
			v.X, v.Y = plan.nextX, plan.nextY
		} else {
			v.WaitSteps++
			if plan.blockedBy == "signal" {
				e.blockedSignal++
			}
			if plan.blockedBy == "traffic" {
				e.blockedTraffic++
			}
		}

		nextVehicles = append(nextVehicles, v)
	}

	e.vehicles = nextVehicles
}

func nextCell(v Vehicle) (int, int) {
	nextX, nextY := v.X, v.Y
	switch v.Direction {
	case Up:
		nextY--
	case Down:
		nextY++
	case Left:
		nextX--
	case Right:
		nextX++
	}
	return nextX, nextY
}

func (e *Engine) updateLight() {
	e.light.Timer++
	cycle := e.cfg.Signal.VerticalGreenSteps + e.cfg.Signal.HorizontalGreenSteps
	if cycle <= 0 {
		return
	}
	stepInCycle := (e.light.Timer - 1) % cycle
	if stepInCycle < e.cfg.Signal.VerticalGreenSteps {
		e.light.VerticalGreen = true
	} else {
		e.light.VerticalGreen = false
	}
}

func (e *Engine) metrics() Metrics {
	completed := 0
	for _, dir := range e.dirDone {
		completed += dir
	}

	m := Metrics{
		ScenarioName:        e.cfg.Name,
		Steps:               e.cfg.Steps,
		VehiclesSpawned:     len(e.vehicles) + completed,
		VehiclesCompleted:   completed,
		ActiveVehicles:      len(e.vehicles),
		BlockedBySignal:     e.blockedSignal,
		BlockedByTraffic:    e.blockedTraffic,
		PotentialCollisions: e.potentialCrash,
		TotalDistance:       e.totalDistance,
		MaxQueueOverall:     e.maxQueueOverall,
		DirectionStats:      map[Direction]DirStats{},
	}

	if e.totalVehicleStep > 0 {
		m.AverageNetworkSpeed = float64(e.totalDistance) / float64(e.totalVehicleStep)
	}
	if completed > 0 {
		m.AverageWaitPerTrip = float64(e.totalWaitEnded) / float64(completed)
		m.AverageTripDuration = float64(e.totalTripEnded) / float64(completed)
	}
	if e.cfg.Steps > 0 {
		m.ThroughputPer100Step = float64(completed) / float64(e.cfg.Steps) * 100
	}

	for dir, lane := range e.laneStates {
		stat := DirStats{
			Spawned:   e.dirSpawn[dir],
			Completed: e.dirDone[dir],
			MaxQueue:  lane.MaxQueueObserved,
		}
		if e.dirDone[dir] > 0 {
			stat.AverageWait = float64(e.dirWaitEnded[dir]) / float64(e.dirDone[dir])
			stat.AverageDuration = float64(e.dirTripEnded[dir]) / float64(e.dirDone[dir])
		}
		m.DirectionStats[dir] = stat
	}

	return m
}
