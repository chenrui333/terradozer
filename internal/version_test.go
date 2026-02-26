package internal

import (
	"runtime"
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildVersionString_UsesRuntimeBuildInfoWhenLdflagsAreUnset(t *testing.T) {
	resetBuildMetadata(t)

	version = "dev"
	commit = "?"
	date = "?"
	readBuildInfo = func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{
			Settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "abc123def456"},
				{Key: "vcs.time", Value: "2026-02-26T00:00:00Z"},
			},
		}, true
	}

	assert.Equal(
		t,
		"version: dev\ncommit: abc123def456\nbuilt at: 2026-02-26T00:00:00Z\nusing: "+runtime.Version(),
		BuildVersionString(),
	)
}

func TestBuildVersionString_PrefersLdflagsValuesOverRuntimeBuildInfo(t *testing.T) {
	resetBuildMetadata(t)

	version = "v0.2.0"
	commit = "fa1f598"
	date = "2026-02-26T20:00:00Z"
	readBuildInfo = func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{
			Settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "should-not-be-used"},
				{Key: "vcs.time", Value: "1970-01-01T00:00:00Z"},
			},
		}, true
	}

	assert.Equal(
		t,
		"version: v0.2.0\ncommit: fa1f598\nbuilt at: 2026-02-26T20:00:00Z\nusing: "+runtime.Version(),
		BuildVersionString(),
	)
}

func TestBuildVersionString_FallsBackToDefaultsWhenBuildInfoUnavailable(t *testing.T) {
	resetBuildMetadata(t)

	version = "dev"
	commit = "?"
	date = "?"
	readBuildInfo = func() (*debug.BuildInfo, bool) {
		return nil, false
	}

	assert.Equal(
		t,
		"version: dev\ncommit: ?\nbuilt at: ?\nusing: "+runtime.Version(),
		BuildVersionString(),
	)
}

func resetBuildMetadata(t *testing.T) {
	t.Helper()

	previousVersion := version
	previousCommit := commit
	previousDate := date
	previousReadBuildInfo := readBuildInfo

	t.Cleanup(func() {
		version = previousVersion
		commit = previousCommit
		date = previousDate
		readBuildInfo = previousReadBuildInfo
	})
}
