package main

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	awstoolsProvider "github.com/chenrui333/terradozer/internal/awstools/terraform/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
)

func TestSetAWSRegionFromDefault(t *testing.T) {
	testCases := []struct {
		name              string
		awsRegion         string
		awsDefaultRegion  string
		expectedAWSRegion string
	}{
		{
			name:              "uses AWS_DEFAULT_REGION when AWS_REGION is unset",
			awsRegion:         "",
			awsDefaultRegion:  "us-east-1",
			expectedAWSRegion: "us-east-1",
		},
		{
			name:              "keeps existing AWS_REGION",
			awsRegion:         "us-west-2",
			awsDefaultRegion:  "us-east-1",
			expectedAWSRegion: "us-west-2",
		},
		{
			name:              "leaves AWS_REGION empty when no fallback exists",
			awsRegion:         "",
			awsDefaultRegion:  "",
			expectedAWSRegion: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("AWS_REGION", tc.awsRegion)
			t.Setenv("AWS_DEFAULT_REGION", tc.awsDefaultRegion)

			setAWSRegionFromDefault()

			assert.Equal(t, tc.expectedAWSRegion, os.Getenv("AWS_REGION"))
		})
	}
}

func TestSetAWSProfileToDefault(t *testing.T) {
	testCases := []struct {
		name               string
		awsProfile         string
		expectedAWSProfile string
	}{
		{
			name:               "uses default profile when AWS_PROFILE is unset",
			awsProfile:         "",
			expectedAWSProfile: "default",
		},
		{
			name:               "keeps existing AWS_PROFILE",
			awsProfile:         "ci",
			expectedAWSProfile: "ci",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("AWS_PROFILE", tc.awsProfile)

			setAWSProfileToDefault()

			assert.Equal(t, tc.expectedAWSProfile, os.Getenv("AWS_PROFILE"))
		})
	}
}

func TestMainExitCodeRejectsInvalidParallel(t *testing.T) {
	testCases := []struct {
		name     string
		parallel string
	}{
		{
			name:     "zero",
			parallel: "0",
		},
		{
			name:     "negative",
			parallel: "-1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			originalArgs := os.Args
			os.Args = []string{"terradozer", "-parallel", tc.parallel, "terraform.tfstate"}
			t.Cleanup(func() {
				os.Args = originalArgs
			})

			stderr := captureStderr(t, func() {
				assert.Equal(t, 1, mainExitCode())
			})

			assert.Contains(t, stderr, "-parallel flag must be greater than 0")
		})
	}
}

func TestMainExitCodeRejectsInvalidStateTimeout(t *testing.T) {
	originalArgs := os.Args
	os.Args = []string{"terradozer", "-state-timeout", "not-a-duration", "terraform.tfstate"}
	t.Cleanup(func() {
		os.Args = originalArgs
	})

	stderr := captureStderr(t, func() {
		assert.Equal(t, 1, mainExitCode())
	})

	assert.Contains(t, stderr, "failed to parse state-timeout flag")
}

func TestMainExitCodeDoesNotDefaultProfileBeforeStateRead(t *testing.T) {
	t.Setenv("AWS_ACCESS_KEY_ID", "access-key")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "secret-key")
	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_PROFILE", "")

	originalArgs := os.Args
	os.Args = []string{"terradozer", "s3://state-bucket/path/to/terraform.tfstate?versionId=123"}
	t.Cleanup(func() {
		os.Args = originalArgs
	})

	stderr := captureStderr(t, func() {
		assert.Equal(t, 1, mainExitCode())
	})

	assert.Contains(t, stderr, "must not include query or fragment components")
	assert.Empty(t, os.Getenv("AWS_PROFILE"))
}

func TestMainExitCodeRejectsInvalidRecursiveSourceBeforeProfileDefault(t *testing.T) {
	t.Setenv("AWS_PROFILE", "")

	originalArgs := os.Args
	os.Args = []string{"terradozer", "-recursive", filepath.Join(t.TempDir(), "missing")}
	t.Cleanup(func() {
		os.Args = originalArgs
	})

	stderr := captureStderr(t, func() {
		assert.Equal(t, 1, mainExitCode())
	})

	assert.Contains(t, stderr, "failed to discover Terraform state files")
	assert.Empty(t, os.Getenv("AWS_PROFILE"))
}

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	originalStderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	defer r.Close()

	os.Stderr = w
	defer func() {
		os.Stderr = originalStderr
	}()

	fn()

	require.NoError(t, w.Close())

	output, err := io.ReadAll(r)
	require.NoError(t, err)

	return string(output)
}

