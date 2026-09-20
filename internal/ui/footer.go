package ui

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"
)

const keyHints = "<?> help  </> filter  <ctrl-a> accounts  <ctrl-r> refresh  <q> quit"

// Footer renders the status bar: active filter, account scope, key hints,
// and any per-account warnings from the last fetch.
type Footer struct {
	view *tview.TextView
}

func newFooter() *Footer {
	view := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	return &Footer{view: view}
}

// SetStatus updates the footer. accountFilter is "" for "All accounts".
// warnings lists per-environment errors from the last fetch, if any.
func (f *Footer) SetStatus(textFilter, accountFilter string, warnings []string) {
	var b strings.Builder

	scope := "All accounts"
	if accountFilter != "" {
		scope = accountFilter
	}
	fmt.Fprintf(&b, "[::b]Scope:[-:-:-] %s", scope)

	if textFilter != "" {
		fmt.Fprintf(&b, "  [::b]Filter:[-:-:-] %s", tview.Escape(textFilter))
	}

	b.WriteString("\n" + keyHints)

	if len(warnings) > 0 {
		fmt.Fprintf(&b, "\n[orange]%s[-]", strings.Join(warnings, "; "))
	}

	f.view.SetText(b.String())
}
