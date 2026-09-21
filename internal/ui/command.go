package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// newCommandInput builds the "E" run-command input field. onSubmit is
// called with the entered command on Enter (and only then — unlike the
// filter box, a command shouldn't fire on every keystroke); onCancel on Esc.
func newCommandInput(instName string, onSubmit func(command string), onCancel func()) *tview.InputField {
	field := tview.NewInputField().
		SetLabel(fmt.Sprintf("Run command on %s (no SSH, via SSM): ", instName)).
		SetFieldWidth(0)

	field.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			text := field.GetText()
			if text != "" {
				onSubmit(text)
				return
			}
		}
		onCancel()
	})

	return field
}

// commandResultText formats a command's status and output for the result
// view. Truncated stdout/stderr are the AWS SDK's own limits (8000 chars for
// stderr, 24000 for stdout captured inline), not something ec2s imposes.
// commandID is shown so a failure can be looked up directly in the AWS
// Console (Systems Manager > Run Command > Command history) or via
// `aws ssm list-command-invocations --details` for more detail than
// GetCommandInvocation returns when stdout/stderr come back empty.
func commandResultText(instName, command, commandID, status, statusDetails, stdout, stderr string) string {
	text := fmt.Sprintf("[teal::b]%s[-:-:-] on [fuchsia::b]%s[-:-:-]\n[::b]Status:[-:-:-] %s",
		tview.Escape(command), tview.Escape(instName), status)
	if statusDetails != "" && statusDetails != status {
		text += fmt.Sprintf(" (%s)", statusDetails)
	}
	if commandID != "" {
		text += fmt.Sprintf("\n[::b]Command ID:[-:-:-] %s", commandID)
	}
	text += "\n"

	if stdout != "" {
		text += "\n[green::b]stdout[-:-:-]\n" + tview.Escape(stdout)
	}
	if stderr != "" {
		text += "\n[red::b]stderr[-:-:-]\n" + tview.Escape(stderr)
	}
	if stdout == "" && stderr == "" {
		text += "\n(no output)"
	}

	return text
}

// newCommandResultView builds the scrollable output pane shown after a
// command finishes (or while it's still running).
func newCommandResultView() *tview.TextView {
	view := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	view.SetBorder(true).SetTitle(" Command output (Esc to close) ")
	return view
}
