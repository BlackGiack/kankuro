package api

// persistedDashboardScope derives scope from saved session membership and the
// current dashboard/card inventory.
type persistedDashboardScope struct{ service *Service }

func (s persistedDashboardScope) inventory() (*membershipInventory, error) {
	return s.service.loadMembershipInventory(s.service.sessions.List())
}

func (s persistedDashboardScope) validateScopeInventory() error {
	_, err := s.inventory()
	return err
}

func (s persistedDashboardScope) DashboardForSession(id string) string {
	inventory, err := s.inventory()
	if err != nil {
		return ""
	}
	item, exists := inventory.sessions[id]
	if !exists || !inventory.activeDashboard(item.DashboardID) {
		return ""
	}
	return item.DashboardID
}

func (s persistedDashboardScope) AllowsSession(callerID, targetID string) bool {
	inventory, err := s.inventory()
	if err != nil {
		return false
	}
	caller, exists := inventory.sessions[callerID]
	return exists && inventory.allowsSession(caller.DashboardID, targetID)
}

func (s persistedDashboardScope) AllowsCard(callerID, cardID string) bool {
	inventory, err := s.inventory()
	if err != nil {
		return false
	}
	caller, exists := inventory.sessions[callerID]
	return exists && inventory.allowsCard(caller.DashboardID, cardID)
}

func (s persistedDashboardScope) SessionDashboard(id string) (string, bool) {
	item, exists := s.service.sessions.GetSession(id)
	return item.DashboardID, exists
}
