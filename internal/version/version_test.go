package version_test

import (
	"testing"

	"github.com/lende/127/internal/version"
)

func TestSemantic(t *testing.T) {
	t.Parallel()
	const fallbackVersion = "0.0.0-dev"

	for _, test := range []struct{ input, want string }{
		{"v0.4.0", "0.4.0"},
		{"v0.4.0-0-gc774e5a", "0.4.0"},
		{"v0.4.0-alpha", "0.4.0-alpha"},
		{"v0.4.0-alpha-0-gc774e5a", "0.4.0-alpha"},
		{"v0.4.0-alpha.2", "0.4.0-alpha.2"},
		{"v0.4.0-alpha.2-0-gc774e5a", "0.4.0-alpha.2"},
		{"v0.4.0-0-gc774e5a-dirty", "0.4.1-dev.0+c774e5a.dirty"},
		{"v0.4.0-alpha-3-gc774e5a-dirty", "0.4.0-alpha.dev.3+c774e5a.dirty"},
		{"v0.4.0-foo.bar.2-1-gc774e5a", "0.4.0-foo.bar.2.dev.1+c774e5a"},
		{"0.3.0", "0.3.0"},
		{"0.3.0-alpha-3-gc774e5a-dirty", "0.3.0-alpha.dev.3+c774e5a.dirty"},

		{"", fallbackVersion},
		{"INVALID", fallbackVersion},
		{"v0.4", fallbackVersion},
		{"v0.4.0-3-gc774e5a-INVALID", fallbackVersion},
		{"v0.4.0-3-INVALID-gc774e5a-dirty", fallbackVersion},
		{"v0.4.0-alpha-INVALID-3-gc774e5a", fallbackVersion},
	} {
		if got := version.Semantic(test.input); got != test.want {
			t.Errorf("version.Clean(%q) == %q != %q", test.input, got, test.want)
		}
	}
}
