package pkg

import "runtime/debug"

// Version is stamped at build time by the release pipeline via:
//
//	-ldflags "-X github.com/armandwipangestu/fiber-boilerplate/internal/pkg.Version=<semver>"
//
// See .github/workflows/release.yml for the semantic-release wiring.
var Version = "dev"

// EffectiveVersion returns the stamped version when it was set at build time,
// otherwise a dev build id derived from the embedded VCS revision.
func EffectiveVersion() string {
	if Version != "" && Version != "dev" {
		return Version
	}
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return Version
	}
	for _, s := range bi.Settings {
		if s.Key == "vcs.revision" && s.Value != "" {
			rev := s.Value
			if len(rev) > 7 {
				rev = rev[:7]
			}
			return "dev-" + rev
		}
	}
	return Version
}
