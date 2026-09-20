package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

const defaultFallbackRegion = "us-east-1"

// Fallback builds a Config when no ec2s config file was found. It tries, in
// order: $AWS_PROFILE as a single account, then every profile discovered in
// the local AWS config/credentials files, then finally a single "default"
// account relying on the SDK's own default credential chain.
func Fallback() (*Config, error) {
	if profile := os.Getenv("AWS_PROFILE"); profile != "" {
		return &Config{Accounts: []Account{
			{Name: profile, Profile: profile, Regions: []string{regionForProfile(profile)}},
		}}, nil
	}

	if profiles := discoverLocalProfiles(); len(profiles) > 0 {
		accounts := make([]Account, 0, len(profiles))
		for _, p := range profiles {
			accounts = append(accounts, Account{Name: p, Profile: p, Regions: []string{regionForProfile(p)}})
		}
		return &Config{Accounts: accounts}, nil
	}

	return &Config{Accounts: []Account{
		{Name: "default", Profile: "", Regions: []string{regionForProfile("")}},
	}}, nil
}

// regionForProfile resolves a region for a profile using $AWS_REGION,
// $AWS_DEFAULT_REGION, that profile's own "region" setting in
// ~/.aws/config, or finally a hardcoded fallback.
func regionForProfile(profile string) string {
	if r := os.Getenv("AWS_REGION"); r != "" {
		return r
	}
	if r := os.Getenv("AWS_DEFAULT_REGION"); r != "" {
		return r
	}
	if r := regionFromAWSConfig(profile); r != "" {
		return r
	}
	return defaultFallbackRegion
}

// discoverLocalProfiles enumerates profile names from ~/.aws/config
// ([profile name] sections, "default" section) and ~/.aws/credentials
// ([name] sections), deduplicated.
func discoverLocalProfiles() []string {
	seen := map[string]bool{}
	var profiles []string

	add := func(name string) {
		if name == "" || seen[name] {
			return
		}
		seen[name] = true
		profiles = append(profiles, name)
	}

	for _, section := range sectionHeaders(awsConfigPath()) {
		add(strings.TrimSpace(strings.TrimPrefix(section, "profile ")))
	}
	for _, section := range sectionHeaders(awsCredentialsPath()) {
		add(section)
	}

	return profiles
}

// regionFromAWSConfig reads the "region" key under the matching section of
// ~/.aws/config for the given profile ("" means the [default] section).
func regionFromAWSConfig(profile string) string {
	wantSection := "default"
	if profile != "" {
		wantSection = "profile " + profile
	}

	f, err := os.Open(awsConfigPath())
	if err != nil {
		return ""
	}
	defer f.Close()

	inSection := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if header, ok := parseSectionHeader(line); ok {
			inSection = header == wantSection
			continue
		}
		if !inSection {
			continue
		}
		key, value, ok := parseKeyValue(line)
		if ok && key == "region" {
			return value
		}
	}
	return ""
}

// sectionHeaders returns every "[...]" section header found in path.
func sectionHeaders(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var headers []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if header, ok := parseSectionHeader(strings.TrimSpace(scanner.Text())); ok {
			headers = append(headers, header)
		}
	}
	return headers
}

func parseSectionHeader(line string) (string, bool) {
	if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
		return strings.TrimSpace(line[1 : len(line)-1]), true
	}
	return "", false
}

func parseKeyValue(line string) (key, value string, ok bool) {
	if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
		return "", "", false
	}
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), true
}

func awsConfigPath() string {
	if p := os.Getenv("AWS_CONFIG_FILE"); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".aws", "config")
}

func awsCredentialsPath() string {
	if p := os.Getenv("AWS_SHARED_CREDENTIALS_FILE"); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".aws", "credentials")
}
