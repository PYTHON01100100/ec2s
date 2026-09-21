package ui

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"
)

const keyHints = "/ filter   ctrl-a accounts   ctrl-r refresh   s start   S stop   D terminate   ? help   q quit"

// Footer renders the status bar as a row of e1s-style colored chips (scope,
// active filter, key hints, totals, app version), plus a second line for
// any per-account warnings from the last fetch.
type Footer struct {
	view *tview.Flex

	chips    *tview.Flex
	scope    *tview.TextView
	filter   *tview.TextView
	hints    *tview.TextView
	counts   *tview.TextView
	app      *tview.TextView
	warnings *tview.TextView

	appChipWidth int
}

func newFooter(version string) *Footer {
	f := &Footer{
		chips:    tview.NewFlex().SetDirection(tview.FlexColumn),
		scope:    tview.NewTextView().SetDynamicColors(true),
		filter:   tview.NewTextView().SetDynamicColors(true),
		hints:    tview.NewTextView().SetDynamicColors(true).SetText("[green]" + keyHints + "[-]"),
		counts:   tview.NewTextView().SetDynamicColors(true),
		app:      tview.NewTextView().SetDynamicColors(true),
		warnings: tview.NewTextView().SetDynamicColors(true),
	}

	appText := fmt.Sprintf("ec2s:%s", version)
	f.app.SetText(fmt.Sprintf(footerAppFmt, "ec2s", version))
	f.appChipWidth = len(appText) + 3

	f.view = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(f.chips, 1, 0, false).
		AddItem(f.warnings, 1, 0, false)

	return f
}

// SetStatus rebuilds the chip row and warnings line from current state.
// accountFilter is "" for "all accounts". totalInstances/totalAccounts/
// totalRegions reflect the last full fetch (unfiltered); warnings lists
// per-environment errors from that fetch, if any. loaded is false until the
// first fetch has completed, so a pending fetch reads as "loading" rather
// than indistinguishable from zero accounts/instances.
func (f *Footer) SetStatus(textFilter, accountFilter string, totalAccounts, totalRegions, totalInstances, failedAccounts int, warnings []string, loaded bool) {
	scope := accountFilter
	if scope == "" {
		scope = "all accounts"
	}
	f.scope.SetText(fmt.Sprintf(footerChipActiveFmt, scope))

	f.chips.Clear()
	f.chips.AddItem(f.scope, len(scope)+3, 0, false)

	if textFilter != "" {
		filterLabel := "filter: " + tview.Escape(textFilter)
		f.filter.SetText(fmt.Sprintf(footerChipFmt, filterLabel))
		f.chips.AddItem(f.filter, len(filterLabel)+3, 0, false)
	}

	f.chips.AddItem(f.hints, 0, 1, false)

	countsLabel := fmt.Sprintf("%d accounts · %d regions · loading instances…", totalAccounts, totalRegions)
	if loaded {
		countsLabel = fmt.Sprintf("%d accounts · %d regions · %d instances", totalAccounts, totalRegions, totalInstances)
		if failedAccounts > 0 {
			countsLabel += fmt.Sprintf(" · %d failed", failedAccounts)
		}
	}
	f.counts.SetText(fmt.Sprintf(footerChipFmt, countsLabel))
	f.chips.AddItem(f.counts, len(countsLabel)+3, 0, false)

	f.chips.AddItem(f.app, f.appChipWidth, 0, false)

	if len(warnings) > 0 {
		f.warnings.SetText("💥 [orange]" + strings.Join(warnings, "; ") + "[-]")
	} else {
		f.warnings.SetText("")
	}
}

// SetNotice overrides the warnings line with a transient action message
// (e.g. the result of a stop/terminate). The caller is responsible for
// restoring the normal line afterwards, typically by calling SetStatus
// again once the notice has been shown for a while.
func (f *Footer) SetNotice(text string) {
	f.warnings.SetText(text)
}
