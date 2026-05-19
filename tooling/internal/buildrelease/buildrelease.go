package buildrelease

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/toolingfreshness"
)

type Target struct {
	GOOS   string
	GOARCH string
}

type BuildResult struct {
	Targets []string
}

var DefaultTargets = []Target{
	{GOOS: "linux", GOARCH: "amd64"},
	{GOOS: "linux", GOARCH: "arm64"},
	{GOOS: "darwin", GOARCH: "amd64"},
	{GOOS: "darwin", GOARCH: "arm64"},
	{GOOS: "windows", GOARCH: "amd64"},
	{GOOS: "windows", GOARCH: "arm64"},
}

func BinaryName(goos, goarch string) string {
	return ToolBinaryName("specflowctl", goos, goarch)
}

func ReaderBinaryName(goos, goarch string) string {
	return ToolBinaryName("specflow-reader", goos, goarch)
}

func ToolBinaryName(base, goos, goarch string) string {
	name := fmt.Sprintf("specflowctl-%s-%s", goos, goarch)
	if base != "specflowctl" {
		name = fmt.Sprintf("%s-%s-%s", base, goos, goarch)
	}
	if goos == "windows" {
		return name + ".exe"
	}
	return name
}

func CurrentBinaryName() string {
	return BinaryName(runtime.GOOS, runtime.GOARCH)
}

func CurrentReaderBinaryName() string {
	return ReaderBinaryName(runtime.GOOS, runtime.GOARCH)
}

func BuildAll(repoRoot string, targets []Target) (BuildResult, error) {
	if len(targets) == 0 {
		targets = DefaultTargets
	}

	fingerprint, _, err := toolingfreshness.LiveFingerprint(repoRoot)
	if err != nil {
		return BuildResult{}, err
	}

	binDir := filepath.Join(repoRoot, "specflow/tooling/bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return BuildResult{}, fmt.Errorf("mkdir bin dir: %w", err)
	}
	cacheDir := filepath.Join(repoRoot, ".tmp", "go-build")
	modCacheDir := filepath.Join(repoRoot, ".tmp", "go-mod-cache")

	result := BuildResult{Targets: make([]string, 0, len(targets))}
	for _, target := range targets {
		builds := []struct {
			outputName  string
			packagePath string
		}{
			{outputName: BinaryName(target.GOOS, target.GOARCH), packagePath: "./cmd/specflowctl"},
			{outputName: ReaderBinaryName(target.GOOS, target.GOARCH), packagePath: "./cmd/specflow-reader"},
		}
		for _, build := range builds {
			outputPath := filepath.Join(binDir, build.outputName)
			ldflags := ldflagsForFingerprint(fingerprint)
			cmd := exec.Command("go", buildCommandArgs(ldflags, outputPath, build.packagePath)...)
			cmd.Dir = filepath.Join(repoRoot, "specflow/tooling")
			cmd.Env = append(os.Environ(),
				"GOOS="+target.GOOS,
				"GOARCH="+target.GOARCH,
				"CGO_ENABLED=0",
				"GOCACHE="+cacheDir,
				"GOMODCACHE="+modCacheDir,
			)
			if output, err := cmd.CombinedOutput(); err != nil {
				return result, fmt.Errorf("build %s/%s %s failed: %v: %s", target.GOOS, target.GOARCH, build.packagePath, err, string(output))
			}
			result.Targets = append(result.Targets, filepath.ToSlash(filepath.Join("specflow/tooling/bin", build.outputName)))
		}
	}

	return result, nil
}

func ldflagsForFingerprint(fingerprint string) string {
	return fmt.Sprintf(
		"-s -w -buildid= -X github.com/Bingordinary/SpecFlow/specflow/tooling/internal/toolingfreshness.BuildFingerprint=%s",
		fingerprint,
	)
}

func buildCommandArgs(ldflags, outputPath, packagePath string) []string {
	return []string{
		"build",
		"-trimpath",
		"-buildvcs=false",
		"-ldflags=" + ldflags,
		"-o",
		outputPath,
		packagePath,
	}
}
