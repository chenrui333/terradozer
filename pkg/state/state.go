// Package state provides primitives to list all resources and providers in a Terraform state file.
package state

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/chenrui333/terradozer/internal"
	"github.com/chenrui333/terradozer/internal/awstools/terraform"
	"github.com/chenrui333/terradozer/internal/awstools/terraform/provider"
	"github.com/chenrui333/terradozer/pkg/resource"
	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/states"
	"github.com/hashicorp/terraform/states/statefile"
	"github.com/zclconf/go-cty/cty"
)

var modernProviderReferencePattern = regexp.MustCompile(
	`^((?:module\.[^.]+\.)*)provider\["registry\.terraform\.io/[^/]+/([^"\]]+)"\](?:\.(.+))?$`,
)

// State represents a Terraform state.
type State struct {
	state *states.State
}

type s3ObjectReader func(ctx context.Context, bucket, key string) ([]byte, error)

const defaultS3StateReadTimeout = 30 * time.Second

// New creates a state from a given local path or S3 URI to a Terraform state file.
func New(source string) (*State, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultS3StateReadTimeout)
	defer cancel()

	return NewWithContext(ctx, source)
}

// NewWithContext creates a state using the provided context for remote state reads.
func NewWithContext(ctx context.Context, source string) (*State, error) {
	return newWithS3ObjectReader(ctx, source, readS3ObjectFromAWS)
}

func newWithS3ObjectReader(ctx context.Context, source string, reader s3ObjectReader) (*State, error) {
	stateFile, err := getStateFromSource(ctx, source, reader)
	if err != nil {
		return nil, err
	}

	return &State{stateFile.State}, nil
}

// copied from github.com/hashicorp/terraform/command/show.go
func getStateFromSource(ctx context.Context, source string, reader s3ObjectReader) (*statefile.File, error) {
	stateData, err := readStateData(ctx, source, reader)
	if err != nil {
		return nil, err
	}

	stateFile, err := readStateFile(stateData)
	if err != nil {
		return nil, fmt.Errorf("failed reading %s as a statefile: %s", source, err)
	}

	if len(stateFile.State.ProviderAddrs()) > 0 || !stateJSONHasResources(stateData) {
		return stateFile, nil
	}

	normalizedStateData, changed := normalizeProviderReferences(stateData)
	if !changed {
		return stateFile, nil
	}

	normalizedStateFile, normalizeErr := readStateFile(normalizedStateData)
	if normalizeErr == nil {
		log.WithField("source", source).Debug(internal.Pad("normalized Terraform 1.x provider references"))

		return normalizedStateFile, nil
	}

	return stateFile, nil
}

type s3StateSource struct {
	bucket string
	key    string
}

func readStateData(ctx context.Context, source string, reader s3ObjectReader) ([]byte, error) {
	s3Source, ok, err := parseS3StateSource(source)
	if err != nil {
		return nil, err
	}

	if ok {
		return reader(ctx, s3Source.bucket, s3Source.key)
	}

	return os.ReadFile(source)
}

func parseS3StateSource(source string) (s3StateSource, bool, error) {
	if !strings.HasPrefix(strings.ToLower(source), "s3://") {
		return s3StateSource{}, false, nil
	}

	if s3SourceHasEmbeddedCredentials(source) {
		return s3StateSource{}, true, errors.New("S3 state path must not include embedded credentials")
	}

	parsed, err := url.Parse(source)
	if err != nil {
		return s3StateSource{}, true, fmt.Errorf("failed to parse S3 state path: %w", err)
	}

	if parsed.User != nil {
		return s3StateSource{}, true, errors.New("S3 state path must not include embedded credentials")
	}

	if parsed.Host == "" {
		return s3StateSource{}, true, fmt.Errorf("S3 state path %q must include a bucket", source)
	}

	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return s3StateSource{}, true, fmt.Errorf("S3 state path %q must not include query or fragment components", source)
	}

	key := strings.TrimPrefix(parsed.Path, "/")
	if key == "" {
		return s3StateSource{}, true, fmt.Errorf("S3 state path %q must include a key", source)
	}

	return s3StateSource{bucket: parsed.Host, key: key}, true, nil
}

