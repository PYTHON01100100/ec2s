package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "ec2s.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writing temp config: %v", err)
	}
	return path
}

func TestLoad_Valid(t *testing.T) {
	path := writeTempConfig(t, `
accounts:
  - name: prod
    profile: prod-profile
    regions: [us-east-1, eu-west-1]
  - name: staging
    profile: staging-profile
    region: us-east-1
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cfg.Accounts) != 2 {
		t.Fatalf("expected 2 accounts, got %d", len(cfg.Accounts))
	}

	staging := cfg.Accounts[1]
	if staging.Region != "" {
		t.Errorf("expected region shorthand normalized to empty, got %q", staging.Region)
	}
	if len(staging.Regions) != 1 || staging.Regions[0] != "us-east-1" {
		t.Errorf("expected regions=[us-east-1], got %v", staging.Regions)
	}
}

func TestLoad_DuplicateName(t *testing.T) {
	path := writeTempConfig(t, `
accounts:
  - name: prod
    profile: a
    region: us-east-1
  - name: prod
    profile: b
    region: us-east-1
`)

	if _, err := Load(path); err == nil {
		t.Fatal("expected error for duplicate account name, got nil")
	}
}

func TestLoad_MissingProfile(t *testing.T) {
	path := writeTempConfig(t, `
accounts:
  - name: prod
    region: us-east-1
`)

	if _, err := Load(path); err == nil {
		t.Fatal("expected error for missing profile, got nil")
	}
}

func TestLoad_BothRegionAndRegions(t *testing.T) {
	path := writeTempConfig(t, `
accounts:
  - name: prod
    profile: p
    region: us-east-1
    regions: [eu-west-1]
`)

	if _, err := Load(path); err == nil {
		t.Fatal("expected error for both region and regions set, got nil")
	}
}

func TestLoad_NoRegion(t *testing.T) {
	path := writeTempConfig(t, `
accounts:
  - name: prod
    profile: p
`)

	if _, err := Load(path); err == nil {
		t.Fatal("expected error for missing region, got nil")
	}
}

func TestLoad_NoAccounts(t *testing.T) {
	path := writeTempConfig(t, `accounts: []`)

	if _, err := Load(path); err == nil {
		t.Fatal("expected error for empty accounts list, got nil")
	}
}

func TestLoad_ExplicitPathNotFound(t *testing.T) {
	if _, err := Load(filepath.Join(t.TempDir(), "missing.yaml")); err == nil {
		t.Fatal("expected error for missing explicit config path, got nil")
	}
}

func TestFallback_UsesAWSProfileEnv(t *testing.T) {
	t.Setenv("AWS_PROFILE", "my-profile")
	t.Setenv("AWS_REGION", "ap-south-1")

	cfg, err := Fallback()
	if err != nil {
		t.Fatalf("Fallback: %v", err)
	}
	if len(cfg.Accounts) != 1 {
		t.Fatalf("expected 1 account, got %d", len(cfg.Accounts))
	}
	got := cfg.Accounts[0]
	if got.Name != "my-profile" || got.Profile != "my-profile" {
		t.Errorf("expected account named/profiled my-profile, got %+v", got)
	}
	if len(got.Regions) != 1 || got.Regions[0] != "ap-south-1" {
		t.Errorf("expected region ap-south-1 from env, got %v", got.Regions)
	}
}

func TestFallback_NoProfileNoLocalFiles(t *testing.T) {
	t.Setenv("AWS_PROFILE", "")
	t.Setenv("AWS_REGION", "")
	t.Setenv("AWS_DEFAULT_REGION", "")
	t.Setenv("AWS_CONFIG_FILE", filepath.Join(t.TempDir(), "nonexistent-config"))
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", filepath.Join(t.TempDir(), "nonexistent-creds"))

	cfg, err := Fallback()
	if err != nil {
		t.Fatalf("Fallback: %v", err)
	}
	if len(cfg.Accounts) != 1 || cfg.Accounts[0].Name != "default" {
		t.Fatalf("expected single default account, got %+v", cfg.Accounts)
	}
	if cfg.Accounts[0].Regions[0] != defaultFallbackRegion {
		t.Errorf("expected fallback region %q, got %q", defaultFallbackRegion, cfg.Accounts[0].Regions[0])
	}
}

func TestFallback_DiscoversLocalProfiles(t *testing.T) {
	t.Setenv("AWS_PROFILE", "")
	t.Setenv("AWS_REGION", "")
	t.Setenv("AWS_DEFAULT_REGION", "")

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config")
	if err := os.WriteFile(configPath, []byte("[default]\nregion = us-west-2\n\n[profile work]\nregion = eu-central-1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	credsPath := filepath.Join(dir, "credentials")
	if err := os.WriteFile(credsPath, []byte("[personal]\naws_access_key_id = x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AWS_CONFIG_FILE", configPath)
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", credsPath)

	cfg, err := Fallback()
	if err != nil {
		t.Fatalf("Fallback: %v", err)
	}

	names := map[string]Account{}
	for _, a := range cfg.Accounts {
		names[a.Name] = a
	}
	if _, ok := names["work"]; !ok {
		t.Errorf("expected discovered profile 'work', got accounts %+v", cfg.Accounts)
	}
	if a, ok := names["work"]; ok && (len(a.Regions) != 1 || a.Regions[0] != "eu-central-1") {
		t.Errorf("expected work profile region eu-central-1 from config file, got %v", a.Regions)
	}
	if _, ok := names["personal"]; !ok {
		t.Errorf("expected discovered profile 'personal' from credentials file, got accounts %+v", cfg.Accounts)
	}
}

func TestExpand(t *testing.T) {
	accounts := []Account{
		{Name: "prod", Profile: "prod-profile", Regions: []string{"us-east-1", "eu-west-1"}},
		{Name: "staging", Profile: "staging-profile", Regions: []string{"us-east-1"}},
	}

	envs := Expand(accounts)
	if len(envs) != 3 {
		t.Fatalf("expected 3 environments, got %d: %+v", len(envs), envs)
	}
	if envs[0].AccountName != "prod" || envs[0].Region != "us-east-1" {
		t.Errorf("unexpected first environment: %+v", envs[0])
	}
}
