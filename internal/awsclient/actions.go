package awsclient

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"

	"github.com/PYTHON01100100/ec2s/internal/config"
)

// StartInstance requests a start of instanceID in env. It builds a client
// for that specific account/region using the same credential chain as
// discovery.
func StartInstance(ctx context.Context, env config.Environment, instanceID string) error {
	client, err := NewClient(ctx, env)
	if err != nil {
		return err
	}
	if _, err := client.StartInstances(ctx, &ec2.StartInstancesInput{
		InstanceIds: []string{instanceID},
	}); err != nil {
		return fmt.Errorf("start %s: %w", instanceID, err)
	}
	return nil
}

// StopInstance requests a graceful stop of instanceID in env. It builds a
// client for that specific account/region using the same credential chain
// as discovery.
func StopInstance(ctx context.Context, env config.Environment, instanceID string) error {
	client, err := NewClient(ctx, env)
	if err != nil {
		return err
	}
	if _, err := client.StopInstances(ctx, &ec2.StopInstancesInput{
		InstanceIds: []string{instanceID},
	}); err != nil {
		return fmt.Errorf("stop %s: %w", instanceID, err)
	}
	return nil
}

// TerminateInstance requests permanent termination of instanceID in env.
// This is irreversible; callers are responsible for confirming with the
// user before calling it.
func TerminateInstance(ctx context.Context, env config.Environment, instanceID string) error {
	client, err := NewClient(ctx, env)
	if err != nil {
		return err
	}
	if _, err := client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		InstanceIds: []string{instanceID},
	}); err != nil {
		return fmt.Errorf("terminate %s: %w", instanceID, err)
	}
	return nil
}
