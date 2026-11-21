package mcp

import (
	"encoding/json"
	"testing"
)

func TestWorkspaceIDsJSONShape(t *testing.T) {
	for _, tc := range []struct {
		name, input      string
		wantErr, wantNil bool
	}{
		{"omitted", `{}`, false, true},
		{"null", `{"workspace_ids":null}`, true, false},
		{"empty", `{"workspace_ids":[]}`, false, false},
		{"ordered", `{"workspace_ids":["main","sub"]}`, false, false},
		{"scalar", `{"workspace_ids":"main"}`, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var input createSessionInput
			err := json.Unmarshal([]byte(tc.input), &input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("unmarshal error: %v", err)
			}
			if err == nil && (input.WorkspaceIDs == nil) != tc.wantNil {
				t.Fatalf("workspace IDs: %#v", input.WorkspaceIDs)
			}
		})
	}
}
