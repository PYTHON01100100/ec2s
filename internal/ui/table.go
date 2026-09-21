package ui

import (
	"fmt"
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/PYTHON01100100/ec2s/internal/awsclient"
)

var tableColumns = []string{"NAME", "INSTANCE ID", "STATE", "TYPE", "OS", "ACCOUNT", "REGION", "ZONE", "VPC ID", "SUBNET ID", "PUBLIC IP", "PRIVATE IP", "LAUNCH TIME"}

const stateColumn = 2

// Table renders the aggregated instances table.
type Table struct {
	view      *tview.Table
	instances []awsclient.Instance // currently displayed rows, in row order

	// onSelect is invoked whenever the selected row changes (interactively
	// or via SetInstances), with nil when there is no selectable row.
	onSelect func(inst *awsclient.Instance)
}

func newTable() *Table {
	view := tview.NewTable().
		SetSelectable(true, false).
		SetFixed(1, 0).
		SetBorders(false)
	view.SetBorder(true).
		SetBorderPadding(0, 0, 1, 1).
		SetTitle(fmt.Sprintf(tableTitleFmt, "Instances", "all accounts", 0))
	view.SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorTeal).Foreground(tcell.ColorBlack))

	t := &Table{view: view}
	t.drawHeader()

	view.SetSelectionChangedFunc(func(row, column int) {
		if t.onSelect == nil {
			return
		}
		inst, ok := t.SelectedInstance()
		if !ok {
			t.onSelect(nil)
			return
		}
		t.onSelect(&inst)
	})

	return t
}

// SetOnSelect registers fn to be called whenever the selected instance
// changes. It is not called for the table built by newTable until the first
// SetInstances call.
func (t *Table) SetOnSelect(fn func(inst *awsclient.Instance)) {
	t.onSelect = fn
}

func (t *Table) drawHeader() {
	for col, name := range tableColumns {
		cell := tview.NewTableCell(name).
			SetSelectable(false).
			SetTextColor(colorHeader).
			SetAttributes(tcell.AttrBold)
		t.view.SetCell(0, col, cell)
	}
}

// SetInstances replaces the displayed rows, sorted by name, and updates the
// table title to reflect scope and count. It preserves the current row
// selection when possible, and always notifies onSelect so dependent views
// (e.g. the info panel) stay in sync.
func (t *Table) SetInstances(instances []awsclient.Instance, scope string) {
	sorted := make([]awsclient.Instance, len(instances))
	copy(sorted, instances)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })

	prevRow, _ := t.view.GetSelection()

	t.instances = sorted
	t.view.Clear()
	t.drawHeader()
	t.view.SetTitle(fmt.Sprintf(tableTitleFmt, "Instances", scope, len(sorted)))

	if len(sorted) == 0 {
		cell := tview.NewTableCell("No instances found").
			SetSelectable(false).
			SetAlign(tview.AlignCenter)
		t.view.SetCell(1, 0, cell)
		if t.onSelect != nil {
			t.onSelect(nil)
		}
		return
	}

	for i, inst := range sorted {
		t.setRow(i+1, inst)
	}

	row := prevRow
	if row < 1 || row > len(sorted) {
		row = 1
	}
	t.view.Select(row, 0)
	if t.onSelect != nil {
		selected := sorted[row-1]
		t.onSelect(&selected)
	}
}

func (t *Table) setRow(row int, inst awsclient.Instance) {
	launch := ""
	if !inst.LaunchTime.IsZero() {
		launch = inst.LaunchTime.Local().Format("2006-01-02 15:04:05")
	}

	values := []string{
		inst.Name,
		inst.ID,
		inst.State,
		inst.Type,
		orDash(inst.Platform),
		inst.AccountName,
		inst.Region,
		orDash(inst.AvailabilityZone),
		orDash(inst.VPCId),
		orDash(inst.SubnetId),
		orDash(inst.PublicIP),
		orDash(inst.PrivateIP),
		launch,
	}

	for col, v := range values {
		cell := tview.NewTableCell(v)
		if col == stateColumn {
			cell.SetTextColor(stateColor(inst.State))
		}
		t.view.SetCell(row, col, cell)
	}
}

// SelectedInstance returns the instance under the current selection, if any.
func (t *Table) SelectedInstance() (awsclient.Instance, bool) {
	row, _ := t.view.GetSelection()
	idx := row - 1
	if idx < 0 || idx >= len(t.instances) {
		return awsclient.Instance{}, false
	}
	return t.instances[idx], true
}

// SelectTop moves the selection to the first data row.
func (t *Table) SelectTop() {
	if len(t.instances) > 0 {
		t.view.Select(1, 0)
	}
}

// SelectBottom moves the selection to the last data row.
func (t *Table) SelectBottom() {
	if len(t.instances) > 0 {
		t.view.Select(len(t.instances), 0)
	}
}
