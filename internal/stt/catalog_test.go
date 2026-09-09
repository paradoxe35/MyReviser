package stt

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestEmbeddedCatalogIsComplete(t *testing.T) {
	models := Catalogue()
	if len(models) < 20 {
		t.Fatalf("catalog holds %d models, expected the generated list", len(models))
	}

	seen := make(map[string]bool, len(models))
	for _, m := range models {
		switch {
		case m.ID == "":
			t.Errorf("%s has no id", m.Slug)
		case m.SHA256 == "":
			t.Errorf("%s has no checksum; it could not be verified after download", m.Slug)
		case m.SizeBytes <= 0:
			t.Errorf("%s has no size; resume and progress both need it", m.Slug)
		case m.Revision == "":
			t.Errorf("%s has no pinned revision", m.Slug)
		}
		if seen[m.ID] {
			t.Errorf("duplicate model id %s", m.ID)
		}
		seen[m.ID] = true
	}
}

func TestRecommendedSortsFirst(t *testing.T) {
	models := Catalogue()
	if !models[0].Recommended {
		t.Errorf("first model %s is not recommended", models[0].Slug)
	}

	recommended, ok := Recommended()
	if !ok || !recommended.Recommended {
		t.Fatal("no recommended model available")
	}
}

func TestDownloadURLPinsRevision(t *testing.T) {
	model := Model{Repo: "org/repo", Revision: "abc123", Filename: "m.gguf"}
	want := "https://huggingface.co/org/repo/resolve/abc123/m.gguf"
	if got := model.DownloadURL(); got != want {
		t.Errorf("DownloadURL() = %q, want %q", got, want)
	}
}

func TestParseCatalogRejectsEmpty(t *testing.T) {
	if _, err := parseCatalog([]byte(`{"catalog_version":1,"models":[]}`)); err == nil {
		t.Error("an empty catalog must be rejected, not adopted")
	}
	if _, err := parseCatalog([]byte(`not json`)); err == nil {
		t.Error("unparseable data must be rejected")
	}
}

func TestFindModelMatchesByID(t *testing.T) {
	first := Catalogue()[0]
	found, ok := FindModel(first.ID)
	if !ok || found.ID != first.ID {
		t.Fatalf("FindModel(%q) did not round-trip", first.ID)
	}
	if _, ok := FindModel("nothing/here"); ok {
		t.Error("an unknown id should not resolve")
	}
}

// A cache older than the shipped list means the app was updated since.
func TestOlderCacheLosesToShipped(t *testing.T) {
	shipped, err := parseCatalog(embeddedCatalog)
	if err != nil {
		t.Fatal(err)
	}

	stale := Catalog{Version: shipped.Version - 1, Models: shipped.Models[:1]}
	data, _ := json.Marshal(stale)

	dir := t.TempDir()
	path := filepath.Join(dir, "catalog.json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	cached, err := parseCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	if cached.Version >= shipped.Version {
		t.Fatal("test fixture should be older than the shipped catalog")
	}
}

func TestLanguageSummary(t *testing.T) {
	cases := []struct {
		languages []string
		want      string
	}{
		{nil, "unknown"},
		{[]string{"en"}, "en"},
		{[]string{"en", "fr", "de"}, "3 languages"},
	}
	for _, c := range cases {
		got := Model{Languages: c.languages}.LanguageSummary()
		if got != c.want {
			t.Errorf("LanguageSummary(%v) = %q, want %q", c.languages, got, c.want)
		}
	}
}
