package specpaths

import "fmt"

const (
	ModulesRootDir           = "docs/specs/units"
	CandidateDir             = ModulesRootDir + "/candidate"
	StableDir                = ModulesRootDir + "/stable"
	CandidateAppendixDir     = CandidateDir + "/appendix"
	StableAppendixDir        = StableDir + "/appendix"
	FlowsRootDir             = "docs/specs/scenarios"
	CandidateFlowDir         = FlowsRootDir + "/candidate"
	StableFlowDir            = FlowsRootDir + "/stable"
	RepositoryMappingFileRef = "docs/specs/repository_mapping.md"
)

func MainSpecFileRef(layer, unit string) (string, error) {
	return ObjectMainSpecFileRef("unit", layer, unit)
}

func ObjectMainSpecFileRef(objectType, layer, object string) (string, error) {
	switch layer {
	case "candidate":
		switch objectType {
		case "unit":
			return fmt.Sprintf("%s/c_unit_%s.md", CandidateDir, object), nil
		case "scenario":
			return fmt.Sprintf("%s/c_scenario_%s.md", CandidateFlowDir, object), nil
		}
	case "stable":
		switch objectType {
		case "unit":
			return fmt.Sprintf("%s/s_unit_%s.md", StableDir, object), nil
		case "scenario":
			return fmt.Sprintf("%s/s_scenario_%s.md", StableFlowDir, object), nil
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

func CandidateAppendixGlob(unit string) string {
	return fmt.Sprintf("%s/c_unit_%s_*.md", CandidateAppendixDir, unit)
}

func ObjectCandidateAppendixGlob(objectType, object string) (string, error) {
	switch objectType {
	case "unit":
		return fmt.Sprintf("%s/c_unit_%s_*.md", CandidateAppendixDir, object), nil
	case "scenario":
		return fmt.Sprintf("%s/appendix/c_scenario_%s_*.md", CandidateFlowDir, object), nil
	default:
		return "", fmt.Errorf("unsupported object type %q", objectType)
	}
}
