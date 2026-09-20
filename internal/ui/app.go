// Package ui implements the ec2s tview terminal UI: an aggregated,
// multi-account instances table with filtering and account scoping.
package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/PYTHON01100100/ec2s/internal/awsclient"
)

const (
	pageMain     = "main"
	pageFilter   = "filter"
	pageAccounts = "accounts"
	pageHelp     = "help"
)

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

	// onRefresh is invoked on the UI goroutine when the user presses
	// Ctrl-R. It must not block — the caller is expected to launch its own
	// goroutine and report results back via SetInstancesAsync.
	onRefresh func()
}

// New builds the ec2s UI shell. accountNames are the configured account
// names, used to populate the Ctrl-A account selector.
func New(accountNames []string, onRefresh func()) *App {
	a := &App{
		tapp:         tview.NewApplication(),
		pages:        tview.NewPages(),
		header:       newHeader(),
		table:        newTable(),
		footer:       newFooter(),
		accountNames: accountNames,
		onRefresh:    onRefresh,
	}

	a.root = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.header.view, 2, 0, false).
		AddItem(a.table.view, 0, 1, true).
		AddItem(a.footer.view, 3, 0, false)

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
		a.applyFilters()
		a.header.SetSummary(totalAccounts, totalRegions, len(a.all), len(warnings))
		a.refreshFooter()
	})
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
	a.table.SetInstances(filtered)
}

func (a *App) refreshFooter() {
	a.footer.SetStatus(a.textFilter, a.accountFilter, a.warnings)
}

func (a *App) handleKey(event *tcell.EventKey) *tcell.EventKey {
	// Overlay pages (filter/accounts/help) handle their own keys; only
	// intercept global keys while the main page is on top.
	if name, _ := a.pages.GetFrontPage(); name != pageMain {
		return event
	}

	switch event.Key() {
	case tcell.KeyCtrlC:
		a.tapp.Stop()
		return nil
	case tcell.KeyCtrlR:
		if a.onRefresh != nil {
			a.onRefresh()
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
	case 'q':
		a.tapp.Stop()
		return nil
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
	}

	return event
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
