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

//go:embed xfce_themes.yaml
var xfceThemesYAML []byte

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

	xfceOnce  sync.Once
	xfceCache []themeEntry
	xfceErr   error

	cursorOnce  sync.Once
	cursorCache []cursorEntry
	cursorErr   error

	iconOnce  sync.Once
	iconCache []iconEntry
	iconErr   error
)

func loadThemeCatalogue() ([]themeEntry, error) {
	themeOnce.Do(func() {
		var w struct {
			Themes []themeEntry `yaml:"themes"`
		}
		if err := yaml.Unmarshal(gnomeThemesYAML, &w); err != nil {
			themeErr = fmt.Errorf("parse gnome themes catalogue: %w", err)
			return
		}
		themeCache = w.Themes
	})
	return themeCache, themeErr
}

func loadCinnamonThemeCatalogue() ([]themeEntry, error) {
	cinnamonOnce.Do(func() {
		var w struct {
			Themes []themeEntry `yaml:"themes"`
		}
		if err := yaml.Unmarshal(cinnamonThemesYAML, &w); err != nil {
			cinnamonErr = fmt.Errorf("parse cinnamon themes catalogue: %w", err)
			return
		}
		cinnamonCache = w.Themes
	})
	return cinnamonCache, cinnamonErr
}

func loadXFCEThemeCatalogue() ([]themeEntry, error) {
	xfceOnce.Do(func() {
		var w struct {
			Themes []themeEntry `yaml:"themes"`
		}
		if err := yaml.Unmarshal(xfceThemesYAML, &w); err != nil {
			xfceErr = fmt.Errorf("parse xfce themes catalogue: %w", err)
			return
		}
		xfceCache = w.Themes
	})
	return xfceCache, xfceErr
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
