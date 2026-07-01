package gnome

import (
	_ "embed"
	"fmt"
	"sync"

	"gopkg.in/yaml.v3"
)

//go:embed gnome_themes.yaml
var gnomeThemesYAML []byte

//go:embed cinnamon_themes.yaml
var cinnamonThemesYAML []byte

//go:embed cursors.yaml
var cursorsYAML []byte

//go:embed icons.yaml
var iconsYAML []byte

var (
	themeOnce  sync.Once
	themeCache []themeEntry
	themeErr   error

	cinnamonOnce  sync.Once
	cinnamonCache []themeEntry
	cinnamonErr   error

	cursorOnce  sync.Once
	cursorCache []cursorEntry
	cursorErr   error

	iconOnce  sync.Once
	iconCache []iconEntry
	iconErr   error
)

// parseThemeCatalogue is the shared loader for the theme catalogues (GNOME,
// Cinnamon). Both use the same YAML shape and themeEntry type; only the
// sync.Once guard, raw bytes, display name, and cache pointers differ.
func parseThemeCatalogue(once *sync.Once, raw []byte, name string, cache *[]themeEntry, cerr *error) ([]themeEntry, error) {
	once.Do(func() {
		var w struct {
			Themes []themeEntry `yaml:"themes"`
		}
		if err := yaml.Unmarshal(raw, &w); err != nil {
			*cerr = fmt.Errorf("parse %s catalogue: %w", name, err)
			return
		}
		*cache = w.Themes
	})
	return *cache, *cerr
}

func loadThemeCatalogue() ([]themeEntry, error) {
	return parseThemeCatalogue(&themeOnce, gnomeThemesYAML, "gnome themes", &themeCache, &themeErr)
}

func loadCinnamonThemeCatalogue() ([]themeEntry, error) {
	return parseThemeCatalogue(&cinnamonOnce, cinnamonThemesYAML, "cinnamon themes", &cinnamonCache, &cinnamonErr)
}

func loadCursorCatalogue() ([]cursorEntry, error) {
	cursorOnce.Do(func() {
		var w struct {
			Cursors []cursorEntry `yaml:"cursors"`
		}
		if err := yaml.Unmarshal(cursorsYAML, &w); err != nil {
			cursorErr = fmt.Errorf("parse cursor catalogue: %w", err)
			return
		}
		cursorCache = w.Cursors
	})
	return cursorCache, cursorErr
}

func loadIconCatalogue() ([]iconEntry, error) {
	iconOnce.Do(func() {
		var w struct {
			Icons []iconEntry `yaml:"icons"`
		}
		if err := yaml.Unmarshal(iconsYAML, &w); err != nil {
			iconErr = fmt.Errorf("parse icon catalogue: %w", err)
			return
		}
		iconCache = w.Icons
	})
	return iconCache, iconErr
}
