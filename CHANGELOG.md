# Changelog

## v0.3.1 - 2026-05-28

### Fixed

- Add API Gateway destroy priority chain to prevent stage deletion failures caused by method_settings dependency ordering.

## v0.3.0 - 2026-05-04

### Added

- Support reading Terraform state directly from S3 paths.
- Add recursive discovery for local directories and S3 prefixes.
- Add AWS resource destroy priority ordering to reduce dependency retry loops.
- Publish release artifacts with GitHub Artifact Attestations.

### Changed

- Consolidate the awstools source into the main tree and modernize the Go module setup.
- Update dependency security posture, including go-getter, OpenTelemetry, gRPC, and go-jose.
- Improve CI naming and Go module caching.

### Fixed

- Reject invalid non-positive destroy parallelism values.
- Harden S3 state URI parsing, credential rejection, and state read/discovery timeouts.
- Avoid provider/client pool lifecycle leaks and worker error races.

