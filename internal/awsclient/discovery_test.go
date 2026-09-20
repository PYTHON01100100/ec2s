package awsclient

import (
	"context"
	"errors"
	"sort"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/PYTHON01100100/ec2s/internal/config"
)

// fakeEC2Client is a minimal stand-in for ec2.DescribeInstancesAPIClient so
// tests never call real AWS.
type fakeEC2Client struct {
	output *ec2.DescribeInstancesOutput
	err    error
}

func (f *fakeEC2Client) DescribeInstances(_ context.Context, _ *ec2.DescribeInstancesInput, _ ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.output, nil
}

func instanceOutput(id, name string) *ec2.DescribeInstancesOutput {
	return &ec2.DescribeInstancesOutput{
		Reservations: []ec2types.Reservation{
			{
				Instances: []ec2types.Instance{
					{
						InstanceId:   aws.String(id),
						InstanceType: ec2types.InstanceTypeT3Micro,
						State:        &ec2types.InstanceState{Name: ec2types.InstanceStateNameRunning},
						Tags: []ec2types.Tag{
							{Key: aws.String("Name"), Value: aws.String(name)},
						},
					},
				},
			},
		},
	}
}

func TestDiscover_AggregatesAcrossEnvironments(t *testing.T) {
	envs := []config.Environment{
		{AccountName: "prod", Profile: "prod-profile", Region: "us-east-1"},
		{AccountName: "staging", Profile: "staging-profile", Region: "eu-west-1"},
	}

	factory := func(_ context.Context, env config.Environment) (ec2.DescribeInstancesAPIClient, error) {
		return &fakeEC2Client{output: instanceOutput("i-"+env.AccountName, env.AccountName+"-box")}, nil
	}

	results := discover(context.Background(), envs, factory)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	sort.Slice(results, func(i, j int) bool { return results[i].Env.AccountName < results[j].Env.AccountName })

	for i, r := range results {
		if r.Err != nil {
			t.Errorf("result %d: unexpected error: %v", i, r.Err)
		}
		if len(r.Instances) != 1 {
			t.Errorf("result %d: expected 1 instance, got %d", i, len(r.Instances))
			continue
		}
		if r.Instances[0].AccountName != r.Env.AccountName {
			t.Errorf("result %d: instance not tagged with its environment account", i)
		}
	}
}

func TestDiscover_PartialFailureIsolated(t *testing.T) {
	envs := []config.Environment{
		{AccountName: "good", Profile: "good-profile", Region: "us-east-1"},
		{AccountName: "bad", Profile: "bad-profile", Region: "us-east-1"},
	}

	factory := func(_ context.Context, env config.Environment) (ec2.DescribeInstancesAPIClient, error) {
		if env.AccountName == "bad" {
			return nil, errors.New("boom: invalid profile")
		}
		return &fakeEC2Client{output: instanceOutput("i-good", "good-box")}, nil
	}

	results := discover(context.Background(), envs, factory)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	var goodResult, badResult Result
	for _, r := range results {
		switch r.Env.AccountName {
		case "good":
			goodResult = r
		case "bad":
			badResult = r
		}
	}

	if goodResult.Err != nil {
		t.Errorf("expected good environment to succeed, got error: %v", goodResult.Err)
	}
	if len(goodResult.Instances) != 1 {
		t.Errorf("expected good environment to have 1 instance, got %d", len(goodResult.Instances))
	}
	if badResult.Err == nil {
		t.Error("expected bad environment to have an error, got nil")
	}
}

func TestDiscover_DescribeInstancesErrorIsolated(t *testing.T) {
	envs := []config.Environment{
		{AccountName: "expired-sso", Profile: "p", Region: "us-east-1"},
	}

	factory := func(_ context.Context, _ config.Environment) (ec2.DescribeInstancesAPIClient, error) {
		return &fakeEC2Client{err: errors.New("sso token expired")}, nil
	}

	results := discover(context.Background(), envs, factory)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Err == nil {
		t.Fatal("expected an error from DescribeInstances failure, got nil")
	}
}
