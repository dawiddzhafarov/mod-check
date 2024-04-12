package module

import (
	"mod-check/internal/version"
)

// Modules is an alias to a slice of modules
type Modules []Module

// Module is the main struct containing path, current version, available versions and status
type Module struct {
	Path              string
	CurrentVersion    *version.Version
	AvailableVersions version.Versions
	IsCurrent         bool
}

// NewModule parses the current version and gets all possible versions for given dependency
func NewModule(path, ver string) (*Module, error) {
	// Parse current version
	current, err := version.ParseVersion(ver, nil)
	if err != nil {
		return nil, err
	}

	// Get all available versions
	versions := version.GetProxyVersions(current, path, true)
	if len(versions) == 0 {
		return nil, nil // TODO shouldn't return error, simply ignore if its up to date, prev it was nil, fmt.Error(..)
	}

	latest := versions[0]

	// Check if current version is up-to-date
	var isCurrent bool
	if current.Original == latest.Original {
		isCurrent = true
	}

	return &Module{
		Path:              path,
		CurrentVersion:    current,
		AvailableVersions: versions,
		IsCurrent:         isCurrent,
	}, nil
}
