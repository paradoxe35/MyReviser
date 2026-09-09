package language

import "testing"

func TestFindIsCaseAndSpaceInsensitive(t *testing.T) {
	for _, code := range []string{"FR", " fr ", "Fr"} {
		if got := Find(code).Code; got != "fr" {
			t.Errorf("Find(%q).Code = %q, want fr", code, got)
		}
	}
}

func TestFindUnknownCodeSurvives(t *testing.T) {
	got := Find("zzz")
	if got.Code != "zzz" || got.Name != "zzz" {
		t.Errorf("Find(zzz) = %+v, want both fields zzz", got)
	}
	if IsKnown("zzz") {
		t.Error("IsKnown(zzz) = true, want false")
	}
}

func TestSearchPrefersPrefixMatches(t *testing.T) {
	results := Search("en")
	if len(results) == 0 {
		t.Fatal("Search(en) returned nothing")
	}
	if results[0].Code != "en" {
		t.Errorf("Search(en)[0] = %q, want en to sort first", results[0].Code)
	}
}

func TestSearchMatchesEndonym(t *testing.T) {
	for _, l := range Search("français") {
		if l.Code == "fr" {
			return
		}
	}
	t.Error("Search(français) did not find fr")
}

func TestSearchEmptyReturnsAll(t *testing.T) {
	if len(Search("  ")) != len(All()) {
		t.Error("blank query should return the whole registry")
	}
}

func TestDefaultsNeverDegenerate(t *testing.T) {
	cases := []string{"en", "en-US", "fr", "fr-FR", "de", "zzz", ""}
	for _, tag := range cases {
		primary, secondary := Defaults(tag)
		if primary == secondary {
			t.Errorf("Defaults(%q) = %q, %q; pair must differ", tag, primary, secondary)
		}
		if primary == "" || secondary == "" {
			t.Errorf("Defaults(%q) returned an empty code", tag)
		}
	}
}

func TestDefaultsUsesRegionSubtag(t *testing.T) {
	if primary, _ := Defaults("de-AT"); primary != "de" {
		t.Errorf("Defaults(de-AT) primary = %q, want de", primary)
	}
}

func TestEndonymFallsBackToEnglishName(t *testing.T) {
	unknown := Language{Code: "zzz", Name: "Zzz"}
	if got := unknown.Endonym(); got != "Zzz" {
		t.Errorf("Endonym() = %q, want the English name", got)
	}
}

func TestRegistryHasNoDuplicateCodes(t *testing.T) {
	seen := make(map[string]bool, len(all))
	for _, l := range all {
		if seen[l.Code] {
			t.Errorf("duplicate code %q", l.Code)
		}
		seen[l.Code] = true
	}
}
