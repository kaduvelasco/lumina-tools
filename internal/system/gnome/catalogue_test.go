package gnome

import "testing"

func TestLoadThemeCatalogue(t *testing.T) {
	entries, err := loadThemeCatalogue()
	if err != nil {
		t.Fatalf("loadThemeCatalogue() error: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("loadThemeCatalogue() returned empty catalogue")
	}
	for _, e := range entries {
		if e.Name == "" {
			t.Error("theme entry with empty name")
		}
		if e.DirPattern == "" {
			t.Errorf("theme %q: empty dir_pattern", e.Name)
		}
	}
}

func TestLoadCinnamonThemeCatalogue(t *testing.T) {
	entries, err := loadCinnamonThemeCatalogue()
	if err != nil {
		t.Fatalf("loadCinnamonThemeCatalogue() error: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("loadCinnamonThemeCatalogue() returned empty catalogue")
	}
	for _, e := range entries {
		if e.Name == "" {
			t.Error("cinnamon theme entry with empty name")
		}
	}
}

func TestLoadXFCEThemeCatalogue(t *testing.T) {
	entries, err := loadXFCEThemeCatalogue()
	if err != nil {
		t.Fatalf("loadXFCEThemeCatalogue() error: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("loadXFCEThemeCatalogue() returned empty catalogue")
	}
	for _, e := range entries {
		if e.Name == "" {
			t.Error("xfce theme entry with empty name")
		}
		if e.DirPattern == "" {
			t.Errorf("xfce theme %q: empty dir_pattern", e.Name)
		}
		if e.RepoURL == "" && e.UserScript == "" {
			t.Errorf("xfce theme %q: needs either repo_url or user_script", e.Name)
		}
	}
}

func TestLoadCursorCatalogue(t *testing.T) {
	entries, err := loadCursorCatalogue()
	if err != nil {
		t.Fatalf("loadCursorCatalogue() error: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("loadCursorCatalogue() returned empty catalogue")
	}
	for _, e := range entries {
		if e.Name == "" {
			t.Error("cursor entry with empty name")
		}
		if e.DirPattern == "" {
			t.Errorf("cursor %q: empty dir_pattern", e.Name)
		}
	}
}

func TestLoadIconCatalogue(t *testing.T) {
	entries, err := loadIconCatalogue()
	if err != nil {
		t.Fatalf("loadIconCatalogue() error: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("loadIconCatalogue() returned empty catalogue")
	}
	for _, e := range entries {
		if e.Name == "" {
			t.Error("icon entry with empty name")
		}
		if e.DirPattern == "" {
			t.Errorf("icon %q: empty dir_pattern", e.Name)
		}
	}
}

func TestThemeCatalogueYaruScript(t *testing.T) {
	entries, err := loadThemeCatalogue()
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Name == "Yaru" {
			if e.CustomScript == "" {
				t.Error("Yaru theme: custom_script is empty")
			}
			if len(e.PurgePackages) == 0 {
				t.Error("Yaru theme: purge_packages is empty")
			}
			return
		}
	}
	t.Error("Yaru theme not found in catalogue")
}
