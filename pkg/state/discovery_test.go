package state

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoverSourcesFindsLocalTFStateFiles(t *testing.T) {
	tmpDir := t.TempDir()

	first := writeTestFile(t, tmpDir, "a/one.tfstate")
	second := writeTestFile(t, tmpDir, "b/two.tfstate")
	writeTestFile(t, tmpDir, "b/two.tfstate.backup")
	writeTestFile(t, tmpDir, "c/state.json")
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "d", "directory.tfstate"), 0o755))

	actual, err := DiscoverSources(tmpDir)

	require.NoError(t, err)
	assert.Equal(t, []string{first, second}, actual)
}

func TestDiscoverSourcesRejectsInvalidLocalRecursiveSources(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := writeTestFile(t, tmpDir, "terraform.tfstate")
	emptyDir := filepath.Join(tmpDir, "empty")
	require.NoError(t, os.Mkdir(emptyDir, 0o755))

	tests := []struct {
		name        string
		source      string
		expectedErr string
	}{
		{
			name:        "missing path",
			source:      filepath.Join(tmpDir, "missing"),
			expectedErr: "no such file or directory",
		},
		{
			name:        "file path",
			source:      stateFile,
			expectedErr: "must be a directory",
		},
		{
			name:        "no matches",
			source:      emptyDir,
			expectedErr: "no Terraform state files found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := DiscoverSources(tc.source)

			require.Error(t, err)
			assert.Nil(t, actual)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

func TestDiscoverSourcesFindsS3TFStateObjects(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		expectedBucket string
		expectedPrefix string
		expected       []string
	}{
		{
			name:           "bucket root",
			source:         "s3://state-bucket",
			expectedBucket: "state-bucket",
			expectedPrefix: "",
			expected: []string{
				"s3://state-bucket/a.tfstate",
				"s3://state-bucket/nested/b.tfstate",
				"s3://state-bucket/z.tfstate",
			},
		},
		{
			name:           "prefix without trailing slash",
			source:         "s3://state-bucket/path/to/states",
			expectedBucket: "state-bucket",
			expectedPrefix: "path/to/states/",
			expected: []string{
				"s3://state-bucket/path/to/states/a.tfstate",
				"s3://state-bucket/path/to/states/nested/b.tfstate",
			},
		},
		{
			name:           "prefix with trailing slash",
			source:         "s3://state-bucket/path/to/states/",
			expectedBucket: "state-bucket",
			expectedPrefix: "path/to/states/",
			expected: []string{
				"s3://state-bucket/path/to/states/a.tfstate",
				"s3://state-bucket/path/to/states/nested/b.tfstate",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := discoverSourcesWithS3ObjectLister(context.Background(), tc.source,
				func(ctx context.Context, bucket, prefix string) ([]string, error) {
					assert.Equal(t, tc.expectedBucket, bucket)
					assert.Equal(t, tc.expectedPrefix, prefix)

					return []string{
						prefix + "nested/b.tfstate",
						prefix + "skip.txt",
						prefix + "a.tfstate",
						prefix + "terraform.tfstate.backup",
						"z.tfstate",
					}, nil
				})

			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestDiscoverSourcesReturnsS3ListErrors(t *testing.T) {
	actual, err := discoverSourcesWithS3ObjectLister(context.Background(), "s3://state-bucket/path/",
		func(ctx context.Context, bucket, prefix string) ([]string, error) {
			return nil, errors.New("access denied")
		})

	require.Error(t, err)
	assert.Nil(t, actual)
	assert.Contains(t, err.Error(), "access denied")
}

func TestDiscoverSourcesReturnsS3ContextErrors(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	actual, err := discoverSourcesWithS3ObjectLister(ctx, "s3://state-bucket/path/",
		func(ctx context.Context, bucket, prefix string) ([]string, error) {
			return nil, ctx.Err()
		})

	require.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, actual)
}

func TestDiscoverSourcesRejectsInvalidS3RecursiveSources(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		expectedErr string
	}{
		{
			name:        "missing bucket",
			source:      "s3:///path/to/states/",
			expectedErr: "must include a bucket",
		},
		{
			name:        "embedded credentials",
			source:      "s3://access:secret@state-bucket/path/to/states/",
			expectedErr: "must not include embedded credentials",
		},
		{
			name:        "query component",
			source:      "s3://state-bucket/path/to/states/?versionId=123",
			expectedErr: "must not include query or fragment",
		},
		{
			name:        "fragment component",
			source:      "s3://state-bucket/path/to/states/#version",
			expectedErr: "must not include query or fragment",
		},
		{
			name:        "malformed credentialed URI",
			source:      "s3://access:secret@state-bucket/path%zz",
			expectedErr: "must not include embedded credentials",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := DiscoverSourcesWithContext(context.Background(), tc.source)

			require.Error(t, err)
			assert.Nil(t, actual)
			assert.Contains(t, err.Error(), tc.expectedErr)
			assert.NotContains(t, err.Error(), "secret")
		})
	}
}

func TestDiscoverSourcesReturnsErrorWhenS3PrefixHasNoMatches(t *testing.T) {
	actual, err := discoverSourcesWithS3ObjectLister(context.Background(), "s3://state-bucket/path/",
		func(ctx context.Context, bucket, prefix string) ([]string, error) {
			return []string{"path/state.json", "path/terraform.tfstate.backup"}, nil
		})

	require.Error(t, err)
	assert.Nil(t, actual)
	assert.Contains(t, err.Error(), "no Terraform state files found")
}

func writeTestFile(t *testing.T, baseDir, relPath string) string {
	t.Helper()

	path := filepath.Join(baseDir, filepath.FromSlash(relPath))
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte("{}"), 0o644))

	return path
}
