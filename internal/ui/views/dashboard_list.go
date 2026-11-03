package views

import (
	"errors"
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/component"
	"github.com/BlackGiack/kankuro/internal/ui/components"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/styles"
)

const (
	dashboardListBorderInset   = 2
	dashboardListInset         = 4
	dashboardListVerticalInset = 7
	dashboardListViewportInset = 4
	dashboardListInputInset    = 8
)

type dashboardItem struct{ dashboard config.Dashboard }

func (i dashboardItem) Title() string { return components.DisplayText(i.dashboard.Name) }
func (i dashboardItem) Description() string {
	description := components.DisplayText(i.dashboard.Description)
	if i.dashboard.State == config.DashboardDeleting {
		return "Deleting — Delete to retry. " + description
	}
	return description
}
func (i dashboardItem) FilterValue() string { return i.dashboard.Name + " " + i.dashboard.Description }

// DashboardListCallbacks dispatches commands without exposing coordinator state.
type DashboardListCallbacks struct {
	Open            func(config.Dashboard) tea.Cmd
	Save            func(id, name, description string) tea.Cmd
	PreflightDelete func(id string) tea.Cmd
	ConfirmDelete   func(id, token string, foreignConfirmed bool) tea.Cmd
	RetryDelete     func(id string) tea.Cmd
	InspectDelete   func(id string) tea.Cmd
}

type dashboardListMode int

const (
	dashboardListModeList dashboardListMode = iota
	dashboardListModeForm
	dashboardListModePreparing
	dashboardListModeConfirm
	dashboardListModeForeign
	dashboardListModeStatus
)

// DashboardList is the human entry screen for selecting and managing dashboards.
type DashboardList struct {
	component.Base

	list            list.Model
	inputs          [2]textinput.Model
	focus           int
	editingID       string
	mode            dashboardListMode
	pending         messages.DashboardDeletionPreparedMsg
	preparingStatus bool
	viewport        viewport.Model
	err             error
	callbacks       DashboardListCallbacks
}

