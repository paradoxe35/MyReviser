package stt

import "testing"

func TestEstimatedRealtimeScalesWithCores(t *testing.T) {
	model := Model{RealtimeFactor: 8}

	weak := model.EstimatedRealtime(Machine{Cores: 4})
	same := model.EstimatedRealtime(Machine{Cores: referenceCores})
	strong := model.EstimatedRealtime(Machine{Cores: 16})

	if same != 8 {
		t.Errorf("reference machine should reproduce the measured factor, got %v", same)
	}
	if !(weak < same && same < strong) {
		t.Errorf("expected %v < %v < %v", weak, same, strong)
	}
}

func TestEstimatedRealtimeWithoutMeasurement(t *testing.T) {
	if got := (Model{}).EstimatedRealtime(Machine{Cores: 8}); got != 0 {
		t.Errorf("an unmeasured model should not claim a speed, got %v", got)
	}
}

func TestFitsMemoryLeavesHeadroom(t *testing.T) {
	host := Machine{Cores: 8, MemoryMB: 8000}

	if !(Model{SizeBytes: 1 << 30}).FitsMemory(host) {
		t.Error("1 GB should fit in 8 GB")
	}
	if (Model{SizeBytes: 6 << 30}).FitsMemory(host) {
		t.Error("6 GB should not be offered on an 8 GB machine")
	}
}

// An unknown memory size must not hide every model.
func TestFitsMemoryUnknownIsPermissive(t *testing.T) {
	if !(Model{SizeBytes: 40 << 30}).FitsMemory(Machine{Cores: 8}) {
		t.Error("with memory unknown the check should pass rather than exclude")
	}
}

func TestFitClassification(t *testing.T) {
	host := Machine{Cores: 8, MemoryMB: 8000}

	cases := []struct {
		name  string
		model Model
		want  Fit
	}{
		{"fast and small", Model{RealtimeFactor: 20, SizeBytes: 100 << 20}, FitComfortable},
		{"slow", Model{RealtimeFactor: 1, SizeBytes: 100 << 20}, FitSlow},
		{"huge", Model{RealtimeFactor: 20, SizeBytes: 7 << 30}, FitTooLarge},
		// A custom model has no benchmark; guessing "slow" would misstate what the catalog knows.
		{"unmeasured", Model{SizeBytes: 100 << 20}, FitUnknown},
	}

	for _, c := range cases {
		if got := c.model.Fit(host); got != c.want {
			t.Errorf("%s: Fit() = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestRankPrefersDownloadedThenFit(t *testing.T) {
	host := Machine{Cores: 8, MemoryMB: 8000}

	huge := Model{ID: "huge", RealtimeFactor: 20, SizeBytes: 7 << 30}
	slow := Model{ID: "slow", RealtimeFactor: 1, SizeBytes: 100 << 20}
	quick := Model{ID: "quick", RealtimeFactor: 20, SizeBytes: 100 << 20}

	models := []Model{huge, slow, quick}
	RankForMachine(models, host, nil)

	if models[0].ID != "quick" || models[1].ID != "slow" || models[2].ID != "huge" {
		t.Fatalf("ranked %s, %s, %s; want quick, slow, huge",
			models[0].ID, models[1].ID, models[2].ID)
	}

	// A downloaded model outranks a better one that is not on disk.
	models = []Model{quick, slow}
	RankForMachine(models, host, func(m Model) bool { return m.ID == "slow" })
	if models[0].ID != "slow" {
		t.Errorf("a downloaded model should sort first, got %s", models[0].ID)
	}
}

func TestHostReportsCores(t *testing.T) {
	if Host().Cores <= 0 {
		t.Error("core count should be positive")
	}
}
