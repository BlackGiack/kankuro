package views

import "testing"

// TestMCPFieldTableMatchesConstants ties the index constants to the field
// table and the input slice: no gaps, no duplicates, in order. Adding a
// constant without a table row (or the reverse) fails here.
func TestMCPFieldTableMatchesConstants(t *testing.T) {
	table := mcpFieldTable()
	if len(table) != mcpInputCount {
		t.Fatalf("field table has %d rows, mcpInputCount is %d", len(table), mcpInputCount)
	}
	for i, field := range table {
		if field.idx != i {
			t.Errorf("table row %d (%q) has idx %d", i, field.label, field.idx)
		}
	}
	if got := len(NewMCPForm(MCPFormCallbacks{}).Inputs()); got != mcpInputCount {
		t.Errorf("form has %d inputs, mcpInputCount is %d", got, mcpInputCount)
	}
}
