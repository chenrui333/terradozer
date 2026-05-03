package terraform

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
)

func TestUpdateWorkerHandlesNilState(t *testing.T) {
	resources := make(chan UpdatableResource, 1)
	results := make(chan updateWorkerResult, 1)
	resources <- &updateWorkerTestResource{}
	close(resources)

	updateWorker(resources, results)

	result := <-results
	require.NoError(t, result.err)
	assert.NotNil(t, result.resource)
}

func TestUpdateWorkerReturnsErrorForNullState(t *testing.T) {
	nullState := cty.NullVal(cty.DynamicPseudoType)
	resources := make(chan UpdatableResource, 1)
	results := make(chan updateWorkerResult, 1)
	resources <- &updateWorkerTestResource{state: &nullState}
	close(resources)

	updateWorker(resources, results)

	result := <-results
	require.Error(t, result.err)
	assert.Equal(t, "resource doesn't exist anymore", result.err.Error())
}

type updateWorkerTestResource struct {
	state *cty.Value
	err   error
}

func (r *updateWorkerTestResource) Type() string {
	return "aws_test_resource"
}

func (r *updateWorkerTestResource) ID() string {
	return "test-id"
}

func (r *updateWorkerTestResource) State() *cty.Value {
	return r.state
}

func (r *updateWorkerTestResource) UpdateState() error {
	return r.err
}
