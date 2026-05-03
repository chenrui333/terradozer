package state

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseS3StateSource(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		expected    s3StateSource
		expectedS3  bool
		expectedErr string
	}{
		{
			name:   "local path",
			source: "terraform.tfstate",
		},
		{
			name:       "S3 path",
			source:     "s3://state-bucket/path/to/terraform.tfstate",
			expected:   s3StateSource{bucket: "state-bucket", key: "path/to/terraform.tfstate"},
			expectedS3: true,
		},
		{
			name:        "missing bucket",
			source:      "s3:///path/to/terraform.tfstate",
			expectedS3:  true,
			expectedErr: "must include a bucket",
		},
		{
			name:        "missing key",
			source:      "s3://state-bucket",
			expectedS3:  true,
			expectedErr: "must include a key",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual, isS3, err := parseS3StateSource(tc.source)

			assert.Equal(t, tc.expectedS3, isS3)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestNewReadsS3StateSource(t *testing.T) {
	stateData, err := os.ReadFile("../../test/test-fixtures/tfstates/version4.tfstate")
	require.NoError(t, err)

	actualState, err := newWithS3ObjectReader("s3://state-bucket/path/to/terraform.tfstate",
		func(ctx context.Context, bucket, key string) ([]byte, error) {
			assert.Equal(t, "state-bucket", bucket)
			assert.Equal(t, "path/to/terraform.tfstate", key)

			return stateData, nil
		})
	require.NoError(t, err)
	require.NotNil(t, actualState)
	assert.Equal(t, []string{"aws"}, actualState.ProviderNames())
}

func TestNewReturnsS3ReadError(t *testing.T) {
	actualState, err := newWithS3ObjectReader("s3://state-bucket/path/to/terraform.tfstate",
		func(ctx context.Context, bucket, key string) ([]byte, error) {
			return nil, errors.New("access denied")
		})

	require.Error(t, err)
	assert.Nil(t, actualState)
	assert.Contains(t, err.Error(), "access denied")
}
