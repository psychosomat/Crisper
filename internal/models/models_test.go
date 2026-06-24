package models

import (
	"testing"
	"Crisper/internal/hardware"
)

func TestRecommendModel(t *testing.T) {
	specs := hardware.Detect()
	r := RecommendModel()
	t.Logf("CPU: %d threads, RAM: %.1f GB → Recommended: %s (%.2f GB, %.1fx speed)",
		specs.CPUThreads, specs.TotalRAMGB, r.DisplayName, r.SizeGB, r.SpeedFactor)

	if specs.CPUThreads <= 4 && specs.TotalRAMGB >= 4 {
		if r.SizeGB > 1.5 {
			t.Errorf("low-thread CPU (%d) should not get model >1.5 GB, got %s", specs.CPUThreads, r.Name)
		}
	}
	if specs.TotalRAMGB < r.MinRAMGB {
		t.Errorf("recommended model %s needs %.1f GB RAM but system has %.1f GB", r.Name, r.MinRAMGB, specs.TotalRAMGB)
	}
}
