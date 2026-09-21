// Package ui implements the ec2s tview terminal UI: an aggregated,
// multi-account instances table with filtering and account scoping.
package ui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/PYTHON01100100/ec2s/internal/awsclient"
)

const (
	pageMain     = "main"
	pageFilter   = "filter"
	pageAccounts = "accounts"
	pageHelp     = "help"
	pageConfirm  = "confirm"
)

const noticeDuration = 4 * time.Second

// infoPanelHeight is the info panel's fixed height: enough rows for the
// taller of the two columns (instance fields vs. keybindings), plus the top
// and bottom border.
var infoPanelHeight = max(infoItemRows, len(headerKeys)) + 2

// App is the ec2s terminal UI.
type App struct {
	tapp  *tview.Application
	pages *tview.Pages
	root  *tview.Flex

	header *Header
	table  *Table
	footer *Footer

	accountNames  []string
	all           []awsclient.Instance
	accountFilter string
	textFilter    string
	warnings      []string
	totalAccounts int
	totalRegions  int
	loaded        bool // false until the first fetch completes

	actions Actions
}

// Actions are the operations the UI triggers in response to user input; the
// caller supplies these so the ui package stays free of AWS SDK calls.
// None of these may block — implementations are expected to launch their
// own goroutine and report results back via SetInstancesAsync / Notify.
type Actions struct {
	Refresh   func()
	Start     func(inst awsclient.Instance)
	Stop      func(inst awsclient.Instance)
	Terminate func(inst awsclient.Instance)
}

// New builds the ec2s UI shell. accountNames are the configured account
// names, used to populate the Ctrl-A account selector. version is displayed
// in the footer's app chip.
func New(accountNames []string, actions Actions, version string) *App {
	applyTheme()

	a := &App{
		tapp:         tview.NewApplication(),
		pages:        tview.NewPages(),
		header:       newHeader(),
		table:        newTable(),
		footer:       newFooter(version),
		accountNames: accountNames,
		actions:      actions,
	}
	a.table.SetOnSelect(a.header.SetInstance)

	a.root = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.header.view, infoPanelHeight, 0, false).
		AddItem(a.table.view, 0, 1, true).
		AddItem(a.footer.view, 2, 0, false)

	a.pages.AddPage(pageMain, a.root, true, true)
	a.tapp.SetRoot(a.pages, true).SetFocus(a.table.view)
	a.tapp.SetInputCapture(a.handleKey)

	a.refreshFooter()

	return a
}

// Run starts the tview event loop; it blocks until the user quits.
func (a *App) Run() error {
	return a.tapp.Run()
}

// SetInstancesAsync updates the displayed instances. It is safe to call from
// any goroutine (e.g. the discovery goroutine after a refresh) — it marshals
// the update onto the UI goroutine via QueueUpdateDraw.
func (a *App) SetInstancesAsync(instances []awsclient.Instance, warnings []string, totalAccounts, totalRegions int) {
	a.tapp.QueueUpdateDraw(func() {
		a.all = instances
		a.warnings = warnings
		a.totalAccounts = totalAccounts
		a.totalRegions = totalRegions
		a.loaded = true
		a.applyFilters()
		a.refreshFooter()
	})
}

// SetTotalsAsync sets the known account/region counts before the first
// fetch completes, so the footer can show "N accounts · M regions ·
// loading instances…" instead of reading as zero accounts configured. Safe
// to call from any goroutine.
func (a *App) SetTotalsAsync(totalAccounts, totalRegions int) {
	a.tapp.QueueUpdateDraw(func() {
		a.totalAccounts = totalAccounts
		a.totalRegions = totalRegions
		a.refreshFooter()
	})
}

// Notify shows a transient action message (e.g. the result of a stop/
// terminate) in the footer, then restores the normal status line after a
// few seconds. Safe to call from any goroutine.
func (a *App) Notify(text string, isError bool) {
	icon, color := "✅", "green"
	if isError {
		icon, color = "💥", "orange"
	}
	a.tapp.QueueUpdateDraw(func() {
		a.footer.SetNotice(fmt.Sprintf("%s [%s]%s[-]", icon, color, text))
	})

	go func() {
		time.Sleep(noticeDuration)
		a.tapp.QueueUpdateDraw(a.refreshFooter)
	}()
}

func (a *App) applyFilters() {
	var filtered []awsclient.Instance
	for _, inst := range a.all {
		if a.accountFilter != "" && inst.AccountName != a.accountFilter {
			continue
		}
		if !matchesFilter(inst, a.textFilter) {
			continue
		}
		filtered = append(filtered, inst)
	}
	scope := a.accountFilter
	if scope == "" {
		scope = "all accounts"
	}
	a.table.SetInstances(filtered, scope)
}

func (a *App) refreshFooter() {
	a.footer.SetStatus(a.textFilter, a.accountFilter, a.totalAccounts, a.totalRegions, len(a.all), len(a.warnings), a.warnings, a.loaded)
}

