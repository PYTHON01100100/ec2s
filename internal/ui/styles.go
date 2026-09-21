package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/PYTHON01100100/ec2s/internal/awsclient"
)

// Color palette, modeled on e1s/k9s: a fixed dark background so the UI reads
// consistently regardless of the terminal's own theme, teal borders, fuchsia
// accents, yellow table headers, green key hints, black-on-color status bar
// chips.
var (
	colorBackground = tcell.NewRGBColor(0x28, 0x28, 0x28)
	colorForeground = tcell.NewRGBColor(0xeb, 0xdb, 0xb2)
	colorBorder     = tcell.ColorTeal

	colorHeader = tcell.ColorYellow

	// Info panel: " info(<name>) "
	infoTitleFmt = " [teal]info([fuchsia::b]%s[teal:-:-]) "
	// Info panel field: " Name:[teal::b] Value "
	infoItemFmt = " %s:[teal::b] %s "
	// Info panel keybinding: " <key> description "
	infoKeyFmt = " [fuchsia::b]<%s> [green:-:-]%s "

	// Table title: " <Instances>scope(N) "
	tableTitleFmt = " [teal::-]<[fuchsia::b]%s[teal::-]>[teal::b]%s[teal::-]([fuchsia::b]%d[teal::-]) "

	// Footer chips
	footerChipFmt       = "[black:gray:] %s [-:-:-]"
	footerChipActiveFmt = "[black:teal:b] %s [-:-:-]"
	footerAppFmt        = "[black:fuchsia:bi] %s:%s [-:-:-]"
)

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

// ssmStatusColor returns the display color for an instance's SSM Agent
// connectivity — this is what E (run command) actually depends on, so it
// gets the same color-coding treatment as EC2 lifecycle state.
func ssmStatusColor(status string) tcell.Color {
	switch status {
	case "Online":
		return tcell.ColorGreen
	case "ConnectionLost", "Inactive":
		return tcell.ColorRed
	case awsclient.SSMStatusNotManaged:
		return tcell.ColorGray
	default:
		return tcell.ColorWhite
	}
}

// applyTheme overrides tview's default styles with a fixed dark theme, so
// ec2s looks the same regardless of the user's terminal color scheme. Must
// be called before any tview primitives are constructed.
func applyTheme() {
	tview.Styles.PrimitiveBackgroundColor = colorBackground
	tview.Styles.ContrastBackgroundColor = colorBackground
	tview.Styles.MoreContrastBackgroundColor = colorBackground
	tview.Styles.BorderColor = colorBorder
	tview.Styles.TitleColor = colorForeground
	tview.Styles.PrimaryTextColor = colorForeground
	tview.Styles.SecondaryTextColor = colorForeground
	tview.Styles.TertiaryTextColor = colorForeground
	tview.Styles.InverseTextColor = colorForeground
	tview.Styles.ContrastSecondaryTextColor = colorForeground
	tview.Styles.GraphicsColor = colorForeground
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
