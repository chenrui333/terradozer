package terraform

import (
	"testing"

	"github.com/chenrui333/terradozer/internal/awstools/aws"
	"github.com/stretchr/testify/assert"
	"github.com/zclconf/go-cty/cty"
)

func TestProviderConfigForClientKeyDoesNotInheritStaticCredentials(t *testing.T) {
	t.Setenv("AWS_ACCESS_KEY_ID", "env-access-key")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "env-secret-key")
	t.Setenv("AWS_SESSION_TOKEN", "env-session-token")
	t.Setenv("AWS_PROFILE", "env-profile")
	t.Setenv("AWS_REGION", "us-east-1")

	config := providerConfigForClientKey(aws.ClientKey{
		Profile: "target-profile",
		Region:  "us-west-2",
	}).AsValueMap()

	assert.True(t, config[logFieldProfile].RawEquals(cty.StringVal("target-profile")))
	assert.True(t, config[logFieldRegion].RawEquals(cty.StringVal("us-west-2")))
	assert.False(t, config["access_key"].IsWhollyKnown())
	assert.False(t, config["secret_key"].IsWhollyKnown())
	assert.False(t, config["token"].IsWhollyKnown())
}
