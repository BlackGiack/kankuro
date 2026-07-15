package config_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"kankuro/internal/config"
)

// TestSaveSkillSourcesConcurrentWritesDoNotCorrupt documents the safety
// FileRepository's internal locking (#11) provides: concurrent calls to a
// single mutator never corrupt the underlying file or panic/deadlock the
// repository, and every individual write is durably applied. It does not
// assert all N entries survive - List-then-append-then-Save from N separate
// goroutines is a classic read-modify-write race at the *caller* level
// (each goroutine's read can miss another's concurrent write) that
// repository-internal locking alone cannot close; that is a caller-side
// concern, not this primitive's.
func TestSaveSkillSourcesConcurrentWritesDoNotCorrupt(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	const n = 10
	var wg sync.WaitGroup
	errs := make(chan error, n)
	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			existing, listErr := repo.ListSkillSources()
			if listErr != nil {
				errs <- fmt.Errorf("list: %w", listErr)
				return
			}
			if saveErr := repo.SaveSkillSources(append(existing, fmt.Sprintf("source-%d", i))); saveErr != nil {
				errs <- fmt.Errorf("save: %w", saveErr)
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("unexpected error from concurrent save: %v", err)
	}

	final, err := repo.ListSkillSources()
	if err != nil {
		t.Fatalf("final list: %v", err)
	}
	if len(final) < 1 || len(final) > n {
		t.Fatalf("expected between 1 and %d sources, got %d: %v", n, len(final), final)
	}
	seen := map[string]bool{}
	for _, s := range final {
		if seen[s] {
			t.Fatalf("duplicate entry in final list, file may be corrupted: %v", final)
		}
		seen[s] = true
	}
}

// TestUpdateSkillSourcesConcurrentAppendsAllSurvive is the strict counterpart
// to TestSaveSkillSourcesConcurrentWritesDoNotCorrupt: UpdateSkillSources
// reads, mutates, and saves the list under one lock acquisition, so unlike
// the List-then-Save pattern above, concurrent appends cannot be lost. Every
// one of the n goroutines' entries must survive.
func TestUpdateSkillSourcesConcurrentAppendsAllSurvive(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	// Use real, readable directories so the post-save registry resync
	// succeeds too - this test is about the source *list* never losing an
	// entry, not about sync errors (that's TestUpdateSkillSourcesRefreshesUIOnSyncFailure).
	const n = 10
	dirs := make([]string, n)
	for i := range n {
		dirs[i] = t.TempDir()
	}

	var wg sync.WaitGroup
	errs := make(chan error, n)
	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			updateErr := repo.UpdateSkillSources(func(sources []string) ([]string, error) {
				return append(sources, dirs[i]), nil
			})
			if updateErr != nil {
				errs <- updateErr
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("unexpected error from concurrent update: %v", err)
	}

	final, err := repo.ListSkillSources()
	if err != nil {
		t.Fatalf("final list: %v", err)
	}
	if len(final) != n {
		t.Fatalf("expected exactly %d sources (no lost updates), got %d: %v", n, len(final), final)
	}
	seen := map[string]bool{}
	for _, s := range final {
		if seen[s] {
			t.Fatalf("duplicate entry in final list: %v", final)
		}
		seen[s] = true
	}
	for _, dir := range dirs {
		if !seen[dir] {
			t.Fatalf("expected %q to survive, final list: %v", dir, final)
		}
	}
}

// TestRepositoryConcurrentMixedOperationsNoDeadlock exercises several
// domains' mutators and readers concurrently to catch the reentrancy hazard
// called out in #11: a mutator that calls a public List* method while
// already holding FileRepository.mu would deadlock the whole TUI. If this
// test does not hang, no such reentrant call exists on these paths.
func TestRepositoryConcurrentMixedOperationsNoDeadlock(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	wsDir := t.TempDir()
	if saveErr := repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wsDir}}); saveErr != nil {
		t.Fatalf("seed workspace source: %v", saveErr)
	}
	workspaces, err := repo.ListWorkspaces()
	if err != nil || len(workspaces) == 0 {
		t.Fatalf("seed workspace: %v (workspaces=%v)", err, workspaces)
	}
	wsID := workspaces[0].ID

	var wg sync.WaitGroup
	done := make(chan struct{})
	errs := make(chan error, 64)

	worker := func(fn func() error) {
		defer wg.Done()
		if workerErr := fn(); workerErr != nil {
			errs <- workerErr
		}
	}

	for i := range 5 {
		wg.Add(4)
		go worker(func() error {
			_, createErr := repo.CreateIDCard(config.IDCard{
				Name:        fmt.Sprintf("card-%d", i),
				Driver:      "cat",
				WorkspaceID: wsID,
			})
			return createErr
		})
		go worker(func() error {
			_, listErr := repo.ListIDCards()
			return listErr
		})
		go worker(func() error {
			_, upsertErr := repo.UpsertSkill(config.Skill{Name: fmt.Sprintf("skill-%d", i)})
			return upsertErr
		})
		go worker(func() error {
			_, settingsErr := repo.GetAppSettings()
			return settingsErr
		})
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("concurrent repository operations deadlocked")
	}
	close(errs)
	for err := range errs {
		t.Fatalf("unexpected error from concurrent operation: %v", err)
	}
}
