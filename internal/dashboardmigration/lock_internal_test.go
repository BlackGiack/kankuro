package dashboardmigration

import (
	"testing"

	"github.com/BlackGiack/kankuro/internal/layout"
)

func TestExclusiveLock(t *testing.T) {
	dir := t.TempDir()
	first, err := layout.AcquireLock(dir)
	if err != nil {
		t.Fatal(err)
	}
	if second, errAt11 := layout.AcquireLock(dir); errAt11 == nil {
		_ = second.Close()
		t.Fatal("second writer acquired lock")
	}
	if errAt15 := first.Close(); errAt15 != nil {
		t.Fatal(errAt15)
	}
	second, err := layout.AcquireLock(dir)
	if err != nil {
		t.Fatal(err)
	}
	if errAt22 := second.Close(); errAt22 != nil {
		t.Fatal(errAt22)
	}
}
