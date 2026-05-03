package terraform

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/apex/log"
	"github.com/chenrui333/terradozer/internal/awstools/aws"
	"github.com/chenrui333/terradozer/internal/awstools/terraform/provider"
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

func (p *providerPoolThreadSafe) set(key aws.ClientKey, pr provider.TerraformProvider) {
	p.Lock()
	p.providers[key] = pr
	p.Unlock()
}

func (p *providerPoolThreadSafe) closeAll(cause error) error {
	closeErrs := []error{cause}

	p.Lock()
	defer p.Unlock()

	for key, pr := range p.providers {
		closeErr := pr.Close()
		if closeErr != nil {
			closeErrs = append(closeErrs, fmt.Errorf("failed to close provider for key %v: %w", key, closeErr))
		}
	}

	return errors.Join(closeErrs...)
}

func providerConfigForClientKey(key aws.ClientKey) cty.Value {
	config := provider.AWSProviderConfig().AsValueMap()
	config["access_key"] = cty.UnknownVal(cty.DynamicPseudoType)
	config["secret_key"] = cty.UnknownVal(cty.DynamicPseudoType)
	config["token"] = cty.UnknownVal(cty.DynamicPseudoType)
	config[logFieldProfile] = cty.StringVal(key.Profile)
	config[logFieldRegion] = cty.StringVal(key.Region)

	return cty.ObjectVal(config)
}

func closeProviderAfterError(pr *provider.TerraformProvider, cause error) error {
	closeErr := pr.Close()
	if closeErr != nil {
		return errors.Join(cause, fmt.Errorf("failed to close provider: %w", closeErr))
	}

	return cause
}

func configureProviderForPool(pr *provider.TerraformProvider, config cty.Value, name, version string) error {
	err := pr.Configure(config)
	if err == nil {
		return nil
	}

	configureErr := fmt.Errorf("failed to configure provider (name=%s, version=%s): %w", name, version, err)
	return closeProviderAfterError(pr, configureErr)
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

	for _, clientKey := range clientKeys {
		key := clientKey
		wg.Go(func() {
			err := startupCtx.Err()
			if err != nil {
				recordErr(err)
				return
			}

			log.WithFields(log.Fields{
				logFieldProfile: key.Profile,
				logFieldRegion:  key.Region,
			}).Debugf("start launching new instance of Terraform AWS Provider")

			pr, err := provider.Launch(ctx, metaPlugin.Path, timeout)
			if err != nil {
				recordErr(fmt.Errorf("failed to launch provider (%s): %w", metaPlugin.Path, err))
				return
			}

			err = startupCtx.Err()
			if err != nil {
				recordErr(closeProviderAfterError(pr, err))
				return
			}

			err = configureProviderForPool(
				pr,
				providerConfigForClientKey(key),
				metaPlugin.Name,
				fmt.Sprint(metaPlugin.Version),
			)
			if err != nil {
				recordErr(err)
				return
			}

			err = startupCtx.Err()
			if err != nil {
				recordErr(closeProviderAfterError(pr, err))
				return
			}

			providerPool.set(key, *pr)

			log.WithFields(log.Fields{
				logFieldProfile: key.Profile,
				logFieldRegion:  key.Region,
			}).Debugf("launched new instance of Terraform AWS Provider")
		})
	}

	wg.Wait()

	if firstErr != nil {
		return nil, providerPool.closeAll(firstErr)
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
