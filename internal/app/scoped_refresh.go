package app

import (
	"errors"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/api"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
)

func (r refresher) scopedFull(service *api.Service, dashboardID string, generation uint64, previewID string) tea.Cmd {
	return func() tea.Msg {
		data := r.loadGlobalConfigData()
		data.DashboardID, data.Generation = dashboardID, generation
		// A failed scoped read must leave prior dashboard data intact.
		data.Cards, data.Sessions = nil, nil
		if dashboardID != "" {
			scoped, err := service.ReadDashboardData(dashboardID)
			if err != nil {
				data.Err = errors.Join(data.Err, err)
			} else {
				data.Cards, data.Sessions = scoped.Cards, scoped.Sessions
				for _, item := range scoped.Sessions {
					if item.ID == previewID {
						data.PreviewSessionID, data.PreviewScrollback = previewID, r.sessions.GetScrollback(previewID)
						break
					}
				}
			}
		}
		stamps, err := r.repo.Fingerprint()
		if err != nil {
			return data
		}
		return configFingerprintMsg{stamps: stamps, data: &data}
	}
}

func (r refresher) scopedSessions(service *api.Service, dashboardID string, generation uint64, previewID string) tea.Cmd {
	return func() tea.Msg {
		result := messages.SessionsRefreshedMsg{DashboardID: dashboardID, Generation: generation}
		if dashboardID == "" {
			return result
		}
		data, err := service.ReadDashboardData(dashboardID)
		if err != nil {
			result.Err = err
			return result
		}
		result.Cards, result.Sessions = data.Cards, data.Sessions
		for _, item := range data.Sessions {
			if item.ID == previewID {
				result.PreviewSessionID, result.PreviewScrollback = previewID, r.sessions.GetScrollback(previewID)
				break
			}
		}
		return result
	}
}

func (r refresher) scopedConfigIfChanged(
	service *api.Service,
	dashboardID string,
	generation uint64,
	prev map[string]config.FileStamp,
) tea.Cmd {
	return func() tea.Msg {
		stamps, err := r.repo.Fingerprint()
		if err == nil && fingerprintsEqual(prev, stamps) {
			return configFingerprintMsg{}
		}
		data := r.loadGlobalConfigData()
		data.DashboardID, data.Generation = dashboardID, generation
		data.Cards = nil
		if dashboardID != "" {
			scoped, scopeErr := service.ReadDashboardData(dashboardID)
			if scopeErr != nil {
				data.Err = errors.Join(data.Err, scopeErr)
			} else {
				data.Cards = scoped.Cards
			}
		}
		return configFingerprintMsg{stamps: stamps, data: &data}
	}
}
