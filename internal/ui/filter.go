package ui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/PYTHON01100100/ec2s/internal/awsclient"
)

// matchesFilter reports whether an instance matches a filter query. A query
// with no ":" is a case-insensitive substring match against the instance
// name and ID. A "column:value" query matches that specific field instead
// (state, account, region, type, os, zone, vpc, subnet).
func matchesFilter(inst awsclient.Instance, query string) bool {
	query = strings.TrimSpace(query)
	if query == "" {
		return true
	}

	if column, value, ok := strings.Cut(query, ":"); ok {
		value = strings.ToLower(strings.TrimSpace(value))
		switch strings.ToLower(strings.TrimSpace(column)) {
		case "state":
			return strings.Contains(strings.ToLower(inst.State), value)
		case "account":
			return strings.Contains(strings.ToLower(inst.AccountName), value)
		case "region":
			return strings.Contains(strings.ToLower(inst.Region), value)
		case "type":
			return strings.Contains(strings.ToLower(inst.Type), value)
		case "os":
			return strings.Contains(strings.ToLower(inst.Platform), value)
		case "zone":
			return strings.Contains(strings.ToLower(inst.AvailabilityZone), value)
		case "vpc":
			return strings.Contains(strings.ToLower(inst.VPCId), value)
		case "subnet":
			return strings.Contains(strings.ToLower(inst.SubnetId), value)
		}
		// Unrecognized column prefix: fall through to a plain substring match
		// against the whole query so "foo:bar" free text still works.
	}

	q := strings.ToLower(query)
	return strings.Contains(strings.ToLower(inst.Name), q) || strings.Contains(strings.ToLower(inst.ID), q)
}

// newFilterInput builds the "/" filter input field. onApply is called on
// every keystroke change (live filtering) and onClose when the field is
// dismissed via Esc or Enter.
func newFilterInput(onApply func(query string), onClose func()) *tview.InputField {
	field := tview.NewInputField().
		SetLabel("Filter (text or column:value): ").
		SetFieldWidth(0)

	field.SetChangedFunc(onApply)

	field.SetDoneFunc(func(key tcell.Key) {
		onClose()
	})

	return field
}
