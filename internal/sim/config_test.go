package sim

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfigResolvesRelativePaths(t *testing.T) {
	tempDir := t.TempDir()
	profilesDir := filepath.Join(tempDir, "profiles")
	if err := os.MkdirAll(profilesDir, 0o755); err != nil {
		t.Fatalf("mkdir profiles: %v", err)
	}

	csvPath := filepath.Join(profilesDir, "morning.csv")
	if err := os.WriteFile(csvPath, []byte("step,up\n1,2\n2,3\n"), 0o644); err != nil {
		t.Fatalf("write csv: %v", err)
	}

	configPath := filepath.Join(tempDir, "baseline.json")
	content := `{
		"name": "path-resolution",
		"steps": 10,
		"grid": { "width": 20, "height": 10 },
		"signal": { "vertical_green_steps": 4, "horizontal_green_steps": 4 },
		"spawn": {
			"lanes": {
				"up": {
					"entry_x": 10,
					"entry_y": 9,
					"step_interval": 0,
					"profile_csv": "profiles/morning.csv",
					"profile_column": "up"
				}
			}
		},
		"report_path": "reports/local.json"
	}`
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	gotCSV := cfg.Spawn.Lanes[Up].ProfileCSV
	wantCSV := filepath.Join(tempDir, "profiles", "morning.csv")
	if gotCSV != wantCSV {
		t.Fatalf("profile csv path = %q, want %q", gotCSV, wantCSV)
	}

	wantReport := filepath.Join(tempDir, "reports", "local.json")
	if cfg.ReportPath != wantReport {
		t.Fatalf("report path = %q, want %q", cfg.ReportPath, wantReport)
	}
}

func TestLoadConfigRejectsLaneOutsideRoadAxis(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid.json")
	content := `{
		"steps": 10,
		"grid": { "width": 20, "height": 10 },
		"signal": { "vertical_green_steps": 4, "horizontal_green_steps": 4 },
		"spawn": {
			"lanes": {
				"up": { "entry_x": 9, "entry_y": 9, "step_interval": 2 }
			}
		}
	}`
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if !strings.Contains(err.Error(), "entry_x must equal center road x") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadDemandProfileParsesExpectedColumn(t *testing.T) {
	tempDir := t.TempDir()
	file := filepath.Join(tempDir, "profile.csv")
	if err := os.WriteFile(file, []byte("step,up,right\n1,1,3\n2,2,4\n"), 0o644); err != nil {
		t.Fatalf("write csv: %v", err)
	}

	profile, err := LoadDemandProfile(file, "right")
	if err != nil {
		t.Fatalf("load profile: %v", err)
	}
	if profile[1] != 3 || profile[2] != 4 {
		t.Fatalf("unexpected profile values: %#v", profile)
	}
}
