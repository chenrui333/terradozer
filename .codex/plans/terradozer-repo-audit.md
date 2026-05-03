# Terradozer Repo Audit And Backlog Plan

## Summary

- Start from current `origin/main` and keep changes narrow enough for review.
- Fix repo-health blockers before feature work: generated mocks, local checks, lint drift, and dependency automation.
- Keep destroy behavior changes issue-backed and covered by unit tests.

## Key Changes

- Tooling baseline:
  - Migrate `github.com/golang/mock` to `go.uber.org/mock`.
  - Check in `pkg/resource/destroy_mock_test.go` so `go test -short ./...` works from a clean checkout.
  - Align local and CI golangci-lint versions, and fix Makefile phony targets.
- Dependency and PR triage:
  - Patch the golangci-lint v2.12.x failure with a small code cleanup.
  - Keep Terraform-core compatibility packages out of the broad Renovate non-major automerge group.
  - Treat stale broad dependency and monorepo PRs as superseded after replacement checks exist.
- Destroy behavior:
  - Guard invalid parallelism.
  - Filter already-deleted resources before display and destroy.
  - Add bounded zero-progress retry rounds with delay.
  - Apply static AWS destroy priority ordering before the destroy phase.

## Public Interfaces

- Add `--max-retry-rounds` and `--retry-delay` CLI flags.
- Keep the existing `DestroyResources(resources, parallel)` helper as the compatibility wrapper.
- Add an options-based destroy helper for callers that need retry controls.

## Test Plan

- `go generate ./...`
- `git diff --exit-code` after generation when checking generated output stability
- `go test -short ./...`
- `go vet ./...`
- `golangci-lint run`

## Assumptions

- AWS-backed acceptance tests are out of scope unless teardown-safe credentials and fixtures are explicitly provided.
- Dependency PR metadata can be cleaned up after the replacement branch is green.
- JSON output, agent mode, and ARN input should wait until result accounting is explicit.
