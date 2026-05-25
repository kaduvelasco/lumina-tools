package tui

// Styles are no longer global variables.
// The Model holds a TUIStyles value (see theme.go) built from the active Theme.
// Use m.styles.X throughout the TUI instead of package-level style variables.
