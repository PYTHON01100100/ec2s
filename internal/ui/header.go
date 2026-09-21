package ui

import (
	"fmt"

	"github.com/rivo/tview"

	"github.com/PYTHON01100100/ec2s/internal/awsclient"
)

// headerKeys are the static keybinding hints shown in the info panel's right
// column, mirroring e1s's header layout.
var headerKeys = []struct{ key, description string }{
	{"/", "Filter (Esc to clear)"},
	{"ctrl-a", "Filter by account"},
	{"ctrl-r", "Refresh"},
	{"g / G", "Top / bottom"},
	{"s", "Stop instance"},
	{"D", "Terminate instance"},
	{"?", "Help"},
	{"q", "Quit"},
}

// infoItemRows is the number of fields SetInstance renders in the left
// column, used by the caller to size the info panel tall enough for both
// columns.
const infoItemRows = 13

// Header renders the e1s-style "info" panel: details of the currently
// selected instance on the left, static keybinding hints on the right. It
// updates every time the table selection changes.
type Header struct {
	view     *tview.Flex
	itemsCol *tview.Flex
}

func newHeader() *Header {
	itemsCol := tview.NewFlex().SetDirection(tview.FlexRow)

	keysCol := tview.NewFlex().SetDirection(tview.FlexRow)
	for _, k := range headerKeys {
		t := tview.NewTextView().
			SetDynamicColors(true).
			SetText(fmt.Sprintf(infoKeyFmt, k.key, k.description))
		keysCol.AddItem(t, 1, 1, false)
	}

	view := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(itemsCol, 0, 2, false).
		AddItem(keysCol, 0, 1, false)
	view.SetBorder(true).
		SetBorderPadding(0, 0, 1, 1).
		SetTitle(fmt.Sprintf(infoTitleFmt, "no selection"))

	h := &Header{view: view, itemsCol: itemsCol}
	h.SetInstance(nil)
	return h
}

// SetInstance updates the info panel to show inst's details, or a
// placeholder if inst is nil (no rows in the table).
func (h *Header) SetInstance(inst *awsclient.Instance) {
	h.itemsCol.Clear()

	if inst == nil {
		h.view.SetTitle(fmt.Sprintf(infoTitleFmt, "no selection"))
		t := tview.NewTextView().SetDynamicColors(true).SetText(" No instance selected ")
		h.itemsCol.AddItem(t, 1, 1, false)
		return
	}

	h.view.SetTitle(fmt.Sprintf(infoTitleFmt, inst.Name))

	launch := "-"
	if !inst.LaunchTime.IsZero() {
		launch = inst.LaunchTime.Local().Format("2006-01-02 15:04:05")
	}

	items := []struct{ name, value string }{
		{"Instance ID", inst.ID},
		{"State", inst.State},
		{"Type", inst.Type},
		{"OS", orDash(inst.Platform)},
		{"Account", inst.AccountName},
		{"Profile", inst.Profile},
		{"Region", inst.Region},
		{"Zone", orDash(inst.AvailabilityZone)},
		{"VPC ID", orDash(inst.VPCId)},
		{"Subnet ID", orDash(inst.SubnetId)},
		{"Public IP (external)", orDash(inst.PublicIP)},
		{"Private IP (internal)", orDash(inst.PrivateIP)},
		{"Launch Time", launch},
	}
	for _, item := range items {
		t := tview.NewTextView().
			SetDynamicColors(true).
			SetText(fmt.Sprintf(infoItemFmt, item.name, item.value))
		h.itemsCol.AddItem(t, 1, 1, false)
	}
}
