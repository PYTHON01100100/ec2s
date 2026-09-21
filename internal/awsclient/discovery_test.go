package awsclient

import (
	"context"
	"errors"
	"sort"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"

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

// fakeSSMClient is a minimal stand-in for
// ssm.DescribeInstanceInformationAPIClient so tests never call real AWS.
type fakeSSMClient struct {
	statuses map[string]ssmtypes.PingStatus // instance ID -> ping status
	err      error
}

func (f *fakeSSMClient) DescribeInstanceInformation(_ context.Context, _ *ssm.DescribeInstanceInformationInput, _ ...func(*ssm.Options)) (*ssm.DescribeInstanceInformationOutput, error) {
	if f.err != nil {
		return nil, f.err
	}
	list := make([]ssmtypes.InstanceInformation, 0, len(f.statuses))
	for id, status := range f.statuses {
		list = append(list, ssmtypes.InstanceInformation{InstanceId: aws.String(id), PingStatus: status})
	}
	return &ssm.DescribeInstanceInformationOutput{InstanceInformationList: list}, nil
}

// noSSMRecords is the default ssmClientFactory for tests that don't care
// about SSM status: every instance simply has no SSM record.
func noSSMRecords(_ context.Context, _ config.Environment) (ssm.DescribeInstanceInformationAPIClient, error) {
	return &fakeSSMClient{}, nil
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

	results := discover(context.Background(), envs, factory, noSSMRecords)
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

	results := discover(context.Background(), envs, factory, noSSMRecords)
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

	results := discover(context.Background(), envs, factory, noSSMRecords)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Err == nil {
		t.Fatal("expected an error from DescribeInstances failure, got nil")
	}
}

func TestDiscover_MergesSSMStatus(t *testing.T) {
	envs := []config.Environment{
		{AccountName: "prod", Profile: "p", Region: "us-east-1"},
	}

	factory := func(_ context.Context, _ config.Environment) (ec2.DescribeInstancesAPIClient, error) {
		return &fakeEC2Client{output: &ec2.DescribeInstancesOutput{
			Reservations: []ec2types.Reservation{{Instances: []ec2types.Instance{
				{InstanceId: aws.String("i-managed"), InstanceType: ec2types.InstanceTypeT3Micro, State: &ec2types.InstanceState{Name: ec2types.InstanceStateNameRunning}},
				{InstanceId: aws.String("i-unmanaged"), InstanceType: ec2types.InstanceTypeT3Micro, State: &ec2types.InstanceState{Name: ec2types.InstanceStateNameRunning}},
			}}},
		}}, nil
	}
	ssmFactory := func(_ context.Context, _ config.Environment) (ssm.DescribeInstanceInformationAPIClient, error) {
		return &fakeSSMClient{statuses: map[string]ssmtypes.PingStatus{"i-managed": ssmtypes.PingStatusOnline}}, nil
	}

	results := discover(context.Background(), envs, factory, ssmFactory)
	if len(results) != 1 || results[0].Err != nil {
		t.Fatalf("unexpected results: %+v", results)
	}

	byID := map[string]Instance{}
	for _, inst := range results[0].Instances {
		byID[inst.ID] = inst
	}

	if got := byID["i-managed"].SSMStatus; got != "Online" {
		t.Errorf("expected i-managed SSMStatus=Online, got %q", got)
	}
	if got := byID["i-unmanaged"].SSMStatus; got != SSMStatusNotManaged {
		t.Errorf("expected i-unmanaged SSMStatus=%q, got %q", SSMStatusNotManaged, got)
	}
}

func TestDiscover_SSMErrorLeavesStatusUnknown(t *testing.T) {
	envs := []config.Environment{
		{AccountName: "prod", Profile: "p", Region: "us-east-1"},
	}

	factory := func(_ context.Context, _ config.Environment) (ec2.DescribeInstancesAPIClient, error) {
		return &fakeEC2Client{output: instanceOutput("i-1", "box")}, nil
	}
	ssmFactory := func(_ context.Context, _ config.Environment) (ssm.DescribeInstanceInformationAPIClient, error) {
		return &fakeSSMClient{err: errors.New("AccessDenied: missing ssm:DescribeInstanceInformation")}, nil
	}

	results := discover(context.Background(), envs, factory, ssmFactory)
	if len(results) != 1 || results[0].Err != nil {
		t.Fatalf("unexpected results: %+v", results)
	}
	if len(results[0].Instances) != 1 {
		t.Fatalf("expected EC2 listing to still succeed despite SSM error, got %+v", results[0])
	}
	if got := results[0].Instances[0].SSMStatus; got != "" {
		t.Errorf("expected unknown (empty) SSMStatus when SSM call fails, got %q", got)
	}
}
