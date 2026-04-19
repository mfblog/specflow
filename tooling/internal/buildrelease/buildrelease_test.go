package buildrelease

import (
	"strings"
	"testing"
)

func TestLdflagsForFingerprintFixesBuildID(t *testing.T) {
	ldflags := ldflagsForFingerprint("abc123")

	if !strings.Contains(ldflags, "-buildid=") {
		t.Fatalf("expected ldflags to clear Go build id, got %q", ldflags)
	}
	if !strings.Contains(ldflags, "toolingfreshness.BuildFingerprint=abc123") {
		t.Fatalf("expected ldflags to embed tooling fingerprint, got %q", ldflags)
	}
}

func TestBuildCommandArgsDisablesVCSMetadata(t *testing.T) {
	args := strings.Join(buildCommandArgs("flags", "out"), " ")

	if !strings.Contains(args, "-buildvcs=false") {
		t.Fatalf("expected build args to disable VCS metadata, got %q", args)
	}
}
