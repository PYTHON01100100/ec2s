package config

// Environment is one (account, region) pair to discover EC2 instances in,
// named after the CDK Environment{Account, Region} concept.
type Environment struct {
	AccountName string
	Profile     string
	Region      string
}

// Expand flattens each account's region list into one Environment per
// (account, region) pair.
func Expand(accounts []Account) []Environment {
	var envs []Environment
	for _, a := range accounts {
		for _, region := range a.Regions {
			envs = append(envs, Environment{
				AccountName: a.Name,
				Profile:     a.Profile,
				Region:      region,
			})
		}
	}
	return envs
}
