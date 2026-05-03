package provider

import (
	"fmt"
	"os"

	"github.com/zclconf/go-cty/cty"
)

// AWSProviderVersion is the Terraform AWS provider version used by terradozer.
const AWSProviderVersion = "v5.100.0"

// config returns a default configuration for the Terraform Provider given by name (e.g. "aws").
func config(name string) (cty.Value, string, error) {
	switch name {
	case "aws":
		return AWSProviderConfig(), AWSProviderVersion, nil
	default:
		return cty.NilVal, "", fmt.Errorf("provider config not found: %s", name)
	}
}

// AWSProviderConfig returns a default configuration for the Terraform AWS Provider.
func AWSProviderConfig() cty.Value {
	config := map[string]cty.Value{
		"access_key":                         cty.StringVal(os.Getenv("AWS_ACCESS_KEY_ID")),
		"allowed_account_ids":                cty.UnknownVal(cty.DynamicPseudoType),
		"assume_role":                        cty.UnknownVal(cty.DynamicPseudoType),
		"assume_role_with_web_identity":      cty.UnknownVal(cty.DynamicPseudoType),
		"custom_ca_bundle":                   cty.UnknownVal(cty.DynamicPseudoType),
		"default_tags":                       cty.UnknownVal(cty.DynamicPseudoType),
		"ec2_metadata_service_endpoint":      cty.UnknownVal(cty.DynamicPseudoType),
		"ec2_metadata_service_endpoint_mode": cty.UnknownVal(cty.DynamicPseudoType),
		"endpoints":                          cty.UnknownVal(cty.DynamicPseudoType),
		"forbidden_account_ids":              cty.UnknownVal(cty.DynamicPseudoType),
		"http_proxy":                         cty.UnknownVal(cty.DynamicPseudoType),
		"https_proxy":                        cty.UnknownVal(cty.DynamicPseudoType),
		"ignore_tags":                        cty.UnknownVal(cty.DynamicPseudoType),
		"insecure":                           cty.UnknownVal(cty.DynamicPseudoType),
		"max_retries":                        cty.UnknownVal(cty.DynamicPseudoType),
		"no_proxy":                           cty.UnknownVal(cty.DynamicPseudoType),
		"profile":                            cty.StringVal(os.Getenv("AWS_PROFILE")),
		"region":                             cty.StringVal(os.Getenv("AWS_REGION")),
		"retry_mode":                         cty.UnknownVal(cty.DynamicPseudoType),
		"s3_us_east_1_regional_endpoint":     cty.UnknownVal(cty.DynamicPseudoType),
		"s3_use_path_style":                  cty.UnknownVal(cty.DynamicPseudoType),
		"secret_key":                         cty.StringVal(os.Getenv("AWS_SECRET_ACCESS_KEY")),
		"shared_config_files":                cty.UnknownVal(cty.DynamicPseudoType),
		"shared_credentials_files":           cty.UnknownVal(cty.DynamicPseudoType),
		"skip_credentials_validation":        cty.UnknownVal(cty.DynamicPseudoType),
		"skip_metadata_api_check":            cty.UnknownVal(cty.DynamicPseudoType),
		"skip_region_validation":             cty.UnknownVal(cty.DynamicPseudoType),
		"skip_requesting_account_id":         cty.UnknownVal(cty.DynamicPseudoType),
		"sts_region":                         cty.UnknownVal(cty.DynamicPseudoType),
		"token":                              cty.StringVal(os.Getenv("AWS_SESSION_TOKEN")),
		"token_bucket_rate_limiter_capacity": cty.UnknownVal(cty.DynamicPseudoType),
		"use_dualstack_endpoint":             cty.UnknownVal(cty.DynamicPseudoType),
		"use_fips_endpoint":                  cty.UnknownVal(cty.DynamicPseudoType),
	}

	if sharedCredentialsFile := os.Getenv("AWS_SHARED_CREDENTIALS_FILE"); sharedCredentialsFile != "" {
		config["shared_credentials_files"] = cty.ListVal([]cty.Value{cty.StringVal(sharedCredentialsFile)})
	}

	if sharedConfigFile := os.Getenv("AWS_CONFIG_FILE"); sharedConfigFile != "" {
		config["shared_config_files"] = cty.ListVal([]cty.Value{cty.StringVal(sharedConfigFile)})
	}

	return cty.ObjectVal(config)
}
