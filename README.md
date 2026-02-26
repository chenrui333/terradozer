<p align="center">
  <img alt="terradozer" src="https://github.com/jckuester/terradozer/blob/master/img/logo.png" height="180" />
  <h3 align="center">terradozer</h3>
  <p align="center">Terraform destroy using the state only - no *.tf files needed</p>
</p>

---
[![Release](https://img.shields.io/github/release/chenrui333/terradozer.svg?style=for-the-badge)](https://github.com/chenrui333/terradozer/releases/latest)
[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=for-the-badge)](/LICENSE.md)
[![CI](https://img.shields.io/github/actions/workflow/status/chenrui333/terradozer/ci.yml?style=for-the-badge)](https://github.com/chenrui333/terradozer/actions/workflows/ci.yml)
[![Codecov branch](https://img.shields.io/codecov/c/github/jckuester/terradozer/master.svg?style=for-the-badge)](https://codecov.io/gh/jckuester/terradozer)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=for-the-badge)](http://godoc.org/github.com/jckuester/terradozer)

Terradozer takes a Terraform state file as input and destroys all resources it finds in it - without needing any *.tf
files. This works currently only for resources of the Terraform AWS Provider. If you need support for any other provider,
let me know, and I will try to help.

Happy (terra)dozing!

## Example

![](img/example.gif)

## Features

* Nothing will be deleted without your confirmation. Terradozer always lists all resources first and then waits for
  your approval
* Using the `-force` flag (dangerous!), terradozer can run in an automated fashion without human interaction and approval,
  for example, as part of your CI pipeline
* **Planned**, if you want me to implement this, [please upvote](https://github.com/jckuester/terradozer/issues/9):
  Allow terradozer pointing directly to a state file stored in S3, i.e., `terradozer s3://path/to/terraform.tfstate`
* **Planned**, if you want me to implement this, [please upvote](https://github.com/jckuester/terradozer/issues/8):
  A `-recursive` flag to delete resources of all states found under a given directory, i.e.,
  `terradozer -recursive s3://bucket-with-states/`. This is especially helpful if
  you orchestrate Terraform modules with [Terragrunt](https://github.com/gruntwork-io/terragrunt) and store all states
  under the same directory or in the same S3 bucket. This way, a complete Terragrunt project could be cleaned up in an
  automated fashion.

## Installation

It's recommended to install a specific version of terradozer available on the
[releases page](https://github.com/chenrui333/terradozer/releases).

GoReleaser publishes release assets with these names:

- `terradozer_<version>_<os>_<arch>.tar.gz`
- `terradozer_<version>_checksums.txt`

Supported release targets:

- `os`: `darwin`, `linux`
- `arch`: `amd64`, `arm64`

Here is the recommended way to install a specific version (example: `v0.1.2`):

```bash
# install it into ./bin/
curl -sSfL https://raw.githubusercontent.com/chenrui333/terradozer/main/install.sh | sh -s v0.1.2
```

## Usage

To delete all resources in a Terraform state file:

    terradozer [flags] <path/to/terraform.tfstate>

To see all options, run `terradozer --help`. Provide credentials for the AWS account you want to destroy resources in
via the usual [environment variables](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html), e.g.,
`AWS_PROFILE=<myaccount>` and either `AWS_REGION=<myregion>` or `AWS_DEFAULT_REGION=<myregion>`.
If `AWS_PROFILE` is unset, terradozer uses the `default` profile.

The region information is needed as it is not stored as part of the state. Having multiple providers with different
regions in one state file is not yet supported.

### State file format

Terradozer expects a valid Terraform state JSON document (the same content format as `terraform.tfstate`).

- The file extension is not used for detection; parsing is content-based.
- Common names like `terraform.tfstate`, `*.json`, and `*.tfstate.json` are all supported.
- The file must contain Terraform-managed resources in state format (unsupported or malformed JSON will fail to parse).
 
## How it works

Terradozer first scans a given Terraform state file (read-only) to find all resources (excluding data sources),
then downloads the necessary Terraform Provider Plugins to call the destroy function for each resource on the respective
CRUD API via GRPC (e.g., calling the Terraform AWS Provider to destroy a `aws_instance` resource).

## Dependency updates

This repository uses Renovate with config at `.github/renovate.json`.

- Go modules (`go.mod`/`go.sum`) and GitHub Actions dependencies are monitored.
- Minor/patch updates are grouped by manager and configured for auto-merge.
- Major updates are isolated and require manual dashboard approval/review.
- Renovate uses semantic commit messages (`chore(deps): ...`) and applies dependency labels.

## Tests

This section is only relevant if you want to contribute to Terradozer and therefore run the tests. Terradozer has
acceptance tests, integration tests checking against changes of behaviour in the Terraform Provider API, and of course
unit tests.

Run unit tests

    make test
    
Run acceptance and integration tests

    AWS_PROFILE=<myaccount> AWS_REGION=<myregion> make test-all
