package state

import (
	"testing"

	"github.com/hashicorp/terraform/states"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
)

func TestDecodeResourceInstanceObject_UnknownAttributeCompatibility(t *testing.T) {
	schemaType := cty.Object(map[string]cty.Type{
		"id": cty.String,
		"default_action": cty.List(cty.Object(map[string]cty.Type{
			"type": cty.String,
		})),
	})

	resInstance := states.NewResourceInstance()
	resInstance.Current = &states.ResourceInstanceObjectSrc{
		AttrsJSON: []byte(`{
			"id":"listener-123",
			"default_action":[{"type":"forward"}],
			"mutual_authentication":[{"mode":"off"}]
		}`),
	}

	_, strictDecodeErr := resInstance.Current.Decode(schemaType)
	require.Error(t, strictDecodeErr)
	assert.Contains(t, strictDecodeErr.Error(), `unsupported attribute "mutual_authentication"`)

	resInstanceObj, filteredUnknownAttrs, err := decodeResourceInstanceObject(resInstance, schemaType)
	require.NoError(t, err)
	require.NotNil(t, resInstanceObj)
	assert.True(t, filteredUnknownAttrs)
	assert.Equal(t, cty.StringVal("listener-123"), resInstanceObj.Value.GetAttr("id"))

	attrs := resInstanceObj.Value.Type().AttributeTypes()
	_, hasMutualAuthentication := attrs["mutual_authentication"]
	assert.False(t, hasMutualAuthentication)
}

func TestPruneUnknownAttributesFromJSON_PreservesMapEntries(t *testing.T) {
	schemaType := cty.Object(map[string]cty.Type{
		"id":   cty.String,
		"tags": cty.Map(cty.String),
	})

	attrsJSON := []byte(`{
		"id":"listener-123",
		"tags":{"managed_by":"Terraform"},
		"mutual_authentication":[{"mode":"off"}]
	}`)

	prunedJSON, changed, err := pruneUnknownAttributesFromJSON(attrsJSON, schemaType)
	require.NoError(t, err)
	assert.True(t, changed)
	assert.JSONEq(t, `{
		"id":"listener-123",
		"tags":{"managed_by":"Terraform"}
	}`, string(prunedJSON))
}
