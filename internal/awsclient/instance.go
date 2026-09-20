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
	ID          string
	Name        string
	AccountName string
	Profile     string
	Region      string
	Type        string
	State       string
	PublicIP    string
	PrivateIP   string
	LaunchTime  time.Time
	VPCId       string
	Tags        map[string]string
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
		PublicIP:    strOrEmpty(raw.PublicIpAddress),
		PrivateIP:   strOrEmpty(raw.PrivateIpAddress),
		VPCId:       strOrEmpty(raw.VpcId),
		Tags:        map[string]string{},
	}

	if raw.State != nil {
		inst.State = string(raw.State.Name)
	}
	if raw.LaunchTime != nil {
		inst.LaunchTime = *raw.LaunchTime
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