func s3SourceHasEmbeddedCredentials(source string) bool {
	remainder := source[len("s3://"):]
	authorityEnd := strings.IndexAny(remainder, "/?#")
	if authorityEnd == -1 {
		authorityEnd = len(remainder)
	}

	return strings.Contains(remainder[:authorityEnd], "@")
}

func readS3ObjectFromAWS(ctx context.Context, bucket, key string) ([]byte, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)
	object, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get s3://%s/%s: %w", bucket, key, err)
	}

	if object.Body == nil {
		return nil, fmt.Errorf("failed to get s3://%s/%s: response body is nil", bucket, key)
	}

	stateData, err := io.ReadAll(object.Body)
	closeErr := object.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to read s3://%s/%s: %w", bucket, key, err)
	}

	if closeErr != nil {
		return nil, fmt.Errorf("failed to close s3://%s/%s response body: %w", bucket, key, closeErr)
	}

	return stateData, nil
}

func readStateFile(stateData []byte) (*statefile.File, error) {
	stateFile, err := statefile.Read(bytes.NewReader(stateData))
	if err != nil {
		return nil, err
	}

	return stateFile, nil
}

func stateJSONHasResources(stateData []byte) bool {
	var rawState struct {
		Resources []json.RawMessage `json:"resources"`
	}

	err := json.Unmarshal(stateData, &rawState)
	if err != nil {
		return false
	}

	return len(rawState.Resources) > 0
}

func normalizeProviderReferences(stateData []byte) ([]byte, bool) {
	var rawState map[string]any

	err := json.Unmarshal(stateData, &rawState)
	if err != nil {
		return stateData, false
	}

	rawResources, ok := rawState["resources"].([]any)
	if !ok {
		return stateData, false
	}

	changed := false

	for _, rawResource := range rawResources {
		resourceMap, ok := rawResource.(map[string]any)
		if !ok {
			continue
		}

		providerRef, ok := resourceMap["provider"].(string)
		if !ok {
			continue
		}

		normalizedProviderRef, providerRefChanged := normalizeProviderReference(providerRef)
		if providerRefChanged {
			resourceMap["provider"] = normalizedProviderRef
			changed = true
		}
	}

	if !changed {
		return stateData, false
	}

	normalizedStateData, err := json.Marshal(rawState)
	if err != nil {
		return stateData, false
	}

	return normalizedStateData, true
}

func normalizeProviderReference(providerRef string) (string, bool) {
	matches := modernProviderReferencePattern.FindStringSubmatch(providerRef)
	if len(matches) != 4 {
		return providerRef, false
	}

	modulePrefix := matches[1]
	providerType := matches[2]
	providerAlias := matches[3]

	normalizedProviderRef := modulePrefix + "provider." + providerType
	if providerAlias != "" {
		normalizedProviderRef = normalizedProviderRef + "." + providerAlias
	}

	return normalizedProviderRef, true
}

// ProviderNames returns a list of all provider names (e.g., "aws", "google") in the state.
// The result of provider names is deduplicated.
func (s *State) ProviderNames() []string {
	var providers []string

	log.WithField("addresses", s.state.ProviderAddrs()).Debug(internal.Pad("providers found in state"))

	for _, pAddr := range s.state.ProviderAddrs() {
		providers = append(providers, pAddr.ProviderConfig.StringCompact())
	}

	return removeDuplicates(providers)
}

func removeDuplicates(elements []string) []string {
	encountered := map[string]bool{}

	var result []string

	for i := range elements {
		if encountered[elements[i]] {
			// do not add duplicate
		} else {
			// record this element as an encountered element
			encountered[elements[i]] = true
			result = append(result, elements[i])
		}
	}

	return result
}

