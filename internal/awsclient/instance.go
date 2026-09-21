// Package awsclient discovers EC2 instances across configured AWS accounts
// and regions concurrently, using the AWS SDK for Go v2's default credential
// chain (static creds, assume-role, credential_process, and SSO are all
// handled by the SDK itself).
package awsclient

import (
	"time"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/PYTHON01100100/ec2s/internal/config"
)

// Instance is a flattened, UI-friendly view of an EC2 instance, tagged with
// the ec2s account/environment it was discovered under.
type Instance struct {
	ID               string
	Name             string
	AccountName      string
	Profile          string
	Region           string
	AvailabilityZone string
	Type             string
	Platform         string
	State            string
	PublicIP         string
	PrivateIP        string
	LaunchTime       time.Time
	VPCId            string
	SubnetId         string
	// SSMStatus is the instance's SSM Agent connectivity (e.g. "Online",
	// "ConnectionLost", "not managed" if it has no SSM record at all, or ""
	// if we couldn't check). It's what E (run command) actually depends
	// on, and is fetched separately from — and best-effort relative to —
	// the core EC2 describe, so a missing ssm:DescribeInstanceInformation
	// permission degrades this field, not the whole instance listing.
	SSMStatus string
	Tags      map[string]string
}

// fromSDK maps an SDK EC2 instance into our Instance model, tagging it with
// the Environment it was fetched from.
func fromSDK(raw ec2types.Instance, env config.Environment) Instance {
	inst := Instance{
		ID:          strOrEmpty(raw.InstanceId),
		AccountName: env.AccountName,
		Profile:     env.Profile,
		Region:      env.Region,
		Type:        string(raw.InstanceType),
		Platform:    strOrEmpty(raw.PlatformDetails),
		PublicIP:    strOrEmpty(raw.PublicIpAddress),
		PrivateIP:   strOrEmpty(raw.PrivateIpAddress),
		VPCId:       strOrEmpty(raw.VpcId),
		SubnetId:    strOrEmpty(raw.SubnetId),
		Tags:        map[string]string{},
	}

	if raw.State != nil {
		inst.State = string(raw.State.Name)
	}
	if raw.LaunchTime != nil {
		inst.LaunchTime = *raw.LaunchTime
	}
	if raw.Placement != nil {
		inst.AvailabilityZone = strOrEmpty(raw.Placement.AvailabilityZone)
	}

	for _, tag := range raw.Tags {
		key := strOrEmpty(tag.Key)
		value := strOrEmpty(tag.Value)
		if key == "" {
			continue
		}
		inst.Tags[key] = value
		if key == "Name" {
			inst.Name = value
		}
	}

	if inst.Name == "" {
		inst.Name = inst.ID
	}

	return inst
}

func strOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
