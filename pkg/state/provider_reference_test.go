package state

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeProviderReference(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		expected   string
		isModified bool
	}{
		{
			name:       "legacy provider reference",
			input:      "provider.aws",
			expected:   "provider.aws",
			isModified: false,
		},
		{
			name:       "Terraform registry provider reference",
			input:      "provider[\"registry.terraform.io/hashicorp/aws\"]",
			expected:   "provider.aws",
			isModified: true,
		},
		{
			name:       "Terraform registry provider reference with alias",
			input:      "provider[\"registry.terraform.io/hashicorp/aws\"].use1",
			expected:   "provider.aws.use1",
			isModified: true,
		},
		{
			name:       "module scoped Terraform registry provider reference",
			input:      "module.echo.provider[\"registry.terraform.io/hashicorp/aws\"]",
			expected:   "module.echo.provider.aws",
			isModified: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual, isModified := normalizeProviderReference(tc.input)
			assert.Equal(t, tc.expected, actual)
			assert.Equal(t, tc.isModified, isModified)
		})
	}
}

func TestNormalizeProviderReferences(t *testing.T) {
	stateData, err := os.ReadFile("../../test/test-fixtures/tfstates/version4-tf19-provider.json")
	require.NoError(t, err)

	normalizedStateData, changed := normalizeProviderReferences(stateData)
	require.True(t, changed)

	stateFile, err := readStateFile(normalizedStateData)
	require.NoError(t, err)
	require.NotNil(t, stateFile.State)
	assert.Equal(t, 1, len(stateFile.State.ProviderAddrs()))
}
