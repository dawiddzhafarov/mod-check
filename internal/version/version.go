package version

import (
	"errors"
	"fmt"
	"github.com/apex/log"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type (
	// Versions represents a list of versions of given dependency
	Versions []Version

	// Version represents single version of given dependency
	Version struct {
		Major, Minor, Patch uint64
		Prerelease          string
		Metadata            string
		Incompatible        bool
		Original            string
		Status              string
	}
)

const (
	allowed = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-0123456789"
	goProxy = "https://proxy.golang.org/%s/@v/list"
)

var versionRegex = regexp.MustCompile(`^v?([0-9]+)(\.[0-9]+)?(\.[0-9]+)?(-([0-9A-Za-z\-]+(\.[0-9A-Za-z\-]+)*))?(\+([0-9A-Za-z\-]+(\.[0-9A-Za-z\-]+)*))?$`)

func (v Versions) Len() int {
	return len(v)
}

func (v Versions) Less(i, j int) bool {
	return v[i].Compare(v[j]) > 0
}

func (v Versions) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v *Version) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func GetProxyVersions(current *Version, url string, skipPrelease bool) Versions {
	resp, err := http.Get(fmt.Sprintf(goProxy, url))
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("%v", err)
	}

	var versions Versions
	versionItem := strings.Split(string(body), "\n")
	for _, ver := range versionItem {
		parsedVersion, err := ParseVersion(ver, current)
		if err != nil {
			// If has error parsing skip it
			continue // TODO should we return an error here?
		}
		if skipPrelease && parsedVersion.Prerelease != "" {
			continue
		}
		versions = append(versions, *parsedVersion)
	}
	sort.Sort(versions)

	return versions
}

// ParseVersion parses a given module version
func ParseVersion(rawVersion string, current *Version) (*Version, error) {
	m := versionRegex.FindStringSubmatch(rawVersion)
	if m == nil {
		return nil, errors.New("invalid semantic version")
	}

	if len(m) < 9 {
		return nil, errors.New("invalid semantic version")
	}
	major := m[1]
	minor := m[2]
	patch := m[3]
	prerelease := m[5]
	metadata := m[8]

	parsedVersion := &Version{
		Prerelease:   prerelease,
		Metadata:     metadata,
		Original:     rawVersion,
		Incompatible: strings.Contains(rawVersion, "incompatible"),
	}

	var err error

	// Check major number
	parsedVersion.Major, err = strconv.ParseUint(major, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing version segment: %s", err)
	}

	// Check minor number
	if minor != "" {
		parsedVersion.Minor, err = strconv.ParseUint(strings.TrimPrefix(minor, "."), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing version segment: %s", err)
		}
	} else {
		parsedVersion.Minor = 0
	}

	// Check patch number
	if patch != "" {
		parsedVersion.Patch, err = strconv.ParseUint(strings.TrimPrefix(patch, "."), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing version segment: %s", err)
		}
	} else {
		parsedVersion.Patch = 0
	}

	// Check prerelease
	if parsedVersion.Prerelease != "" {
		if err = validatePrerelease(parsedVersion.Prerelease); err != nil {
			return nil, err
		}
	}

	// Check metadata
	if parsedVersion.Metadata != "" {
		if err = validateMetadata(parsedVersion.Metadata); err != nil {
			return nil, err
		}
	}

	if current != nil {
		status, err := reconcileStatus(current, parsedVersion)
		if err != nil {
			return nil, err
		}
		parsedVersion.Status = status
	}

	return parsedVersion, nil
}

func (v *Version) Compare(v2 Version) int {
	if d := compareSegment(v.Major, v2.Major); d != 0 {
		return d
	}
	if d := compareSegment(v.Minor, v2.Minor); d != 0 {
		return d
	}
	if d := compareSegment(v.Patch, v2.Patch); d != 0 {
		return d
	}

	// If major, minor, and patch are the same, check prerelease
	if v.Prerelease == "" && v2.Prerelease == "" {
		return 0
	}
	if v.Prerelease == "" {
		return 1
	}
	if v2.Prerelease == "" {
		return -1
	}

	return 0
}

func compareSegment(v, o uint64) int {
	if v < o {
		return -1
	}
	if v > o {
		return 1
	}

	return 0
}

// Like strings.ContainsAny but does an only instead of any.
func containsOnly(s string, comp string) bool {
	return strings.IndexFunc(s, func(r rune) bool {
		return !strings.ContainsRune(comp, r)
	}) == -1
}

// validatePrerelease loops through values to check for valid characters
func validatePrerelease(p string) error {
	parts := strings.Split(p, ".")
	for _, p := range parts {
		if containsOnly(p, "0123456789") {
			if len(p) > 1 && p[0] == '0' {
				return errors.New("version segment starts with 0")
			}
		} else if !containsOnly(p, allowed) {
			return errors.New("invalid Prerelease string")
		}
	}
	return nil
}

// validateMetadata loops through values to check for valid characters
func validateMetadata(m string) error {
	parts := strings.Split(m, ".")
	for _, p := range parts {
		if !containsOnly(p, allowed) {
			return errors.New("invalid metadata string")
		}
	}
	return nil
}

func reconcileStatus(current, latest *Version) (string, error) {
	if current.Major == latest.Major && current.Minor == latest.Minor && current.Patch == latest.Patch {
		return "current", nil
	} else if latest.Major > current.Major {
		return "major", nil
	} else if latest.Minor > current.Minor {
		return "minor", nil
	} else if latest.Patch > current.Patch {
		return "patch", nil
	}

	return "", fmt.Errorf("failed to reconcile version status")
}
