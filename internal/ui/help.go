package ui

import "github.com/rivo/tview"

const helpText = `[::b]ec2s keybindings[-:-:-]

[::b]Navigation[-:-:-]
  j / down       move down
  k / up         move up
  g              jump to top
  G              jump to bottom

[::b]Filtering[-:-:-]
  /              open filter (plain text, or column:value e.g. state:running)
  Esc            clear filter / close overlay

[::b]Accounts[-:-:-]
  Ctrl-A         filter by configured account ("All accounts" to reset)

[::b]Actions[-:-:-]
  s              stop the selected instance (asks to confirm)
  D              terminate the selected instance (asks to confirm, irreversible)

[::b]General[-:-:-]
  Ctrl-R         refresh (re-fetch all accounts/regions)
  ?              toggle this help
  q / Ctrl-C     quit

Press Esc or ? to close this screen.`

// newHelpView builds the "?" help modal.
func newHelpView() *tview.TextView {
	view := tview.NewTextView().
		SetDynamicColors(true).
		SetText(helpText)
	view.SetBorder(true)
	return view
}