func NewDashboardList(callbacks DashboardListCallbacks) *DashboardList {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.Foreground(styles.PrimaryColor)
	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.Foreground(styles.Gray)
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(styles.HighlightColor).BorderForeground(styles.PrimaryColor).UnsetBackground()
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(styles.HighlightColor).BorderForeground(styles.PrimaryColor).UnsetBackground()
	delegate.Styles.DimmedTitle = delegate.Styles.DimmedTitle.Foreground(styles.Gray)
	delegate.Styles.DimmedDesc = delegate.Styles.DimmedDesc.Foreground(styles.Gray)
	l := list.New(nil, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()
	v := &DashboardList{list: l, callbacks: callbacks, viewport: viewport.New()}
	v.inputs[0], v.inputs[1] = textinput.New(), textinput.New()
	v.inputs[0].Placeholder, v.inputs[1].Placeholder = "Name (required)", "Description"
	for index := range v.inputs {
		inputStyles := v.inputs[index].Styles()
		inputStyles.Focused.Prompt = inputStyles.Focused.Prompt.Foreground(styles.HighlightColor)
		inputStyles.Blurred.Prompt = inputStyles.Blurred.Prompt.Foreground(styles.Gray)
		inputStyles.Focused.Placeholder = styles.HelpStyle
		inputStyles.Blurred.Placeholder = styles.HelpStyle
		inputStyles.Cursor.Color = styles.HighlightColor
		v.inputs[index].SetStyles(inputStyles)
	}
	return v
}

// SetDashboards preserves selection by immutable ID when the inventory refreshes.
func (v *DashboardList) SetDashboards(dashboards []config.Dashboard) {
	selected, _ := v.selected()
	items := make([]list.Item, 0, len(dashboards))
	selection := 0
	for index, dashboard := range dashboards {
		items = append(items, dashboardItem{dashboard: dashboard})
		if dashboard.ID == selected.ID {
			selection = index
		}
	}
	v.list.SetItems(items)
	if len(items) > 0 {
		v.list.Select(selection)
	}
}

func (v *DashboardList) selected() (config.Dashboard, bool) {
	item, ok := v.list.SelectedItem().(dashboardItem)
	return item.dashboard, ok
}

func (v *DashboardList) Init() tea.Cmd { return disableKittyProtocolCmd }

func (v *DashboardList) SetSize(width, height int) {
	v.Base.SetSize(width, height)
	v.list.SetSize(max(1, width-dashboardListInset), max(1, height-dashboardListVerticalInset))
	v.viewport.SetWidth(max(1, width-dashboardListViewportInset))
	v.viewport.SetHeight(max(1, height-dashboardListVerticalInset))
	for index := range v.inputs {
		v.inputs[index].SetWidth(max(1, width-dashboardListInputInset))
	}
}

//nolint:gocognit,gocyclo,cyclop,funlen // All modal states are handled in this state machine.
func (v *DashboardList) Update(msg tea.Msg) (component.Model, tea.Cmd) {
	if status, ok := msg.(messages.DashboardDeletionStatusMsg); ok {
		if v.mode != dashboardListModePreparing || !v.preparingStatus || status.DashboardID != v.pending.DashboardID {
			return v, nil
		}
		if status.Err != nil {
			v.err = status.Err
			v.mode = dashboardListModeList
			return v, nil
		}
		v.err = nil
		lines := make([]string, 0, len(status.Summary))
		for _, line := range status.Summary {
			lines = append(lines, components.DisplayText(line))
		}
		v.viewport.SetContent(strings.Join(lines, "\n"))
		v.mode = dashboardListModeStatus
		return v, nil
	}
	if prepared, ok := msg.(messages.DashboardDeletionPreparedMsg); ok {
		if v.mode != dashboardListModePreparing || v.preparingStatus || prepared.DashboardID != v.pending.DashboardID {
			return v, nil
		}
		if prepared.Err != nil {
			v.err = prepared.Err
			v.mode = dashboardListModeList
			return v, nil
		}
		v.err = nil
		v.pending = prepared
		lines := make([]string, 0, len(prepared.Summary)+1)
		for _, line := range prepared.Summary {
			lines = append(lines, components.DisplayText(line))
		}
		v.viewport.SetContent(strings.Join(lines, "\n"))
		v.mode = dashboardListModeConfirm
		return v, nil
	}
	if v.mode == dashboardListModeForm {
		return v.updateForm(msg)
	}
	if v.mode != dashboardListModeList {
		return v.updateConfirmation(msg)
	}
	key, ok := msg.(tea.KeyPressMsg)
	if ok {
		dashboard, selected := v.selected()
		switch {
		case key.Code == tea.KeyEnter && selected && dashboard.State == config.DashboardActive:
			if v.callbacks.Open != nil {
				return v, v.callbacks.Open(dashboard)
			}
		case key.Code == tea.KeyEnter && selected && dashboard.State == config.DashboardDeleting:
			if v.callbacks.InspectDelete != nil {
				v.err = nil
				v.pending = messages.DashboardDeletionPreparedMsg{DashboardID: dashboard.ID}
				v.mode = dashboardListModePreparing
				v.preparingStatus = true
				return v, v.callbacks.InspectDelete(dashboard.ID)
			}
		case key.Text == "n":
			v.openForm(config.Dashboard{})
		case key.Text == "e" && selected && dashboard.State == config.DashboardActive:
			v.openForm(dashboard)
		case (key.Code == tea.KeyDelete || key.Code == tea.KeyBackspace) && selected && dashboard.State == config.DashboardActive:
			if v.callbacks.PreflightDelete != nil {
				v.err = nil
				v.pending = messages.DashboardDeletionPreparedMsg{DashboardID: dashboard.ID}
				v.mode = dashboardListModePreparing
				v.preparingStatus = false
				return v, v.callbacks.PreflightDelete(dashboard.ID)
			}
		case (key.Code == tea.KeyDelete || key.Code == tea.KeyBackspace) && selected && dashboard.State == config.DashboardDeleting:
			return v, v.retryDeletion(dashboard.ID)
		case key.Text == "q":
			return v, tea.Quit
		}
	}
	var cmd tea.Cmd
	v.list, cmd = v.list.Update(msg)
	return v, cmd
}

func (v *DashboardList) openForm(dashboard config.Dashboard) {
	v.mode, v.editingID, v.focus, v.err = dashboardListModeForm, dashboard.ID, 0, nil
	v.inputs[0].SetValue(dashboard.Name)
	v.inputs[1].SetValue(dashboard.Description)
	v.inputs[0].Focus()
	v.inputs[1].Blur()
}

func (v *DashboardList) updateForm(msg tea.Msg) (component.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.Code {
		case tea.KeyEsc:
			v.mode, v.err = dashboardListModeList, nil
			return v, nil
		case tea.KeyTab:
			v.inputs[v.focus].Blur()
			v.focus = (v.focus + 1) % len(v.inputs)
			v.inputs[v.focus].Focus()
			return v, nil
		case tea.KeyEnter:
			name := strings.TrimSpace(v.inputs[0].Value())
			if name == "" {
				v.err = errors.New("name is required")
				return v, nil
			}
			id, description := v.editingID, strings.TrimSpace(v.inputs[1].Value())
			v.mode, v.err = dashboardListModeList, nil
			if v.callbacks.Save != nil {
				return v, v.callbacks.Save(id, name, description)
			}
			return v, nil
		}
	}
	var cmd tea.Cmd
	v.inputs[v.focus], cmd = v.inputs[v.focus].Update(msg)
	return v, cmd
}

