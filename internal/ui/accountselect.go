package ui

import "github.com/rivo/tview"

const allAccountsLabel = "All accounts"

// newAccountSelectList builds the Ctrl-A account-filter overlay. accounts is
// the configured account names; onSelect receives "" for "All accounts" or
// the chosen account name. This filters the already-fetched instances
// client-side — it does not trigger a re-fetch.
func newAccountSelectList(accounts []string, onSelect func(account string), onCancel func()) *tview.List {
	list := tview.NewList().
		ShowSecondaryText(false)

	list.AddItem(allAccountsLabel, "", 0, func() {
		onSelect("")
	})
	for _, name := range accounts {
		account := name
		list.AddItem(account, "", 0, func() {
			onSelect(account)
		})
	}

	list.SetDoneFunc(onCancel)

	return list
}
