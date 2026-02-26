package internal

import (
	"runtime"
)

//nolint:gochecknoglobals
var (
	version = "dev"
	commit  = "?"
	date    = "?"
)

func BuildVersionString() string {
	result := "version: " + version

	if commit != "" {
		result += "\ncommit: " + commit
	}

	if date != "" {
		result += "\nbuilt at: " + date
	}

	result += "\nusing: " + runtime.Version()

	return result
}
