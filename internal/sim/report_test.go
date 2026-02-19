package sim

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWriteReportCreatesOutputFile(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "reports", "result.json")
	report := Report{
		ConfigName: "test-scenario",
		Generated:  time.Unix(1700000000, 0).UTC(),
		Metrics: Metrics{
			ScenarioName:      "test-scenario",
			Steps:             5,
			VehiclesSpawned:   3,
			VehiclesCompleted: 2,
		},
	}

	if err := WriteReport(path, report); err != nil {
		t.Fatalf("write report: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	text := string(data)
	if !strings.Contains(text, `"config_name": "test-scenario"`) {
		t.Fatalf("report missing config name: %s", text)
	}
	if !strings.Contains(text, `"vehicles_completed": 2`) {
		t.Fatalf("report missing metrics: %s", text)
	}
}
