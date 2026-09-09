package ui

import (
	"testing"

	"github.com/paradoxe35/encre/internal/stt"
)

func TestActiveModelTextRemainsVisibleForLongLists(t *testing.T) {
	models := []stt.Model{
		{ID: "small", Name: "Small English"},
		{ID: "large", Name: "Large Multilingual"},
	}

	if got := activeModelText("large", models); got != "Active model: Large Multilingual" {
		t.Fatalf("active model label = %q", got)
	}
	if got := activeModelText("missing", models); got != "No active model selected" {
		t.Fatalf("missing model label = %q", got)
	}
}
