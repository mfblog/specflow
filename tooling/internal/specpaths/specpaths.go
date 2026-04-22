package specpaths

import "fmt"

const (
	ModulesRootDir                 = "docs/specs/modules"
	CandidateDir                   = ModulesRootDir + "/candidate"
	StableDir                      = ModulesRootDir + "/stable"
	CandidateAppendixDir           = CandidateDir + "/appendix"
	StableAppendixDir              = StableDir + "/appendix"
	FlowsRootDir                   = "docs/specs/flows"
	CandidateFlowDir               = FlowsRootDir + "/candidate"
	StableFlowDir                  = FlowsRootDir + "/stable"
	ProjectRootDir                 = "docs/specs/project"
	CandidateProjectDir            = ProjectRootDir + "/candidate"
	StableProjectDir               = ProjectRootDir + "/stable"
	SystemConstraintsStableFileRef = "docs/specs/system/stable/s_system_constraints.md"
)

func MainSpecFileRef(layer, module string) (string, error) {
	return ObjectMainSpecFileRef("module", layer, module)
}

func ObjectMainSpecFileRef(objectType, layer, object string) (string, error) {
	switch layer {
	case "candidate":
		switch objectType {
		case "module":
			return fmt.Sprintf("%s/c_%s.md", CandidateDir, object), nil
		case "flow":
			return fmt.Sprintf("%s/c_flow_%s.md", CandidateFlowDir, object), nil
		case "project":
			return fmt.Sprintf("%s/c_project.md", CandidateProjectDir), nil
		}
	case "stable":
		switch objectType {
		case "module":
			return fmt.Sprintf("%s/s_%s.md", StableDir, object), nil
		case "flow":
			return fmt.Sprintf("%s/s_flow_%s.md", StableFlowDir, object), nil
		case "project":
			return fmt.Sprintf("%s/s_project.md", StableProjectDir), nil
		}
	}
	return "", fmt.Errorf("unsupported object/layer combination %q/%q", objectType, layer)
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
