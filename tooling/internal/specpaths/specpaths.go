package specpaths

import "fmt"

const (
	ModulesRootDir                 = "docs/specs/modules"
	CandidateDir                   = ModulesRootDir + "/candidate"
	StableDir                      = ModulesRootDir + "/stable"
	CandidateAppendixDir           = CandidateDir + "/appendix"
	StableAppendixDir              = StableDir + "/appendix"
	SystemConstraintsStableFileRef = "docs/specs/system/stable/s_system_constraints.md"
)

func MainSpecFileRef(layer, module string) (string, error) {
	switch layer {
	case "candidate":
		return fmt.Sprintf("%s/c_%s.md", CandidateDir, module), nil
	case "stable":
		return fmt.Sprintf("%s/s_%s.md", StableDir, module), nil
	default:
		return "", fmt.Errorf("unsupported layer %q", layer)
	}
}

func AppendixDir(layer string) (string, error) {
	switch layer {
	case "candidate":
		return CandidateAppendixDir, nil
	case "stable":
		return StableAppendixDir, nil
	default:
		return "", fmt.Errorf("unsupported layer %q", layer)
	}
}

func CandidateAppendixGlob(module string) string {
	return fmt.Sprintf("%s/c_%s_*.md", CandidateAppendixDir, module)
}