func (a *App) handleKey(event *tcell.EventKey) *tcell.EventKey {
	front, _ := a.pages.GetFrontPage()

	// Ctrl-C always quits, everywhere, even while typing in the filter box —
	// the standard terminal "abort" convention. Plain q quits from any page
	// too, EXCEPT the filter box, where q is a normal character to search
	// for (e.g. filtering for "web-01" would be impossible otherwise).
	if event.Key() == tcell.KeyCtrlC {
		a.tapp.Stop()
		return nil
	}
	if event.Rune() == 'q' && front != pageFilter {
		a.tapp.Stop()
		return nil
	}

	// The remaining shortcuts only make sense on the main page; overlays
	// (filter/accounts/help/confirm) handle their own keys otherwise.
	if front != pageMain {
		return event
	}

	switch event.Key() {
	case tcell.KeyCtrlR:
		if a.actions.Refresh != nil {
			a.actions.Refresh()
		}
		return nil
	case tcell.KeyCtrlA:
		a.showAccountSelect()
		return nil
	case tcell.KeyEsc:
		if a.textFilter != "" || a.accountFilter != "" {
			a.textFilter = ""
			a.accountFilter = ""
			a.applyFilters()
			a.refreshFooter()
		}
		return nil
	}

	switch event.Rune() {
	case '/':
		a.showFilter()
		return nil
	case '?':
		a.showHelp()
		return nil
	case 'g':
		a.table.SelectTop()
		return nil
	case 'G':
		a.table.SelectBottom()
		return nil
	case 's':
		a.doStart()
		return nil
	case 'S':
		a.confirmStop()
		return nil
	case 'D':
		a.confirmTerminate()
		return nil
	}

	return event
}

// doStart starts the selected instance immediately, with no confirmation —
// unlike stop/terminate, starting an instance doesn't interrupt anything or
// lose data, so asking "are you sure?" would just be friction.
func (a *App) doStart() {
	inst, ok := a.table.SelectedInstance()
	if !ok || a.actions.Start == nil {
		return
	}
	a.actions.Start(inst)
}

func (a *App) confirmStop() {
	inst, ok := a.table.SelectedInstance()
	if !ok || a.actions.Stop == nil {
		return
	}
	text := fmt.Sprintf("Stop %s\n(%s)?", inst.Name, inst.ID)
	a.showConfirm(text, func() { a.actions.Stop(inst) })
}

func (a *App) confirmTerminate() {
	inst, ok := a.table.SelectedInstance()
	if !ok || a.actions.Terminate == nil {
		return
	}
	text := fmt.Sprintf("Terminate %s\n(%s)?\nThis cannot be undone.", inst.Name, inst.ID)
	a.showConfirm(text, func() { a.actions.Terminate(inst) })
}

func (a *App) showConfirm(text string, onYes func()) {
	closeConfirm := func() {
		a.pages.RemovePage(pageConfirm)
		a.tapp.SetFocus(a.table.view)
	}

	modal := newConfirmModal(text,
		func() {
			closeConfirm()
			onYes()
		},
		closeConfirm,
	)

	a.pages.AddPage(pageConfirm, modal, true, true)
	a.tapp.SetFocus(modal)
}

func (a *App) showFilter() {
	input := newFilterInput(
		func(query string) {
			a.textFilter = query
			a.applyFilters()
			a.refreshFooter()
		},
		func() {
			a.pages.RemovePage(pageFilter)
			a.tapp.SetFocus(a.table.view)
		},
	)
	input.SetText(a.textFilter)
	input.SetBorder(true).SetTitle(" Filter ")

	a.pages.AddPage(pageFilter, centered(input, 70, 3), true, true)
	a.tapp.SetFocus(input)
}

func (a *App) showAccountSelect() {
	closeAccounts := func() {
		a.pages.RemovePage(pageAccounts)
		a.tapp.SetFocus(a.table.view)
	}

	list := newAccountSelectList(a.accountNames,
		func(account string) {
			a.accountFilter = account
			a.applyFilters()
			a.refreshFooter()
			closeAccounts()
		},
		closeAccounts,
	)
	list.SetBorder(true).SetTitle(" Filter by account ")

	height := len(a.accountNames) + 3
	a.pages.AddPage(pageAccounts, centered(list, 40, height), true, true)
	a.tapp.SetFocus(list)
}

func (a *App) showHelp() {
	view := newHelpView()

	closeHelp := func() {
		a.pages.RemovePage(pageHelp)
		a.tapp.SetFocus(a.table.view)
	}
	view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc || event.Rune() == '?' {
			closeHelp()
			return nil
		}
		return event
	})

	a.pages.AddPage(pageHelp, centered(view, 60, 20), true, true)
	a.tapp.SetFocus(view)
}

// centered wraps p in a Flex that pins it to a fixed width/height in the
// middle of the screen, the standard tview idiom for modal-style overlays.
func centered(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(p, height, 1, true).
			AddItem(nil, 0, 1, false), width, 1, true).
		AddItem(nil, 0, 1, false)
}
