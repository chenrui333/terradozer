package terraform

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/apex/log"
	"github.com/jckuester/terradozer/internal/awstools/aws"
	"github.com/jckuester/terradozer/internal/awstools/terraform/provider"
	"github.com/zclconf/go-cty/cty"
)

const (
	logFieldID      = "id"
	logFieldProfile = "profile"
	logFieldRegion  = "region"
	logFieldType    = "type"
)

// providerPoolThreadSafe is a concurrent map implementation to store multiple Terraform AWS Providers.
type providerPoolThreadSafe struct {
	sync.Mutex

	providers map[aws.ClientKey]provider.TerraformProvider
}

// NewProviderPool launches a set of Terraform AWS Providers with the configuration of the given clientKeys
// (combination of AWS profile and region).
// Providers are launched only once in case of duplicate clientKeys.
func NewProviderPool(ctx context.Context, clientKeys []aws.ClientKey, version, installDir string,
	timeout time.Duration) (
	map[aws.ClientKey]provider.TerraformProvider, error) {
	metaPlugin, err := provider.Install("aws", version, installDir)
	if err != nil {
		return nil, fmt.Errorf("failed to install provider (%s): %w", "aws", err)
	}

	// startupCtx coordinates pool creation without owning launched provider lifetimes.
	startupCtx, cancelStartup := context.WithCancel(ctx)
	defer cancelStartup()

	var wg sync.WaitGroup
	var errOnce sync.Once
	var firstErr error

	recordErr := func(err error) {
		if err == nil {
			return
		}

		errOnce.Do(func() {
			firstErr = err
			cancelStartup()
		})
	}

	providerPool := &providerPoolThreadSafe{
		providers: make(map[aws.ClientKey]provider.TerraformProvider),
	}

	clientKeys = removeDuplicateClientKeys(clientKeys)

	if len(clientKeys) > 0 {
		for _, clientKey := range clientKeys {
			p := clientKey.Profile
			r := clientKey.Region
			wg.Go(func() {
				err := startupCtx.Err()
				if err != nil {
					recordErr(err)
					return
				}

				log.WithFields(log.Fields{
					logFieldProfile: p,
					logFieldRegion:  r,
				}).Debugf("start launching new instance of Terraform AWS Provider")

				pr, err := provider.Launch(ctx, metaPlugin.Path, timeout)
				if err != nil {
					recordErr(fmt.Errorf("failed to launch provider (%s): %w", metaPlugin.Path, err))
					return
				}

				config := cty.ObjectVal(map[string]cty.Value{
					logFieldProfile:               cty.StringVal(p),
					logFieldRegion:                cty.StringVal(r),
					"access_key":                  cty.UnknownVal(cty.DynamicPseudoType),
					"allowed_account_ids":         cty.UnknownVal(cty.DynamicPseudoType),
					"assume_role":                 cty.UnknownVal(cty.DynamicPseudoType),
					"default_tags":                cty.UnknownVal(cty.DynamicPseudoType),
					"endpoints":                   cty.UnknownVal(cty.DynamicPseudoType),
					"forbidden_account_ids":       cty.UnknownVal(cty.DynamicPseudoType),
					"ignore_tag_prefixes":         cty.UnknownVal(cty.DynamicPseudoType),
					"ignore_tags":                 cty.UnknownVal(cty.DynamicPseudoType),
					"insecure":                    cty.UnknownVal(cty.DynamicPseudoType),
					"max_retries":                 cty.UnknownVal(cty.DynamicPseudoType),
					"s3_force_path_style":         cty.UnknownVal(cty.DynamicPseudoType),
					"secret_key":                  cty.UnknownVal(cty.DynamicPseudoType),
					"shared_credentials_file":     cty.UnknownVal(cty.DynamicPseudoType),
					"skip_credentials_validation": cty.UnknownVal(cty.DynamicPseudoType),
					"skip_get_ec2_platforms":      cty.UnknownVal(cty.DynamicPseudoType),
					"skip_metadata_api_check":     cty.UnknownVal(cty.DynamicPseudoType),
					"skip_region_validation":      cty.UnknownVal(cty.DynamicPseudoType),
					"skip_requesting_account_id":  cty.UnknownVal(cty.DynamicPseudoType),
					"token":                       cty.UnknownVal(cty.DynamicPseudoType),
				})

				err = pr.Configure(config)
				if err != nil {
					_ = pr.Close()

					recordErr(fmt.Errorf("failed to configure provider (name=%s, version=%s): %w",
						metaPlugin.Name, metaPlugin.Version, err))

					return
				}

				err = startupCtx.Err()
				if err != nil {
					_ = pr.Close()
					recordErr(err)
					return
				}

				providerPool.Lock()
				providerPool.providers[aws.ClientKey{Profile: p, Region: r}] = *pr
				providerPool.Unlock()

				log.WithFields(log.Fields{
					logFieldProfile: p,
					logFieldRegion:  r,
				}).Debugf("launched new instance of Terraform AWS Provider")
			})
		}
	}

	wg.Wait()

	if firstErr != nil {
		providerPool.Lock()
		for _, p := range providerPool.providers {
			_ = p.Close()
		}
		providerPool.Unlock()

		return nil, firstErr
	}

	return providerPool.providers, nil
}

func removeDuplicateClientKeys(clientKeys []aws.ClientKey) []aws.ClientKey {
	seen := make(map[aws.ClientKey]bool)
	var result []aws.ClientKey

	for _, clientKey := range clientKeys {
		if _, ok := seen[clientKey]; !ok {
			seen[clientKey] = true
			result = append(result, clientKey)
		}
	}

	return result
}
