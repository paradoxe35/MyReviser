package ui

import (
	"testing"

	"github.com/paradoxe35/encre/internal/stt"
)

func TestActiveModelTextRemainsVisibleForLongLists(t *testing.T) {
	models := []stt.Model{
		{ID: "small", Name: "Small English", SizeBytes: 1},
		{ID: "large", Name: "Large Multilingual", SizeBytes: 1},
	}
	downloaded := func(model stt.Model) bool { return model.ID == "large" }

	if got := activeModelText("large", models, downloaded); got != "Active model: Large Multilingual" {
		t.Fatalf("active model label = %q", got)
	}
	if got := activeModelText("small", models, downloaded); got != "No active model selected" {
		t.Fatalf("undownloaded model label = %q", got)
	}
	if got := activeModelText("missing", models, downloaded); got != "No active model selected" {
		t.Fatalf("missing model label = %q", got)
	}
}
