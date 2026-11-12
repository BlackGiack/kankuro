package app_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/app"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

// Fail at the repository mutation boundary, without relying on chmod behavior
// (which differs for root). Reads and subsequent retries use the real repository.
type mcpWriteFailureRepository struct {
	config.Repository

	saveErr   error
	deleteErr error
}

func (r *mcpWriteFailureRepository) UpsertMCP(server config.MCPServer) (config.MCPServer, error) {
	if r.saveErr != nil {
		return config.MCPServer{}, r.saveErr
	}
	return r.Repository.UpsertMCP(server)
}

func (r *mcpWriteFailureRepository) DeleteMCP(id string) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}
	return r.Repository.DeleteMCP(id)
}

type mcpGapFixture struct {
	coord  *app.Coordinator
	repo   *config.FileRepository
	faults *mcpWriteFailureRepository
	other  config.Dashboard
	server config.MCPServer
}

func newMCPGapFixture(t *testing.T) *mcpGapFixture {
	t.Helper()
	repo, err := newCurrentRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	server, err := repo.UpsertMCP(config.MCPServer{Name: "Alpha", Command: "server", Description: "preserve me"})
	if err != nil {
		t.Fatal(err)
	}
	other, err := repo.CreateDashboard(config.Dashboard{Name: "Other"})
	if err != nil {
		t.Fatal(err)
	}
	faults := &mcpWriteFailureRepository{Repository: repo}
	coord := newOpenedCoordinator(faults, newTestManager(t, repo.Dir(), repo))
	coord.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	drive(t, coord, coord.RefreshData())
	f := &mcpGapFixture{coord: coord, repo: repo, faults: faults, other: other, server: server}
	f.open(t)
	return f
}

func (f *mcpGapFixture) key(t *testing.T, code rune) tea.Cmd {
	t.Helper()
	_, cmd := f.coord.Update(tea.KeyPressMsg{Code: code})
	return cmd
}

func (f *mcpGapFixture) open(t *testing.T) {
	t.Helper()
	_, cmd := f.coord.Update(tea.KeyPressMsg{Text: "m"})
	drive(t, f.coord, cmd)
	if f.coord.CurrentView() != app.ViewMCPManager {
		t.Fatal("MCP manager did not open")
	}
}

func (f *mcpGapFixture) switchDashboard(t *testing.T) {
	t.Helper()
	current := f.coord.RefreshData()().(messages.DataLoadedMsg)
	f.coord.Update(messages.DashboardListRequestedMsg{DashboardID: current.DashboardID, Generation: current.Generation})
	_, cmd := f.coord.Update(messages.DashboardOpenRequestedMsg{Dashboard: f.other, Generation: current.Generation + 1})
	drive(t, f.coord, cmd)
	f.open(t)
}

func (f *mcpGapFixture) confirmDelete(t *testing.T) tea.Cmd {
	t.Helper()
	f.key(t, tea.KeyDelete)
	cmd := f.key(t, tea.KeyEnter)
	if cmd == nil {
		t.Fatal("confirmation produced no request")
	}
	return cmd
}

func (f *mcpGapFixture) registryBytes(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(f.repo.Dir(), config.MCPsFilename))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func mcpDiskFailure(repo *config.FileRepository) error {
	return &os.PathError{Op: "rename", Path: filepath.Join(repo.Dir(), config.MCPsFilename), Err: syscall.ENOSPC}
}

func TestMCPManagerCrossDashboardRequests(t *testing.T) {
	for _, action := range []string{"edit", "delete"} {
		t.Run(action, func(t *testing.T) {
			checkMCPManagerCrossDashboardRequests(t, action)
		})
	}
}

func checkMCPManagerCrossDashboardRequests(t *testing.T, action string) {
	t.Helper()
	f := newMCPGapFixture(t)
	before := f.registryBytes(t)
	held := f.key(t, tea.KeyEnter)
	if action == "delete" {
		held = f.confirmDelete(t)
	}
	f.switchDashboard(t)
	fresh := f.confirmDelete(t)
	_, discarded := f.coord.Update(held())
	if discarded != nil || f.coord.CurrentView() != app.ViewMCPManager {
		t.Fatal("old dashboard request was admitted or navigated the new dashboard")
	}
	if !bytes.Equal(before, f.registryBytes(t)) {
		t.Fatal("discarded request mutated the registry")
	}
	// Discarding the old request must not clear the new visit's pending deletion.
	_, mutation := f.coord.Update(fresh())
	if mutation == nil {
		t.Fatal("stale request cleared the new dashboard's pending confirmation")
	}
	result := mutation().(messages.MCPManagerDeletedMsg)
	if result.Err != nil || result.DashboardID != f.other.ID {
		t.Fatalf("fresh delete: %+v", result)
	}
	_, refresh := f.coord.Update(result)
	drive(t, f.coord, refresh)
	if strings.Contains(f.coord.View(), "Alpha") {
		t.Fatal("successful deletion was not reflected in current dashboard registry")
	}
}

func TestMCPManagerCrossDashboardSaveResults(t *testing.T) {
	for _, failed := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "failure"}[failed], func(t *testing.T) {
			checkMCPManagerCrossDashboardSaveResults(t, failed)
		})
	}
}

