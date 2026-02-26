package config

import (
	"os"
	"testing"
)

func TestValidate_MissingURL(t *testing.T) {
	c := DefaultConfig()
	c.Token = "tok"
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for missing URL")
	}
}

func TestValidate_MissingAuth(t *testing.T) {
	c := DefaultConfig()
	c.URL = "https://mm.example.com"
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for missing auth")
	}
}

func TestValidate_MutuallyExclusiveAuth(t *testing.T) {
	c := DefaultConfig()
	c.URL = "https://mm.example.com"
	c.Token = "tok"
	c.Username = "user"
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for mutually exclusive auth")
	}
}

func TestValidate_InvalidDays(t *testing.T) {
	c := DefaultConfig()
	c.URL = "https://mm.example.com"
	c.Token = "tok"
	c.Days = 0
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for days < 1")
	}
}

func TestValidate_InvalidWorkers(t *testing.T) {
	c := DefaultConfig()
	c.URL = "https://mm.example.com"
	c.Token = "tok"
	c.Workers = 0
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for workers < 1")
	}
}

func TestValidate_InvalidFormat(t *testing.T) {
	c := DefaultConfig()
	c.URL = "https://mm.example.com"
	c.Token = "tok"
	c.Format = "xml"
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for invalid format")
	}
}

func TestValidate_InvalidSortBy(t *testing.T) {
	c := DefaultConfig()
	c.URL = "https://mm.example.com"
	c.Token = "tok"
	c.SortBy = "date"
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for invalid sort-by")
	}
}

func TestValidate_ThresholdOrdering(t *testing.T) {
	c := DefaultConfig()
	c.URL = "https://mm.example.com"
	c.Token = "tok"
	c.ThresholdVeryLow = 5.0
	c.ThresholdLow = 2.0
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for out-of-order thresholds")
	}
}

func TestValidate_OK(t *testing.T) {
	c := DefaultConfig()
	c.URL = "https://mm.example.com"
	c.Token = "tok"
	if err := c.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidate_TrailingSlashTrimmed(t *testing.T) {
	c := DefaultConfig()
	c.URL = "https://mm.example.com/"
	c.Token = "tok"
	if err := c.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.URL != "https://mm.example.com" {
		t.Fatalf("expected trailing slash to be trimmed, got %q", c.URL)
	}
}

func TestResolveEnv(t *testing.T) {
	os.Setenv("MM_URL", "https://env.example.com")
	os.Setenv("MM_TOKEN", "envtoken")
	os.Setenv("MM_USERNAME", "envuser")
	defer os.Unsetenv("MM_URL")
	defer os.Unsetenv("MM_TOKEN")
	defer os.Unsetenv("MM_USERNAME")

	c := DefaultConfig()
	c.ResolveEnv()

	if c.URL != "https://env.example.com" {
		t.Fatalf("expected URL from env, got %q", c.URL)
	}
	if c.Token != "envtoken" {
		t.Fatalf("expected Token from env, got %q", c.Token)
	}
	if c.Username != "envuser" {
		t.Fatalf("expected Username from env, got %q", c.Username)
	}
}

func TestResolveEnv_FlagsTakePrecedence(t *testing.T) {
	os.Setenv("MM_URL", "https://env.example.com")
	os.Setenv("MM_TOKEN", "envtoken")
	defer os.Unsetenv("MM_URL")
	defer os.Unsetenv("MM_TOKEN")

	c := DefaultConfig()
	c.URL = "https://flag.example.com"
	c.Token = "flagtoken"
	c.ResolveEnv()

	if c.URL != "https://flag.example.com" {
		t.Fatalf("expected URL from flag, got %q", c.URL)
	}
	if c.Token != "flagtoken" {
		t.Fatalf("expected Token from flag, got %q", c.Token)
	}
}
