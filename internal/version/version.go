package version

import (
	"fmt"
	"regexp"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
)

// Semantic returns a valid Semantic Version based on the given string. Empty
// strings are replaced by the embedded module version, if available. Valid
// strings are returned as-is. Strings in `git describe --long`-format are
// converted. Returns "0.0.0-dev" if both validation and conversion fails.
func Semantic(v string) string {
	if v == "" {
		info, _ := debug.ReadBuildInfo()
		v = info.Main.Version
	}

	v = strings.TrimPrefix(v, "v")

	if reSemantic.MatchString(v) {
		return v
	}
	if ver, ok := fromGitDescribeLong(v); ok {
		return ver.String()
	}

	return "0.0.0-dev"
}

var reSemantic = regexp.MustCompile(`^\d+.\d+.\d+(-[0-9a-z.]+)?(\+[0-9a-z.]+)?$`)

var reGitDescribeLong = regexp.MustCompile(
	`^(\d+).(\d+).(\d+)(?:-([0-9a-z.]+))?-(\d+)-g([0-9a-f]+)(?:-(dirty))?$`)

const (
	idxMajor = iota + 1
	idxMinor
	idxPatch
	idxPrerelease
	idxAdditionalCommits
	idxRevision
	idxDirty
)

type version struct {
	major, minor, patch int
	prerelease          string

	additionalCommits int
	revision          string
	isDirty           bool
}

func fromGitDescribeLong(s string) (version, bool) {
	m := reGitDescribeLong.FindStringSubmatch(s)
	if len(m) == 0 {
		return version{}, false
	}

	return version{
		major:             atoi(m[idxMajor]),
		minor:             atoi(m[idxMinor]),
		patch:             atoi(m[idxPatch]),
		prerelease:        m[idxPrerelease],
		additionalCommits: atoi(m[idxAdditionalCommits]),
		revision:          m[idxRevision],
		isDirty:           m[idxDirty] == "dirty",
	}, true
}

func (v version) String() string {
	patch := v.patch
	if v.prerelease == "" && !v.isPure() {
		patch++
	}
	ver := fmt.Sprintf("%d.%d.%d", v.major, v.minor, patch)

	pre := v.prerelease
	if !v.isPure() {
		pre = join(".", pre, "dev", strconv.Itoa(v.additionalCommits))
	}
	ver = join("-", ver, pre)

	if v.isPure() {
		return ver
	}

	build := v.revision
	if v.isDirty {
		build = join(".", build, "dirty")
	}

	return join("+", ver, build)
}

func (v version) isPure() bool {
	return v.additionalCommits == 0 && !v.isDirty
}

func join(sep string, elems ...string) string {
	elems = slices.DeleteFunc(elems, func(s string) bool {
		return s == ""
	})
	return strings.Join(elems, sep)
}

func atoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
