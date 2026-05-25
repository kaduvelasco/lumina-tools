package version

// Version is injected at build time via -ldflags.
// Default value is used for development builds.
var Version = "dev"
