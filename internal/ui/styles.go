package ui

import "github.com/gdamore/tcell/v2"

var colorHeader = tcell.ColorAqua

// stateColor returns the display color for an EC2 instance lifecycle state,
// mirroring e1s/k9s's convention of color-coding resource state.
func stateColor(state string) tcell.Color {
	switch state {
	case "running":
		return tcell.ColorGreen
	case "stopped", "terminated":
		return tcell.ColorRed
	case "pending", "stopping", "shutting-down":
		return tcell.ColorYellow
	default:
		return tcell.ColorWhite
	}
}