func checkMCPManagerCrossDashboardSaveResults(t *testing.T, failed bool) {
	t.Helper()
	f := newMCPGapFixture(t)
	drive(t, f.coord, f.key(t, tea.KeyEnter))
	f.coord.Views()[app.ViewMCPForm].(*views.MCPForm).Inputs()[0].SetValue("Saved from first dashboard")
	held := f.key(t, tea.KeyEnter)
	if failed {
		f.faults.saveErr = mcpDiskFailure(f.repo)
	}
	f.switchDashboard(t)
	drive(t, f.coord, f.key(t, tea.KeyEnter))
	form := f.coord.Views()[app.ViewMCPForm].(*views.MCPForm)
	form.Inputs()[0].SetValue("Other dashboard draft")
	marker := errors.New("current dashboard error")
	f.coord.Update(messages.ErrorMsg{Err: marker})
	result := held().(messages.MCPSavedMsg)
	_, refresh := f.coord.Update(result)
	if failed && !errors.Is(result.Err, syscall.ENOSPC) {
		t.Fatalf("missing original write error: %v", result.Err)
	}
	if !failed {
		if result.Err != nil || refresh == nil {
			t.Fatalf("successful stale save must refresh global data: %v", result.Err)
		}
		drive(t, f.coord, refresh)
	}
	if f.coord.CurrentView() != app.ViewMCPForm || form.Inputs()[0].Value() != "Other dashboard draft" ||
		!errors.Is(f.coord.Err(), marker) {
		t.Fatal("old save result overwrote current form, error or navigation")
	}
	want := f.server.Name
	if !failed {
		want = "Saved from first dashboard"
	}
	data := f.coord.RefreshData()().(messages.DataLoadedMsg)
	if data.DashboardID != f.other.ID || data.MCPs[0].Name != want {
		t.Fatalf("global save or current refresh scope incorrect: %+v", data)
	}
}

func TestMCPManagerCrossDashboardDeleteResults(t *testing.T) {
	for _, failed := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "failure"}[failed], func(t *testing.T) {
			checkMCPManagerCrossDashboardDeleteResults(t, failed)
		})
	}
}

func checkMCPManagerCrossDashboardDeleteResults(t *testing.T, failed bool) {
	t.Helper()
	f := newMCPGapFixture(t)
	_, held := f.coord.Update(f.confirmDelete(t)())
	if failed {
		f.faults.deleteErr = mcpDiskFailure(f.repo)
	}
	f.switchDashboard(t)
	marker := errors.New("current dashboard error")
	f.coord.Update(messages.ErrorMsg{Err: marker})
	result := held().(messages.MCPManagerDeletedMsg)
	_, refresh := f.coord.Update(result)
	if failed && !errors.Is(result.Err, syscall.ENOSPC) {
		t.Fatalf("missing original write error: %v", result.Err)
	}
	if !failed {
		if result.Err != nil || refresh == nil {
			t.Fatalf("successful stale deletion must refresh global data: %v", result.Err)
		}
		drive(t, f.coord, refresh)
	}
	if f.coord.CurrentView() != app.ViewMCPManager || !errors.Is(f.coord.Err(), marker) {
		t.Fatal("old delete result overwrote current error or navigation")
	}
	if !failed && strings.Contains(f.coord.View(), "Alpha") {
		t.Fatal("successful deletion did not refresh the new dashboard's manager")
	}
	if failed {
		f.faults.deleteErr = nil
		_, retry := f.coord.Update(f.confirmDelete(t)())
		if retry == nil || retry().(messages.MCPManagerDeletedMsg).Err != nil {
			t.Fatal("old failed deletion stranded pending state across dashboards")
		}
	}
}

func TestMCPManagerWriteFailuresRetainDataAndAllowRetry(t *testing.T) {
	for _, action := range []string{"save", "delete"} {
		t.Run(action, func(t *testing.T) {
			checkMCPManagerWriteFailuresRetainDataAndAllowRetry(t, action)
		})
	}
}

func checkMCPManagerWriteFailuresRetainDataAndAllowRetry(t *testing.T, action string) {
	t.Helper()
	f := newMCPGapFixture(t)
	before := f.registryBytes(t)
	failure := mcpDiskFailure(f.repo)
	var result tea.Msg
	if action == "save" {
		drive(t, f.coord, f.key(t, tea.KeyEnter))
		f.coord.Views()[app.ViewMCPForm].(*views.MCPForm).Inputs()[0].SetValue("Retry draft")
		f.faults.saveErr = failure
		result = f.key(t, tea.KeyEnter)()
	} else {
		f.faults.deleteErr = failure
		_, cmd := f.coord.Update(f.confirmDelete(t)())
		result = cmd()
	}
	f.coord.Update(result)
	assertBannerShown(t, f.coord, failure, "MCP write failure")
	if !bytes.Equal(before, f.registryBytes(t)) {
		t.Fatal("failed write changed persisted MCP definitions")
	}
	if action == "save" {
		if f.coord.CurrentView() != app.ViewMCPForm ||
			f.coord.Views()[app.ViewMCPForm].(*views.MCPForm).Inputs()[0].Value() != "Retry draft" {
			t.Fatal("failed save discarded edited fields")
		}
		f.faults.saveErr = nil
		result = f.key(t, tea.KeyEnter)()
	} else {
		if f.coord.CurrentView() != app.ViewMCPManager || !strings.Contains(f.coord.View(), "Alpha") {
			t.Fatal("failed deletion removed the visible entry")
		}
		f.faults.deleteErr = nil
		_, cmd := f.coord.Update(f.confirmDelete(t)())
		result = cmd()
	}
	_, refresh := f.coord.Update(result)
	drive(t, f.coord, refresh)
	if f.coord.Err() != nil || f.coord.CurrentView() != app.ViewMCPManager || bytes.Equal(before, f.registryBytes(t)) {
		t.Fatal("retry did not persist the mutation and return to the refreshed manager")
	}
}
