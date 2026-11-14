//nolint:testpackage // Tests exercise package-private revision and service fixtures.
package api

import (
	"reflect"
	"testing"
	"time"

	"github.com/BlackGiack/kankuro/internal/config"
)

func TestCardRevisionStoredFields(t *testing.T) {
	base := config.IDCard{Env: map[string]string{"a": "1", "b": "2"}}
	original, err := cardRevision(base)
	if err != nil {
		t.Fatal(err)
	}
	reordered := base
	reordered.Env = map[string]string{"b": "2", "a": "1"}
	same, err := cardRevision(reordered)
	if err != nil || same != original {
		t.Fatalf("map order changed revision: %s %v", same, err)
	}
	typ := reflect.TypeOf(base)
	for i := range typ.NumField() {
		field := typ.Field(i)
		if field.Tag.Get("yaml") == "-" {
			continue
		}
		t.Run(field.Name, func(t *testing.T) {
			changed := base
			value := reflect.ValueOf(&changed).Elem().Field(i)
			//nolint:exhaustive // Unsupported field kinds must make the coverage test fail.
			switch value.Kind() {
			case reflect.String:
				value.SetString("changed")
			case reflect.Bool:
				value.SetBool(true)
			case reflect.Slice:
				value.Set(reflect.MakeSlice(value.Type(), 0, 0))
			case reflect.Map:
				value.Set(reflect.MakeMap(value.Type()))
			case reflect.Struct:
				value.Set(reflect.ValueOf(time.Unix(1, 0)))
			default:
				t.Fatalf("test needs mutation for %s", value.Type())
			}
			revision, revisionErr := cardRevision(changed)
			if revisionErr != nil || revision == original {
				t.Fatalf("field did not change revision: %s %v", field.Name, revisionErr)
			}
		})
	}
}

func TestCardRevisionDistinguishesNilEmptyAndListOrder(t *testing.T) {
	cases := []struct {
		name        string
		left, right config.IDCard
	}{
		{"nil-empty-slice", config.IDCard{}, config.IDCard{SkillIDs: []string{}}},
		{"nil-empty-map", config.IDCard{}, config.IDCard{Env: map[string]string{}}},
		{"list-order", config.IDCard{SkillIDs: []string{"a", "b"}}, config.IDCard{SkillIDs: []string{"b", "a"}}},
		{"system-instruction", config.IDCard{SystemInstruction: "a"}, config.IDCard{SystemInstruction: "b"}},
		{"environment-value", config.IDCard{Env: map[string]string{"k": "a"}}, config.IDCard{Env: map[string]string{"k": "b"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a, err := cardRevision(tc.left)
			if err != nil {
				t.Fatal(err)
			}
			b, err := cardRevision(tc.right)
			if err != nil {
				t.Fatal(err)
			}
			if a == b {
				t.Fatal("distinct content produced equal revisions")
			}
		})
	}
}
