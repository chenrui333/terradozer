package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetAWSRegionFromDefault(t *testing.T) {
	testCases := []struct {
		name              string
		awsRegion         string
		awsDefaultRegion  string
		expectedAWSRegion string
	}{
		{
			name:              "uses AWS_DEFAULT_REGION when AWS_REGION is unset",
			awsRegion:         "",
			awsDefaultRegion:  "us-east-1",
			expectedAWSRegion: "us-east-1",
		},
		{
			name:              "keeps existing AWS_REGION",
			awsRegion:         "us-west-2",
			awsDefaultRegion:  "us-east-1",
			expectedAWSRegion: "us-west-2",
		},
		{
			name:              "leaves AWS_REGION empty when no fallback exists",
			awsRegion:         "",
			awsDefaultRegion:  "",
			expectedAWSRegion: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("AWS_REGION", tc.awsRegion)
			t.Setenv("AWS_DEFAULT_REGION", tc.awsDefaultRegion)

			setAWSRegionFromDefault()

			assert.Equal(t, tc.expectedAWSRegion, os.Getenv("AWS_REGION"))
		})
	}
}

func TestSetAWSProfileToDefault(t *testing.T) {
	testCases := []struct {
		name               string
		awsProfile         string
		expectedAWSProfile string
	}{
		{
			name:               "uses default profile when AWS_PROFILE is unset",
			awsProfile:         "",
			expectedAWSProfile: "default",
		},
		{
			name:               "keeps existing AWS_PROFILE",
			awsProfile:         "ci",
			expectedAWSProfile: "ci",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("AWS_PROFILE", tc.awsProfile)

			setAWSProfileToDefault()

			assert.Equal(t, tc.expectedAWSProfile, os.Getenv("AWS_PROFILE"))
		})
	}
}
