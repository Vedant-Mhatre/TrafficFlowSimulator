package sim

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Direction string

const (
	Up    Direction = "up"
	Down  Direction = "down"
	Left  Direction = "left"
	Right Direction = "right"
)

type Config struct {
	Name       string       `json:"name"`
	Steps      int          `json:"steps"`
	Grid       GridConfig   `json:"grid"`
	Signal     SignalConfig `json:"signal"`
	Spawn      SpawnConfig  `json:"spawn"`
	Render     RenderConfig `json:"render"`
	ReportPath string       `json:"report_path"`
}

type GridConfig struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type SignalConfig struct {
	VerticalGreenSteps   int `json:"vertical_green_steps"`
	HorizontalGreenSteps int `json:"horizontal_green_steps"`
}

type SpawnConfig struct {
	Lanes map[Direction]LaneSpawnConfig `json:"lanes"`
}

type LaneSpawnConfig struct {
	EntryX        int    `json:"entry_x"`
	EntryY        int    `json:"entry_y"`
	StepInterval  int    `json:"step_interval"`
	MaxVehicles   int    `json:"max_vehicles"`
	ProfileCSV    string `json:"profile_csv"`
	ProfileColumn string `json:"profile_column"`
}

type RenderConfig struct {
	Enabled bool `json:"enabled"`
	DelayMS int  `json:"delay_ms"`
}

type DemandProfile map[int]int

func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}

	applyDefaults(&cfg)
	resolveConfigPaths(&cfg, filepath.Dir(path))
	if err := validateConfig(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.Name == "" {
		cfg.Name = "default"
	}
	if cfg.Steps <= 0 {
		cfg.Steps = 100
	}
	if cfg.Grid.Width <= 0 {
		cfg.Grid.Width = 20
	}
	if cfg.Grid.Height <= 0 {
		cfg.Grid.Height = 10
	}
	if cfg.Signal.VerticalGreenSteps <= 0 {
		cfg.Signal.VerticalGreenSteps = 5
	}
	if cfg.Signal.HorizontalGreenSteps <= 0 {
		cfg.Signal.HorizontalGreenSteps = 5
	}
	if cfg.Spawn.Lanes == nil {
		cfg.Spawn.Lanes = map[Direction]LaneSpawnConfig{
			Up: {
				EntryX:       cfg.Grid.Width / 2,
				EntryY:       cfg.Grid.Height - 1,
				StepInterval: 3,
			},
			Right: {
				EntryX:       0,
				EntryY:       cfg.Grid.Height / 2,
				StepInterval: 4,
			},
		}
	}
	if cfg.Render.DelayMS < 0 {
		cfg.Render.DelayMS = 0
	}
}

func validateConfig(cfg Config) error {
	if cfg.Grid.Width < 3 || cfg.Grid.Height < 3 {
		return fmt.Errorf("grid must be at least 3x3")
	}
	if len(cfg.Spawn.Lanes) == 0 {
		return fmt.Errorf("spawn lanes cannot be empty")
	}

	intersectionX := cfg.Grid.Width / 2
	intersectionY := cfg.Grid.Height / 2

	for dir, lane := range cfg.Spawn.Lanes {
		if dir != Up && dir != Down && dir != Left && dir != Right {
			return fmt.Errorf("unsupported direction %q", dir)
		}
		if lane.EntryX < 0 || lane.EntryX >= cfg.Grid.Width || lane.EntryY < 0 || lane.EntryY >= cfg.Grid.Height {
			return fmt.Errorf("lane %q entry is outside grid", dir)
		}
		if lane.StepInterval < 0 {
			return fmt.Errorf("lane %q step_interval must be >= 0", dir)
		}
		if lane.MaxVehicles < 0 {
			return fmt.Errorf("lane %q max_vehicles must be >= 0", dir)
		}
		if (dir == Up || dir == Down) && lane.EntryX != intersectionX {
			return fmt.Errorf("lane %q entry_x must equal center road x=%d", dir, intersectionX)
		}
		if (dir == Left || dir == Right) && lane.EntryY != intersectionY {
			return fmt.Errorf("lane %q entry_y must equal center road y=%d", dir, intersectionY)
		}
	}
	return nil
}

func resolveConfigPaths(cfg *Config, baseDir string) {
	if baseDir == "" {
		return
	}

	if cfg.ReportPath != "" && !filepath.IsAbs(cfg.ReportPath) {
		cfg.ReportPath = filepath.Join(baseDir, cfg.ReportPath)
	}

	for dir, lane := range cfg.Spawn.Lanes {
		if lane.ProfileCSV != "" && !filepath.IsAbs(lane.ProfileCSV) {
			lane.ProfileCSV = filepath.Join(baseDir, lane.ProfileCSV)
			cfg.Spawn.Lanes[dir] = lane
		}
	}
}

func LoadDemandProfile(path string, column string) (DemandProfile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open profile: %w", err)
	}
	defer file.Close()

	r := csv.NewReader(file)
	rows, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read profile csv: %w", err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("profile csv is empty")
	}

	header := rows[0]
	stepIdx := -1
	valueIdx := -1
	for i, col := range header {
		col = strings.TrimSpace(strings.ToLower(col))
		if col == "step" {
			stepIdx = i
		}
		if col == strings.TrimSpace(strings.ToLower(column)) {
			valueIdx = i
		}
	}
	if stepIdx == -1 {
		return nil, fmt.Errorf("profile csv missing 'step' column")
	}
	if valueIdx == -1 {
		return nil, fmt.Errorf("profile csv missing column %q", column)
	}

	profile := DemandProfile{}
	for _, row := range rows[1:] {
		if stepIdx >= len(row) || valueIdx >= len(row) {
			continue
		}
		step, err := strconv.Atoi(strings.TrimSpace(row[stepIdx]))
		if err != nil {
			continue
		}
		count, err := strconv.Atoi(strings.TrimSpace(row[valueIdx]))
		if err != nil {
			continue
		}
		profile[step] = count
	}
	return profile, nil
}
