package ui

import (
	"fmt"

	"github.com/rivo/tview"
)

// Header renders the ec2s banner and a live summary line.
type Header struct {
	view *tview.TextView
}

func newHeader() *Header {
	view := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	return &Header{view: view}
}

// SetSummary updates the header's live counts. failedAccounts is the number
// of account/region environments that failed to load in the last fetch.
func (h *Header) SetSummary(accounts, regions, instances, failedAccounts int) {
	summary := fmt.Sprintf("[::b]ec2s[-:-:-] — Easily manage AWS EC2 resources in the terminal\n%d accounts · %d regions · %d instances",
		accounts, regions, instances)
	if failedAccounts > 0 {
		summary += fmt.Sprintf(" · [orange]%d accounts failed[-]", failedAccounts)
	}
	h.view.SetText(summary)
}
