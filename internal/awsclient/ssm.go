package awsclient

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"

	"github.com/PYTHON01100100/ec2s/internal/config"
)

const (
	commandPollInterval = time.Second
	commandTimeout      = 45 * time.Second
)

// CommandResult is the outcome of running a shell command on an instance via
// SSM Run Command.
type CommandResult struct {
	Status string
	Stdout string
	Stderr string
}

// RunCommand runs command on instanceID via SSM Run Command and waits for it
// to finish or commandTimeout to elapse. This is how ec2s gets a "type a
// command, see the output" experience with zero SSH: no open port 22, no
// SSH keys — the SSM agent already required for the AWS Console's own
// "Connect" button executes the command and returns its output over the
// same IAM-authorized API calls ec2s already uses for everything else. The
// instance needs the SSM agent running and an instance profile with SSM
// permissions; without that, SendCommand itself fails with a clear error
// (surfaced as-is to the caller).
func RunCommand(ctx context.Context, env config.Environment, instanceID, platform, command string) (CommandResult, error) {
	client, err := NewSSMClient(ctx, env)
	if err != nil {
		return CommandResult{}, err
	}

	ctx, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()

	document := "AWS-RunShellScript"
	if strings.Contains(strings.ToLower(platform), "windows") {
		document = "AWS-RunPowerShellScript"
	}

	sendOut, err := client.SendCommand(ctx, &ssm.SendCommandInput{
		DocumentName: &document,
		InstanceIds:  []string{instanceID},
		Parameters:   map[string][]string{"commands": {command}},
	})
	if err != nil {
		return CommandResult{}, fmt.Errorf("send command: %w", err)
	}
	commandID := sendOut.Command.CommandId

	// GetCommandInvocation can 404 (InvocationDoesNotExist) for a brief
	// moment right after SendCommand, before SSM has created the
	// invocation record — that specific error means "still pending", so
	// keep polling. Any OTHER error (AccessDenied, ValidationException,
	// ...) is permanent and must be returned immediately: silently
	// retrying it until the timeout would just mask the real cause behind
	// a useless "timed out" message.
	var notExist *ssmtypes.InvocationDoesNotExist
	for {
		invocation, err := client.GetCommandInvocation(ctx, &ssm.GetCommandInvocationInput{
			CommandId:  commandID,
			InstanceId: &instanceID,
		})
		if err != nil {
			if !errors.As(err, &notExist) {
				return CommandResult{}, fmt.Errorf("get command output: %w", err)
			}
			select {
			case <-ctx.Done():
				return CommandResult{}, fmt.Errorf("timed out waiting for command output: %w", ctx.Err())
			case <-time.After(commandPollInterval):
				continue
			}
		}

		switch invocation.Status {
		case ssmtypes.CommandInvocationStatusPending, ssmtypes.CommandInvocationStatusInProgress, ssmtypes.CommandInvocationStatusDelayed:
			select {
			case <-ctx.Done():
				return CommandResult{}, fmt.Errorf("timed out waiting for command output: %w", ctx.Err())
			case <-time.After(commandPollInterval):
				continue
			}
		default:
			return CommandResult{
				Status: string(invocation.Status),
				Stdout: strOrEmpty(invocation.StandardOutputContent),
				Stderr: strOrEmpty(invocation.StandardErrorContent),
			}, nil
		}
	}
}