// Resources returns a list of resources in the state that are managed by one of the given providers.
//
// Data sources are not returned as these are managed outside the scope of the state and
// therefore shouldn't be destroyed.
func (s *State) Resources(providers map[string]*provider.TerraformProvider) ([]terraform.UpdatableResource, error) {
	var resources []terraform.UpdatableResource

	for _, resAddr := range lookupAllResourceInstanceAddrs(s.state) {
		log.WithField("absolute_address", resAddr.String()).
			Debug(internal.Pad("looked up resource instance address"))

		resInstance := s.state.ResourceInstance(resAddr)

		resID, err := getResourceID(resInstance)
		if err != nil {
			return nil, fmt.Errorf("failed to get id for resource (addr=%s): %s", resAddr.String(), err)
		}

		if resAddr.ContainingResource().Resource.Mode != addrs.ManagedResourceMode {
			log.WithFields(log.Fields{
				"mode": resAddr.Resource.Resource.Mode,
				"type": resAddr.Resource.Resource.Type,
				"id":   resID}).Debug(internal.Pad("ignoring non-managed resource"))

			continue
		}

		providerName := resAddr.Resource.Resource.DefaultProviderConfig().StringCompact()

		p, ok := providers[providerName]
		if !ok {
			log.WithField("name", providerName).Debug(internal.Pad("Terraform provider not found in providers list"))

			continue
		}

		resObject, err := getResourceState(resInstance, resAddr.Resource.Resource.Type, p)
		if err != nil {
			return nil, fmt.Errorf("failed to decode resource into object (addr=%s): %s", resAddr.String(), err)
		}

		r := resource.NewWithState(resAddr.Resource.Resource.Type, resID, p, &resObject)
		resources = append(resources, r)
	}

	return resources, nil
}

// resourceID represents the ID attribute of a Terraform resource.
type resourceID struct {
	ID string `json:"id"`
}

// getResourceID looks up the resource ID amongst all resource attributes.
func getResourceID(resInstance *states.ResourceInstance) (string, error) {
	var result resourceID

	if !resInstance.HasCurrent() {
		return "", errors.New("resource instance has no current object")
	}

	if resInstance.Current.AttrsJSON != nil {
		err := json.Unmarshal(resInstance.Current.AttrsJSON, &result)
		if err != nil {
			log.WithField("attributes", resInstance.Current.AttrsJSON).
				Debug(internal.Pad("JSON-encoded attributes of resource instance"))

			return "", fmt.Errorf("failed to unmarshal JSON-encoded resource instance attributes: %s", err)
		}

		return result.ID, nil
	}

	if resInstance.Current.AttrsFlat == nil {
		log.WithField("attributes", resInstance.Current.AttrsFlat).
			Debug(internal.Pad("legacy attributes of resource instance"))

		return "", errors.New("flat attribute map of resource instance is nil")
	}

	return resInstance.Current.AttrsFlat["id"], nil
}

// getResourceState unmarshals the JSON representation of a resource found in the state file into
// an internal Terraform state object representation.
func getResourceState(resInstance *states.ResourceInstance, rType string,
	provider *provider.TerraformProvider) (cty.Value, error) {
	if !resInstance.HasCurrent() {
		return cty.NilVal, errors.New("resource instance has no current object")
	}

	resourceSchema, err := provider.GetSchemaForResource(rType)
	if err != nil {
		return cty.NilVal, err
	}

	resInstanceObj, filteredUnknownAttrs, err := decodeResourceInstanceObject(
		resInstance, resourceSchema.Block.ImpliedType())
	if err != nil {
		return cty.NilVal, err
	}

	if filteredUnknownAttrs {
		log.WithField("type", rType).Debug(internal.Pad("ignored unknown attributes in resource state"))
	}

	return resInstanceObj.Value, nil
}

func decodeResourceInstanceObject(
	resInstance *states.ResourceInstance, schemaType cty.Type) (*states.ResourceInstanceObject, bool, error) {
	resInstanceObj, err := resInstance.Current.Decode(schemaType)
	if err == nil {
		return resInstanceObj, false, nil
	}

	if resInstance.Current.AttrsJSON == nil {
		return nil, false, err
	}

	attrsJSONPruned, changed, pruneErr := pruneUnknownAttributesFromJSON(resInstance.Current.AttrsJSON, schemaType)
	if pruneErr != nil || !changed {
		return nil, false, err
	}

	current := *resInstance.Current
	current.AttrsJSON = attrsJSONPruned

	resInstanceObj, decodeErr := current.Decode(schemaType)
	if decodeErr != nil {
		return nil, false, err
	}

	return resInstanceObj, true, nil
}

func pruneUnknownAttributesFromJSON(attrsJSON []byte, schemaType cty.Type) ([]byte, bool, error) {
	var raw any

	err := json.Unmarshal(attrsJSON, &raw)
	if err != nil {
		return nil, false, err
	}

	pruned, changed := pruneUnknownAttributes(raw, schemaType)
	if !changed {
		return attrsJSON, false, nil
	}

	prunedJSON, err := json.Marshal(pruned)
	if err != nil {
		return nil, false, err
	}

	return prunedJSON, true, nil
}