func (v *DashboardList) updateConfirmation(msg tea.Msg) (component.Model, tea.Cmd) {
	//nolint:nestif // Dispatch the scoped refresh or confirmation state in order.
	if key, ok := msg.(tea.KeyPressMsg); ok {
		if key.Code == tea.KeyEsc {
			v.mode, v.pending = dashboardListModeList, messages.DashboardDeletionPreparedMsg{}
			return v, nil
		}
		if (key.Code == tea.KeyDelete || key.Code == tea.KeyBackspace) && v.mode == dashboardListModeStatus {
			return v, v.retryDeletion(v.pending.DashboardID)
		}
		if key.Code == tea.KeyEnter && v.mode == dashboardListModeConfirm {
			if v.pending.RequiresForeignConfirmation {
				v.mode = dashboardListModeForeign
				return v, nil
			}
			return v.confirmDeletion(false)
		}
		if key.Code == tea.KeyEnter && v.mode == dashboardListModeForeign {
			return v.confirmDeletion(true)
		}
	}
	if v.mode == dashboardListModeConfirm || v.mode == dashboardListModeStatus {
		var cmd tea.Cmd
		v.viewport, cmd = v.viewport.Update(msg)
		return v, cmd
	}
	return v, nil
}

func (v *DashboardList) confirmDeletion(foreign bool) (component.Model, tea.Cmd) {
	pending := v.pending
	v.mode, v.pending = dashboardListModeList, messages.DashboardDeletionPreparedMsg{}
	if v.callbacks.ConfirmDelete != nil {
		return v, v.callbacks.ConfirmDelete(pending.DashboardID, pending.Token, foreign)
	}
	return v, nil
}

func (v *DashboardList) retryDeletion(id string) tea.Cmd {
	v.mode, v.pending, v.err = dashboardListModeList, messages.DashboardDeletionPreparedMsg{}, nil
	if v.callbacks.RetryDelete != nil {
		return v.callbacks.RetryDelete(id)
	}
	return nil
}

func (v *DashboardList) View() string {
	width := max(1, v.Width)
	if v.mode == dashboardListModeForm {
		body := styles.LabelStyle.Render("Name") + "\n" + v.inputs[0].View() +
			"\n\n" + styles.LabelStyle.Render("Description") + "\n" + v.inputs[1].View()
		if v.err != nil {
			body += "\n" + styles.ErrorStyle.Render(components.DisplayText(v.err.Error()))
		}
		return v.renderPage("Dashboard details", body, "Tab: next  Enter: save  Esc: cancel")
	}
	if v.mode == dashboardListModePreparing {
		return v.renderPage("Preparing deletion impact…", "", "Esc: cancel")
	}
	if v.mode == dashboardListModeStatus {
		return v.renderPage("Dashboard deletion progress", v.viewport.View(), "Delete: retry  Esc: back  ↑/↓: scroll")
	}
	if v.mode == dashboardListModeForeign {
		return v.renderPage("Foreign sessions reference owned cards.",
			"Owned shared cards will be deleted globally; foreign sessions survive.",
			"Enter: confirm global deletion  Esc: cancel")
	}
	if v.mode == dashboardListModeConfirm {
		return v.renderPage("Delete dashboard and its owned sessions and cards?", v.viewport.View(),
			"Enter: continue  Esc: cancel  ↑/↓: scroll")
	}
	body := ""
	if v.err != nil {
		message := components.TruncateHead("Error: "+components.DisplayText(v.err.Error()), max(1, width-dashboardListInset))
		body = styles.ErrorStyle.Render(message) + "\n"
	}
	if _, selected := v.selected(); !selected {
		return v.renderPage("Dashboards", body+styles.HelpStyle.Render("No dashboards. Press n to create one."), "n: new  q: quit")
	}
	dashboard, _ := v.selected()
	detail := components.TruncateTail("ID: "+components.DisplayText(dashboard.ID), max(1, width-dashboardListInset))
	items := v.list
	if v.err != nil {
		items.SetHeight(max(1, items.Height()-1))
	}
	help := "Enter: open  n: new  e: edit  Delete: remove  q: quit"
	if dashboard.State == config.DashboardDeleting {
		help = "Enter: details  Delete: retry  n: new  q: quit"
	}
	return v.renderPage("Dashboards", body+items.View()+"\n"+styles.HelpStyle.Render(detail), help)
}

func (v *DashboardList) renderPage(title, body, help string) string {
	header := lipgloss.NewStyle().Foreground(styles.White).Background(styles.PrimaryColor).
		Padding(0, cardsListPadding).MarginBottom(1).Bold(true).
		Render(components.TruncateTail("KANKURO - "+title, max(1, v.Width-cardsListPadding*2)))
	paneHeight := max(1, v.Height-dashboardHeaderRows-dashboardFooterRows)
	pane := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).BorderForeground(styles.PrimaryColor).
		Padding(0, 1).Width(max(1, v.Width-dashboardListBorderInset)).
		Height(paneHeight).MaxHeight(paneHeight).Render(body)
	content := header + "\n" + pane + "\n" + styles.HelpStyle.Render(help)
	return lipgloss.NewStyle().MaxWidth(max(1, v.Width)).MaxHeight(max(1, v.Height)).Render(content)
}
