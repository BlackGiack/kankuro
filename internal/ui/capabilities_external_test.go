package ui_test

import (
	"github.com/BlackGiack/kankuro/internal/api"
	app "github.com/BlackGiack/kankuro/internal/ui"
)

func testCapabilities(s *api.Service) app.Capabilities {
	return app.Capabilities{Configuration: s, Center: s, Reads: s, Editor: s, Actions: s, Attachment: s, Deletion: s}
}
