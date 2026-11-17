package ui

import "github.com/BlackGiack/kankuro/internal/api"

func testCapabilities(s *api.Service) Capabilities {
	return Capabilities{Configuration: s, Center: s, Reads: s, Editor: s, Actions: s, Attachment: s, Deletion: s}
}
