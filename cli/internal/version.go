package internal

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

//nolint:gochecknoglobals
var (
	version       = "dev"
	commit        = "?"
	date          = "?"
	readBuildInfo = debug.ReadBuildInfo
)

func BuildVersionString() string {
	currentVersion, currentCommit, currentDate := buildMetadata()

	return fmt.Sprintf(
		"version: %s\ncommit: %s\nbuilt at: %s\nusing: %s",
		currentVersion, currentCommit, currentDate, runtime.Version(),
	)
}

func buildMetadata() (string, string, string) {
	currentVersion := normalizeValue(version, "dev")
	currentCommit := normalizeValue(commit, "?")
	currentDate := normalizeValue(date, "?")

	settings := buildInfoSettings()

	if isUnsetValue(currentCommit) {
		currentCommit = normalizeValue(settings["vcs.revision"], currentCommit)
	}

	if isUnsetValue(currentDate) {
		currentDate = normalizeValue(settings["vcs.time"], currentDate)
	}

	return currentVersion, currentCommit, currentDate
}

func buildInfoSettings() map[string]string {
	info, ok := readBuildInfo()
	if !ok || info == nil {
		return nil
	}

	settings := make(map[string]string, len(info.Settings))
	for _, setting := range info.Settings {
		settings[setting.Key] = setting.Value
	}

	return settings
}

func normalizeValue(value, fallback string) string {
	if isUnsetValue(value) {
		return fallback
	}

	return strings.TrimSpace(value)
}

func isUnsetValue(value string) bool {
	trimmedValue := strings.TrimSpace(value)

	return trimmedValue == "" || trimmedValue == "?"
}
