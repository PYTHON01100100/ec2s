package awsclient

import (
	"context"
	"fmt"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"

	"github.com/PYTHON01100100/ec2s/internal/config"
)

// NewClient builds an EC2 client for the given Environment. It relies
// entirely on the AWS SDK's default shared-config credential chain, so
// static credentials, assume-role (role_arn/source_profile),
// credential_process, and SSO profiles all work without any custom handling
// here.
func NewClient(ctx context.Context, env config.Environment) (*ec2.Client, error) {
	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(env.Region),
	}
	if env.Profile != "" {
		opts = append(opts, awsconfig.WithSharedConfigProfile(env.Profile))
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("loading AWS config for profile %q: %w", env.Profile, err)
	}

	return ec2.NewFromConfig(cfg), nil
}
