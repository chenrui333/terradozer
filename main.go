package main

//nolint:lll
//go:generate mockgen -source=pkg/resource/destroy.go -destination=pkg/resource/destroy_mock_test.go -package=resource_test

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	stdlog "log"
	"os"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/chenrui333/terradozer/internal"
	"github.com/chenrui333/terradozer/internal/awstools/terraform"
	"github.com/chenrui333/terradozer/internal/awstools/terraform/provider"
	"github.com/chenrui333/terradozer/pkg/resource"
	"github.com/chenrui333/terradozer/pkg/state"
	"github.com/fatih/color"
)

func main() {
	os.Exit(mainExitCode())
}

//nolint:wsl
func mainExitCode() int {
	var dryRun bool
	var force bool
	var logDebug bool
	var parallel int
	var timeout string
	var version bool

	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	flags.Usage = func() {
		printHelp(flags)
	}

	flags.StringVar(&timeout, "timeout", "30s", "Amount of time to wait for a destroy of a resource to finish")
	flags.BoolVar(&dryRun, "dry-run", false, "Show what would be destroyed")
	flags.BoolVar(&force, "force", false, "Destroy without asking for confirmation")
	flags.BoolVar(&logDebug, "debug", false, "Enable debug logging")
	flags.IntVar(&parallel, "parallel", 10, "Limit the number of concurrent destroy operations")
	flags.BoolVar(&version, "version", false, "Show application version")

	_ = flags.Parse(os.Args[1:])
	args := flags.Args()

	log.SetHandler(cli.Default)

	fmt.Println()
	defer fmt.Println()

	if logDebug {
		log.SetLevel(log.DebugLevel)
	}

	// discard TRACE logs of GRPCProvider
	stdlog.SetOutput(io.Discard)

	if version {
		fmt.Println(internal.BuildVersionString())
		return 0
	}

	if force && dryRun {
		fmt.Fprint(os.Stderr, color.RedString("Error:️ -force and -dry-run flag cannot be used together\n"))
		printHelp(flags)

		return 1
	}

	if parallel < 1 {
		fmt.Fprint(os.Stderr, color.RedString("Error: -parallel flag must be greater than 0\n"))
		printHelp(flags)

		return 1
	}

	timeoutDuration, err := time.ParseDuration(timeout)
	if err != nil {
		fmt.Fprint(os.Stderr, color.RedString("Error: failed to parse timeout flag: %s\n", err))
		printHelp(flags)

		return 1
	}

	if len(args) == 0 {
		fmt.Fprint(os.Stderr, color.RedString("Error: path to Terraform state file expected\n"))
		printHelp(flags)

		return 1
	}

	pathToState := args[0]

	setAWSRegionFromDefault()

	tfstate, err := state.New(pathToState)
	if err != nil {
		fmt.Fprint(os.Stderr, color.RedString("Error:️ failed to read Terraform state file: %s\n", err))

		return 1
	}

	internal.LogTitle("reading state")
	log.WithField("source", pathToState).Info(internal.Pad("using state"))

	setAWSProfileToDefault()

	providers, err := initProviders(tfstate.ProviderNames(), "~/.terradozer", timeoutDuration)
	if err != nil {
		fmt.Fprint(os.Stderr, color.RedString("\nError:️ failed to initialize Terraform providers: %s\n", err))

		return 1
	}

	defer func() {
		for _, p := range providers {
			_ = p.Close()
		}
	}()

	resources, err := tfstate.Resources(providers)
	if err != nil {
		fmt.Fprint(os.Stderr, color.RedString("\nError:️ failed to get resources from Terraform state: %s\n", err))

		return 1
	}

	resourcesWithUpdatedState := terraform.UpdateResources(resources, parallel)

	if !force {
		internal.LogTitle("showing resources that would be deleted (dry run)")

		// always show the resources that would be affected before deleting anything
		for _, r := range resourcesWithUpdatedState {
			log.WithField("id", r.ID()).Warn(internal.Pad(r.Type()))
		}

		if len(resourcesWithUpdatedState) == 0 {
			internal.LogTitle("all resources have already been deleted")
			return 0
		}

		internal.LogTitle(fmt.Sprintf("total number of resources that would be deleted: %d",
			len(resourcesWithUpdatedState)))
	}

	if !dryRun {
		if !internal.UserConfirmedDeletion(os.Stdin, force) {
			return 0
		}

		internal.LogTitle("Starting to delete resources")

		numDeletedResources := resource.DestroyResources(
			convertToDestroyableResources(resourcesWithUpdatedState), parallel)

		internal.LogTitle(fmt.Sprintf("total number of deleted resources: %d", numDeletedResources))
	}

	return 0
}

func convertToDestroyableResources(resources []terraform.UpdatableResource) []resource.DestroyableResource {
	var result []resource.DestroyableResource

	for _, r := range resources {
		result = append(result, r.(resource.DestroyableResource))
	}

	return result
}

func setAWSRegionFromDefault() {
	if os.Getenv("AWS_REGION") != "" {
		return
	}

	defaultRegion := os.Getenv("AWS_DEFAULT_REGION")
	if defaultRegion == "" {
		return
	}

	_ = os.Setenv("AWS_REGION", defaultRegion)
}

func setAWSProfileToDefault() {
	if os.Getenv("AWS_PROFILE") != "" {
		return
	}

	_ = os.Setenv("AWS_PROFILE", "default")
}

func initProviders(providerNames []string, installDir string,
	timeout time.Duration) (map[string]*provider.TerraformProvider, error) {
	providers := map[string]*provider.TerraformProvider{}

	for _, providerName := range providerNames {
		if providerName != "aws" {
			log.WithField("name", providerName).Debug("ignoring resources of (yet) unsupported provider")
			continue
		}

		p, err := initAWSProvider(installDir, timeout)
		if err != nil {
			closeErrs := []error{err}
			for name, started := range providers {
				closeErr := started.Close()
				if closeErr != nil {
					closeErrs = append(closeErrs, fmt.Errorf("failed to close provider %s: %w", name, closeErr))
				}
			}

			return nil, errors.Join(closeErrs...)
		}

		providers[providerName] = p
	}

	return providers, nil
}

func initAWSProvider(installDir string, timeout time.Duration) (*provider.TerraformProvider, error) {
	metaPlugin, err := provider.Install("aws", provider.AWSProviderVersion, installDir)
	if err != nil {
		return nil, fmt.Errorf("failed to install provider (aws): %w", err)
	}

	p, err := provider.Launch(context.Background(), metaPlugin.Path, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to launch provider (%s): %w", metaPlugin.Path, err)
	}

	err = p.Configure(provider.AWSProviderConfig())
	if err != nil {
		configureErr := fmt.Errorf("failed to configure provider (name=%s, version=%s): %w",
			metaPlugin.Name, metaPlugin.Version, err)
		closeErr := p.Close()
		if closeErr != nil {
			return nil, fmt.Errorf("%w; failed to close provider: %w", configureErr, closeErr)
		}

		return nil, configureErr
	}

	log.WithFields(log.Fields{
		"name":    metaPlugin.Name,
		"version": metaPlugin.Version,
	}).Debug("configured provider")

	return p, nil
}

func printHelp(fs *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, "\n%s\n", strings.TrimSpace(help))
	fs.PrintDefaults()
	fmt.Println()
}

const help = `
Terraform destroy using only the state - no *.tf files needed.

USAGE:
  $ terradozer [flags] <path/to/terraform.tfstate|s3://bucket/key>

FLAGS:
`