func pruneUnknownAttributes(raw any, schemaType cty.Type) (any, bool) {
	if raw == nil || schemaType == cty.DynamicPseudoType {
		return raw, false
	}

	if schemaType.IsObjectType() {
		return pruneUnknownObjectAttributes(raw, schemaType)
	}

	if schemaType.IsMapType() {
		return pruneUnknownMapAttributes(raw, schemaType.ElementType())
	}

	if schemaType.IsListType() || schemaType.IsSetType() {
		return pruneUnknownListAttributes(raw, schemaType.ElementType())
	}

	if schemaType.IsTupleType() {
		return pruneUnknownTupleAttributes(raw, schemaType.TupleElementTypes())
	}

	return raw, false
}

func pruneUnknownObjectAttributes(raw any, schemaType cty.Type) (any, bool) {
	rawObject, ok := raw.(map[string]any)
	if !ok {
		return raw, false
	}

	attrTypes := schemaType.AttributeTypes()
	prunedObject := make(map[string]any, len(rawObject))
	changed := false

	for key, value := range rawObject {
		attrType, ok := attrTypes[key]
		if !ok {
			changed = true
			continue
		}

		prunedValue, valueChanged := pruneUnknownAttributes(value, attrType)
		if valueChanged {
			changed = true
		}

		prunedObject[key] = prunedValue
	}

	return prunedObject, changed
}

func pruneUnknownMapAttributes(raw any, elemType cty.Type) (any, bool) {
	rawMap, ok := raw.(map[string]any)
	if !ok {
		return raw, false
	}

	prunedMap := make(map[string]any, len(rawMap))
	changed := false

	for key, value := range rawMap {
		prunedValue, valueChanged := pruneUnknownAttributes(value, elemType)
		if valueChanged {
			changed = true
		}

		prunedMap[key] = prunedValue
	}

	return prunedMap, changed
}

func pruneUnknownListAttributes(raw any, elemType cty.Type) (any, bool) {
	rawList, ok := raw.([]any)
	if !ok {
		return raw, false
	}

	prunedList := make([]any, len(rawList))
	changed := false

	for idx, value := range rawList {
		prunedValue, valueChanged := pruneUnknownAttributes(value, elemType)
		if valueChanged {
			changed = true
		}

		prunedList[idx] = prunedValue
	}

	return prunedList, changed
}

func pruneUnknownTupleAttributes(raw any, elemTypes []cty.Type) (any, bool) {
	rawTuple, ok := raw.([]any)
	if !ok {
		return raw, false
	}

	prunedTuple := make([]any, len(rawTuple))
	changed := false

	for idx, value := range rawTuple {
		elemType := cty.DynamicPseudoType
		if idx < len(elemTypes) {
			elemType = elemTypes[idx]
		}

		prunedValue, valueChanged := pruneUnknownAttributes(value, elemType)
		if valueChanged {
			changed = true
		}

		prunedTuple[idx] = prunedValue
	}

	return prunedTuple, changed
}

// copied (and modified) from github.com/hashicorp/terraform/command/state_meta.go
func lookupAllResourceInstanceAddrs(state *states.State) []addrs.AbsResourceInstance {
	var ret []addrs.AbsResourceInstance

	for _, ms := range state.Modules {
		ret = append(ret, collectModuleResourceInstances(ms)...)
	}

	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Less(ret[j])
	})

	return ret
}

// copied from github.com/hashicorp/terraform/command/state_meta.go
func collectModuleResourceInstances(ms *states.Module) []addrs.AbsResourceInstance {
	var ret []addrs.AbsResourceInstance

	for _, rs := range ms.Resources {
		ret = append(ret, collectResourceInstances(ms.Addr, rs)...)
	}

	return ret
}

// copied from github.com/hashicorp/terraform/command/state_meta.go
func collectResourceInstances(moduleAddr addrs.ModuleInstance, rs *states.Resource) []addrs.AbsResourceInstance {
	var ret []addrs.AbsResourceInstance

	for key := range rs.Instances {
		ret = append(ret, rs.Addr.Instance(key).Absolute(moduleAddr))
	}

	return ret
}
