package views

const (
	// Defaults.
	defaultPuppetStyle = "kankuro"
)

// Smallest terminal the UI is laid out for. Below it the coordinator shows a
// "terminal too small" screen instead of a view (impl_plan_5.md B3).
const (
	MinTerminalWidth  = 80
	MinTerminalHeight = 24
)

// CoordinatorRows is how many terminal rows the coordinator keeps back before
// sizing a view (views are sized to terminal rows - CoordinatorRows). It is the
// one definition: app.resizeAll subtracts it and the dashboard adds it back to
// fill the terminal exactly, so the two cannot disagree.
const CoordinatorRows = 4

const (
	// Common Magic Numbers.
	heightPadding = 1
	widthPadding  = 2

	dashboardSessionsWidthRatio = 4
	dashboardPreviewWidthRatio  = 2

	switchModalDelegateHeight = 4

	workspacePickerDelegateHeight = 3

	// List padding constants for workspace picker.
	workspacePickerListWidthPadding  = 4
	workspacePickerListHeightPadding = 8

	agentSwitchListHeightPadding = 10

	// Input widths.
	defaultInputWidth = 60
	smallInputWidth   = 30
	mediumInputWidth  = 40
)

const (
	// Key strings.
	keyCtrlC    = "ctrl+c"
	keyShiftTab = "shift+tab"
	keyCtrlE    = "ctrl+e"
	keyCtrlS    = "ctrl+s"
)
