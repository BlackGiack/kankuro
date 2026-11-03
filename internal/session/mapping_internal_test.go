package session

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

// TestDataRoundTripCopiesEveryField checks values through data, restore and snapshot.
// Name-presence checks alone cannot detect a field omitted by a mapper.
func TestDataRoundTripCopiesEveryField(t *testing.T) {
	dataType := reflect.TypeFor[Data]()

	src := &Session{}
	fillNonZero(t, reflect.ValueOf(src).Elem(), dataType)

	t.Run("data_then_restore", func(t *testing.T) {
		got := src.data().restore()
		// Status is deliberately excluded: restore() forces Terminated
		// because a process cannot survive a restart, so it never round-trips.
		assertFieldsMatch(t, dataType, reflect.ValueOf(src).Elem(), reflect.ValueOf(got).Elem(),
			map[string]bool{"Status": true, "SchemaVersion": true})
	})

	t.Run("snapshot", func(t *testing.T) {
		got := src.snapshot()
		assertFieldsMatch(t, dataType, reflect.ValueOf(src).Elem(), reflect.ValueOf(&got).Elem(), map[string]bool{"SchemaVersion": true})
	})
}

// fillNonZero sets every field on dst that is also present (by name) on
// fieldsFrom to a distinct non-zero value. Non-zero values are the
// load-bearing part of this test: comparing zero values would pass whether
// or not a mapping function actually copies the field.
func fillNonZero(t *testing.T, dst reflect.Value, fieldsFrom reflect.Type) {
	t.Helper()
	i := 0
	for field := range fieldsFrom.Fields() {
		if field.Name == "SchemaVersion" { // Disk format metadata, not session state.
			continue
		}
		i++
		target := dst.FieldByName(field.Name)
		if !target.IsValid() || !target.CanSet() {
			t.Fatalf("field %s not found or not settable on %s", field.Name, dst.Type())
		}
		setNonZero(t, target, i)
	}
}

func setNonZero(t *testing.T, v reflect.Value, seed int) {
	t.Helper()
	switch v.Interface().(type) {
	case string:
		v.SetString(fmt.Sprintf("value-%d", seed))
	case int:
		v.SetInt(int64(seed))
	case bool:
		v.SetBool(true)
	case time.Time:
		v.Set(reflect.ValueOf(time.Date(2000, 1, seed%28+1, 0, 0, 0, 0, time.UTC)))
	case map[string][]ConversationBinding:
		v.Set(reflect.ValueOf(map[string][]ConversationBinding{"card": {{ID: fmt.Sprintf("id-%d", seed), Name: "name", Driver: "driver"}}}))
	case map[string]bool:
		v.Set(reflect.ValueOf(map[string]bool{"AGENTS.md": true, "CLAUDE.md": seed%2 == 0}))
	case []SubWorkspaceDescriptor:
		v.Set(reflect.ValueOf([]SubWorkspaceDescriptor{{WorkspaceID: "workspace", SourcePath: "/source", DestinationPath: "/destination"}}))
	case Status:
		v.SetString(fmt.Sprintf("status-%d", seed))
	default:
		t.Fatalf("setNonZero: unsupported field type %s; extend this test helper", v.Type())
	}
}

// assertFieldsMatch compares every field named on fieldsType between want
// and got, skipping names present in skip.
func assertFieldsMatch(t *testing.T, fieldsType reflect.Type, want, got reflect.Value, skip map[string]bool) {
	t.Helper()
	for field := range fieldsType.Fields() {
		if skip[field.Name] {
			continue
		}
		wantVal := want.FieldByName(field.Name).Interface()
		gotVal := got.FieldByName(field.Name).Interface()
		if !reflect.DeepEqual(wantVal, gotVal) {
			t.Errorf("field %s did not survive the mapping: got %v want %v; "+
				"check the mapping functions next to the type declarations", field.Name, gotVal, wantVal)
		}
	}
}
