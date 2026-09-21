package awsclient

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ssm"

	"github.com/PYTHON01100100/ec2s/internal/config"
)

// loadConfig resolves credentials/region for env via the AWS SDK's default
// shared-config credential chain, so static credentials, assume-role
// (role_arn/source_profile), credential_process, and SSO profiles all work
// without any custom handling here. Shared by every per-service client
// constructor below.
func loadConfig(ctx context.Context, env config.Environment) (aws.Config, error) {
	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(env.Region),
	}
	if env.Profile != "" {
		opts = append(opts, awsconfig.WithSharedConfigProfile(env.Profile))
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return aws.Config{}, fmt.Errorf("loading AWS config for profile %q: %w", env.Profile, err)
	}
	return cfg, nil
}

// NewClient builds an EC2 client for the given Environment.
func NewClient(ctx context.Context, env config.Environment) (*ec2.Client, error) {
	cfg, err := loadConfig(ctx, env)
	if err != nil {
		return nil, err
	}
	return ec2.NewFromConfig(cfg), nil
}

// NewSSMClient builds a Systems Manager client for the given Environment,
// used to run commands on instances without SSH (see RunCommand).
func NewSSMClient(ctx context.Context, env config.Environment) (*ssm.Client, error) {
	cfg, err := loadConfig(ctx, env)
	if err != nil {
		return nil, err
	}
	return ssm.NewFromConfig(cfg), nil
}
