package ui

import (
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/PYTHON01100100/ec2s/internal/awsclient"
)

var tableColumns = []string{"NAME", "INSTANCE ID", "ACCOUNT", "REGION", "TYPE", "STATE", "PUBLIC IP", "PRIVATE IP", "LAUNCH TIME"}

const stateColumn = 5

// Table renders the aggregated instances table.
type Table struct {
	view      *tview.Table
	instances []awsclient.Instance // currently displayed rows, in row order
}

func newTable() *Table {
	view := tview.NewTable().
		SetSelectable(true, false).
		SetFixed(1, 0).
		SetBorders(false)
	t := &Table{view: view}
	t.drawHeader()
	return t
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

// SetInstances replaces the displayed rows, sorted by name.
func (t *Table) SetInstances(instances []awsclient.Instance) {
	sorted := make([]awsclient.Instance, len(instances))
	copy(sorted, instances)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })

	t.instances = sorted
	t.view.Clear()
	t.drawHeader()

	if len(sorted) == 0 {
		cell := tview.NewTableCell("No instances found").
			SetSelectable(false).
			SetAlign(tview.AlignCenter)
		t.view.SetCell(1, 0, cell)
		return
	}

	for i, inst := range sorted {
		t.setRow(i+1, inst)
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
		inst.AccountName,
		inst.Region,
		inst.Type,
		inst.State,
		inst.PublicIP,
		inst.PrivateIP,
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
