package views

import "kankuro/internal/ui/messages"

// SourceType identifies which kind of sources are being managed.
type SourceType string

const (
	// SourceTypeSkills manages skill sources.
	SourceTypeSkills SourceType = "skills"
	// SourceTypePersonas manages persona sources.
	SourceTypePersonas SourceType = "personas"
	// SourceTypeWorkspaces manages workspace sources.
	SourceTypeWorkspaces SourceType = "workspaces"
	// SourceTypePuppets manages puppet sources.
	SourceTypePuppets SourceType = "puppets"
)

const (
	// formFieldsSimple is the number of form fields for simple sources (skills/personas).
	formFieldsSimple = 1
	// formFieldsComplex is the number of form fields for complex sources (workspaces).
	formFieldsComplex = 2
)

// SourceTypeConfig returns the configuration for a source type.
type SourceTypeConfig struct {
	Title       string
	Legend      string
	FormPrompt  string
	FormFields  int
	Field2Label string
	ReturnType  messages.ViewType
}

// GetSourceTypeConfig returns configuration for the given source type.
func GetSourceTypeConfig(st SourceType) SourceTypeConfig {
	switch st {
	case SourceTypeSkills:
		return SourceTypeConfig{
			Title:      "Manage Skill Sources",
			FormPrompt: "Path to skill source...",
			FormFields: formFieldsSimple,
			ReturnType: messages.ViewCardEditor,
		}
	case SourceTypePersonas:
		return SourceTypeConfig{
			Title:      "Manage Persona Sources",
			FormPrompt: "Path to persona source...",
			FormFields: formFieldsSimple,
			ReturnType: messages.ViewCardEditor,
		}
	case SourceTypePuppets:
		return SourceTypeConfig{
			Title:      "Manage Puppet Sources",
			FormPrompt: "Path to puppet source...",
			FormFields: formFieldsSimple,
			ReturnType: messages.ViewCardEditor,
		}
	case SourceTypeWorkspaces:
		return SourceTypeConfig{
			Title:       "Manage Workspace Sources",
			Legend:      "(R=Root folder containing multiple workspaces)",
			FormPrompt:  "Path to workspace or root directory...",
			FormFields:  formFieldsComplex,
			Field2Label: "Is Root (y/n)",
			ReturnType:  messages.ViewDashboard,
		}
	default:
		return SourceTypeConfig{
			Title:      "Manage Sources",
			FormFields: formFieldsSimple,
			ReturnType: messages.ViewDashboard,
		}
	}
}
