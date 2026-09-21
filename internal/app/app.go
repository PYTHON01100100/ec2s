// Package app wires together config loading, AWS discovery, and the tview
// UI into the running ec2s program.
package app

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/PYTHON01100100/ec2s/internal/awsclient"
	"github.com/PYTHON01100100/ec2s/internal/config"
	"github.com/PYTHON01100100/ec2s/internal/ui"
)

// autoRefreshInterval is how often ec2s re-fetches all accounts/regions on
// its own, on top of the manual Ctrl-R refresh.
const autoRefreshInterval = 60 * time.Second

// Run loads the ec2s configuration (or falls back to auto-discovery),
// performs an initial concurrent discovery across every configured
// account/region, and launches the terminal UI. It blocks until the user
// quits.
func Run(ctx context.Context, configPath, version string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	envs := config.Expand(cfg.Accounts)
	accountNames := accountNames(cfg.Accounts)
	regionCount := len(uniqueRegions(envs))

	var uiApp *ui.App
	refresh := func() {
		go func() {
			results := awsclient.Discover(ctx, envs)
			instances, warnings := flatten(results)
			uiApp.SetInstancesAsync(instances, warnings, len(cfg.Accounts), regionCount)
		}()
	}

	uiApp = ui.New(accountNames, ui.Actions{
		Refresh: refresh,
		Start: func(inst awsclient.Instance) {
			go runAction(ctx, uiApp, inst, "start", awsclient.StartInstance, refresh)
		},
		Stop: func(inst awsclient.Instance) {
			go runAction(ctx, uiApp, inst, "stop", awsclient.StopInstance, refresh)
		},
		Terminate: func(inst awsclient.Instance) {
			go runAction(ctx, uiApp, inst, "terminate", awsclient.TerminateInstance, refresh)
		},
		RunCommand: func(inst awsclient.Instance, command string) {
			go func() {
				env := config.Environment{AccountName: inst.AccountName, Profile: inst.Profile, Region: inst.Region}
				result, err := awsclient.RunCommand(ctx, env, inst.ID, inst.Platform, command)
				if err != nil {
					uiApp.SetCommandErrorAsync(inst, command, err)
					return
				}
				uiApp.SetCommandResultAsync(inst, command, result)
			}()
		},
	}, version)
	uiApp.SetTotals(len(cfg.Accounts), regionCount)
	refresh()
	go autoRefresh(ctx, refresh)

	return uiApp.Run()
}

// autoRefresh re-fetches all accounts/regions every autoRefreshInterval,
// until ctx is cancelled (on quit or a termination signal).
func autoRefresh(ctx context.Context, refresh func()) {
	ticker := time.NewTicker(autoRefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			refresh()
		}
	}
}

// actionPastTense maps the imperative verb used in runAction's log/notify
// messages to its past tense, since "stop"/"terminate" don't share a
// suffix rule ("stopped" vs "terminated").
var actionPastTense = map[string]string{
	"start":     "started",
	"stop":      "stopped",
	"terminate": "terminated",
}

// runAction performs a single-instance action (stop/terminate) against its
// own account/region, reports the result to the UI, and triggers a refresh
// on success so the instance's updated state shows up promptly.
func runAction(ctx context.Context, uiApp *ui.App, inst awsclient.Instance, verb string, do func(context.Context, config.Environment, string) error, refresh func()) {
	env := config.Environment{AccountName: inst.AccountName, Profile: inst.Profile, Region: inst.Region}
	if err := do(ctx, env, inst.ID); err != nil {
		uiApp.Notify(fmt.Sprintf("failed to %s %s: %v", verb, inst.Name, err), true)
		return
	}
	uiApp.Notify(fmt.Sprintf("%s %s", actionPastTense[verb], inst.Name), false)
	refresh()
}

func accountNames(accounts []config.Account) []string {
	names := make([]string, len(accounts))
	for i, a := range accounts {
		names[i] = a.Name
	}
	return names
}

func uniqueRegions(envs []config.Environment) []string {
	seen := map[string]bool{}
	var regions []string
	for _, e := range envs {
		if !seen[e.Region] {
			seen[e.Region] = true
			regions = append(regions, e.Region)
		}
	}
	return regions
}

// flatten merges per-Environment discovery results into one instance slice
// and a list of human-readable warnings for any environments that failed.
func flatten(results []awsclient.Result) ([]awsclient.Instance, []string) {
	var instances []awsclient.Instance
	var warnings []string

	sorted := make([]awsclient.Result, len(results))
	copy(sorted, results)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Env.AccountName < sorted[j].Env.AccountName
	})

	for _, r := range sorted {
		if r.Err != nil {
			warnings = append(warnings, fmt.Sprintf("%s (%s/%s): %v", r.Env.AccountName, r.Env.Profile, r.Env.Region, r.Err))
			continue
		}
		instances = append(instances, r.Instances...)
	}

	// A failure almost always means the running binary resolved AWS config
	// from somewhere other than where the user actually ran `aws configure`
	// (classic on WSL: a Windows-built binary reads %USERPROFILE%\.aws, a
	// separate filesystem from a WSL shell's own $HOME/.aws). Surface
	// exactly where ec2s looked so this is self-diagnosable from the UI.
	if len(warnings) > 0 {
		warnings = append(warnings, config.DebugContext())
	}

	return instances, warnings
}