func TestInitProvidersSkipsUnsupportedProviders(t *testing.T) {
	p, err := initProviders([]string{"google"}, ".terradozer", 10*time.Second)
	assert.NoError(t, err)
	assert.Empty(t, p)
}

func TestAWSProviderConfig(t *testing.T) {
	t.Run("sets supported fields from environment", func(t *testing.T) {
		t.Setenv("AWS_ACCESS_KEY_ID", "key-id")
		t.Setenv("AWS_PROFILE", "ci")
		t.Setenv("AWS_REGION", "us-east-1")
		t.Setenv("AWS_SECRET_ACCESS_KEY", "secret")
		t.Setenv("AWS_CONFIG_FILE", "/tmp/config")
		t.Setenv("AWS_SHARED_CREDENTIALS_FILE", "/tmp/credentials")
		t.Setenv("AWS_SESSION_TOKEN", "token")

		config := awstoolsProvider.AWSProviderConfig().AsValueMap()

		assert.True(t, config["access_key"].RawEquals(cty.StringVal("key-id")))
		assert.True(t, config["profile"].RawEquals(cty.StringVal("ci")))
		assert.True(t, config["region"].RawEquals(cty.StringVal("us-east-1")))
		assert.True(t, config["secret_key"].RawEquals(cty.StringVal("secret")))
		assert.True(t, config["shared_config_files"].RawEquals(
			cty.ListVal([]cty.Value{cty.StringVal("/tmp/config")})))
		assert.True(t, config["shared_credentials_files"].RawEquals(
			cty.ListVal([]cty.Value{cty.StringVal("/tmp/credentials")})))
		assert.True(t, config["token"].RawEquals(cty.StringVal("token")))
	})

	t.Run("uses unknown defaults when optional values are unset", func(t *testing.T) {
		t.Setenv("AWS_ACCESS_KEY_ID", "")
		t.Setenv("AWS_PROFILE", "")
		t.Setenv("AWS_REGION", "")
		t.Setenv("AWS_SECRET_ACCESS_KEY", "")
		t.Setenv("AWS_CONFIG_FILE", "")
		t.Setenv("AWS_SHARED_CREDENTIALS_FILE", "")
		t.Setenv("AWS_SESSION_TOKEN", "")

		config := awstoolsProvider.AWSProviderConfig().AsValueMap()

		assert.True(t, config["access_key"].RawEquals(cty.StringVal("")))
		assert.True(t, config["profile"].RawEquals(cty.StringVal("")))
		assert.True(t, config["region"].RawEquals(cty.StringVal("")))
		assert.True(t, config["secret_key"].RawEquals(cty.StringVal("")))
		assert.True(t, config["token"].RawEquals(cty.StringVal("")))
		assert.False(t, config["shared_config_files"].IsKnown())
		assert.False(t, config["shared_credentials_files"].IsKnown())
	})
}

func TestAWSProviderV5BootstrapContract(t *testing.T) {
	assert.Equal(t, "v5.100.0", awstoolsProvider.AWSProviderVersion)

	requiredKeys := []string{
		"access_key",
		"allowed_account_ids",
		"assume_role",
		"assume_role_with_web_identity",
		"custom_ca_bundle",
		"default_tags",
		"ec2_metadata_service_endpoint",
		"ec2_metadata_service_endpoint_mode",
		"endpoints",
		"forbidden_account_ids",
		"http_proxy",
		"https_proxy",
		"ignore_tags",
		"insecure",
		"max_retries",
		"no_proxy",
		"retry_mode",
		"s3_us_east_1_regional_endpoint",
		"s3_use_path_style",
		"secret_key",
		"shared_config_files",
		"shared_credentials_files",
		"skip_credentials_validation",
		"skip_metadata_api_check",
		"skip_region_validation",
		"skip_requesting_account_id",
		"sts_region",
		"token",
		"token_bucket_rate_limiter_capacity",
		"use_dualstack_endpoint",
		"use_fips_endpoint",
	}

	config := awstoolsProvider.AWSProviderConfig().AsValueMap()
	for _, key := range requiredKeys {
		_, ok := config[key]
		assert.Truef(t, ok, "expected key %q in aws provider bootstrap config", key)
	}
}
