package resource

import (
	"errors"
	"fmt"
	"sort"

	"github.com/apex/log"
	"github.com/chenrui333/terradozer/internal"
)

const (
	fieldType              = "type"
	defaultDestroyPriority = 0
)

// DestroyableResource implementations can destroy a Terraform resource.
type DestroyableResource interface {
	Destroy() error
	Type() string
	ID() string
}

// DestroyResources destroys a given list of resources, which may depend on each other.
//
// If at least one resource is successfully destroyed per run (iteration through the list of given resources),
// the remaining, failed resources will be retried in a next run (until all resources are destroyed or
// some destroys have permanently failed).
func DestroyResources(resources []DestroyableResource, parallel int) int {
	if parallel < 1 {
		log.WithField("parallel", parallel).Warn(internal.Pad("invalid parallelism; using one worker"))
		parallel = 1
	}

	numOfDeletedResources := 0

	var retryableResourceErrors []RetryDestroyError

	resourceBuckets := groupDestroyableResourcesByPriority(resources)

	for _, bucket := range resourceBuckets {
		deletedResources, retryableErrors := destroyResourceBucket(bucket.resources, parallel)
		numOfDeletedResources += deletedResources
		retryableResourceErrors = append(retryableResourceErrors, retryableErrors...)
	}

	if len(retryableResourceErrors) > 0 && numOfDeletedResources > 0 {
		var resourcesToRetry []DestroyableResource
		for _, retryErr := range retryableResourceErrors {
			resourcesToRetry = append(resourcesToRetry, retryErr.Resource)
		}

		numOfDeletedResources += DestroyResources(resourcesToRetry, parallel)
	}

	if len(retryableResourceErrors) > 0 && numOfDeletedResources == 0 {
		internal.LogTitle(fmt.Sprintf("failed to delete the following resources (retries exceeded): %d",
			len(retryableResourceErrors)))

		for _, err := range retryableResourceErrors {
			log.WithError(err).WithField("id", err.Resource.ID()).Warn(internal.Pad(err.Resource.Type()))
		}
	}

	return numOfDeletedResources
}

func destroyResourceBucket(resources []DestroyableResource, parallel int) (int, []RetryDestroyError) {
	numOfResourcesToDelete := len(resources)
	numOfDeletedResources := 0
	retryableResourceErrors := []RetryDestroyError{}

	jobQueue := make(chan DestroyableResource, numOfResourcesToDelete)

	workerResults := make(chan workerResult, numOfResourcesToDelete)

	for i := 1; i <= parallel; i++ {
		go workerDestroy(jobQueue, workerResults)
	}

	log.Debug("start distributing resources to workers for this run")

	for _, r := range resources {
		jobQueue <- r
	}

	close(jobQueue)

	for i := 1; i <= numOfResourcesToDelete; i++ {
		result := <-workerResults

		if result.resourceHasBeenDeleted {
			numOfDeletedResources++

			continue
		}

		if result.Err != nil {
			retryableResourceErrors = append(retryableResourceErrors, *result.Err)
		}
	}

	return numOfDeletedResources, retryableResourceErrors
}

type destroyPriorityBucket struct {
	priority  int
	resources []DestroyableResource
}

func groupDestroyableResourcesByPriority(resources []DestroyableResource) []destroyPriorityBucket {
	priorityByType := destroyPriorityByType()
	resourcesToDelete := append([]DestroyableResource(nil), resources...)

	sort.SliceStable(resourcesToDelete, func(i, j int) bool {
		return destroyPriority(resourcesToDelete[i].Type(), priorityByType) <
			destroyPriority(resourcesToDelete[j].Type(), priorityByType)
	})

	buckets := []destroyPriorityBucket{}
	for _, r := range resourcesToDelete {
		priority := destroyPriority(r.Type(), priorityByType)
		lastBucketIndex := len(buckets) - 1
		if len(buckets) == 0 || buckets[lastBucketIndex].priority != priority {
			buckets = append(buckets, destroyPriorityBucket{priority: priority})
			lastBucketIndex = len(buckets) - 1
		}

		buckets[lastBucketIndex].resources = append(buckets[lastBucketIndex].resources, r)
	}

	return buckets
}

func destroyPriority(resourceType string, priorityByType map[string]int) int {
	priority, ok := priorityByType[resourceType]
	if !ok {
		// Unknown resource types intentionally run first to avoid delaying known dependency chains.
		return defaultDestroyPriority
	}

	return priority
}

func destroyPriorityByType() map[string]int {
	return map[string]int{
		// EKS chain.
		"aws_eks_addon":                     10,
		"aws_eks_node_group":                20,
		"aws_eks_access_entry":              25,
		"aws_eks_access_policy_association": 25,
		"aws_eks_cluster":                   30,

		// Load balancer chain.
		"aws_lb_listener":      10,
		"aws_lb_listener_rule": 10,
		"aws_lb_target_group":  15,
		"aws_lb":               20,
		"aws_alb_listener":     10,
		"aws_alb_target_group": 15,
		"aws_alb":              20,

		// Network chain.
		"aws_route_table_association": 10,
		"aws_route":                   15,
		"aws_route_table":             20,
		"aws_nat_gateway":             25,
		"aws_eip":                     30,
		"aws_security_group_rule":     35,
		"aws_network_interface":       38,
		"aws_subnet":                  40,
		"aws_security_group":          45,
		"aws_internet_gateway":        48,
		"aws_vpc":                     50,

		// IAM chain.
		"aws_iam_role_policy_attachment": 10,
		"aws_iam_role_policy":            10,
		"aws_iam_policy":                 20,
		"aws_iam_instance_profile":       25,
		"aws_iam_role":                   30,

		// S3 chain.
		"aws_s3_bucket_policy":     10,
		"aws_s3_bucket_versioning": 10,
		"aws_s3_bucket":            20,

		// API Gateway chain.
		"aws_api_gateway_method_settings": 10,
		"aws_api_gateway_integration":     15,
		"aws_api_gateway_method":          15,
		"aws_api_gateway_stage":           20,
		"aws_api_gateway_deployment":      25,
		"aws_api_gateway_resource":        30,
		"aws_api_gateway_rest_api":        40,
	}
}

type workerResult struct {
	resourceHasBeenDeleted bool
	// if set, it is worth retrying to delete this resource
	Err *RetryDestroyError
}

// workerDestroy is a worker that destroys a resource.
func workerDestroy(resources <-chan DestroyableResource, result chan<- workerResult) {
	for r := range resources {
		err := r.Destroy()
		if err != nil {
			switch err := err.(type) {
			case *RetryDestroyError:
				log.WithFields(log.Fields{
					fieldType:     r.Type(),
					"resource_id": r.ID(),
				}).Info(internal.Pad("will retry to delete resource"))

				result <- workerResult{
					Err: err,
				}

			default:
				log.WithError(err).WithFields(log.Fields{
					fieldType:     r.Type(),
					"resource_id": r.ID(),
				}).Debug(internal.Pad("unable to delete resource"))

				result <- workerResult{}
			}

			continue
		}

		result <- workerResult{
			resourceHasBeenDeleted: true,
		}
	}
}

// Destroy destroys a Terraform resource.
func (r Resource) Destroy() error {
	if r.State() == nil {
		return errors.New("resource state is nil; need to call update first")
	}

	err := r.Provider.DestroyResource(r.Type(), *r.State())
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"id": r.ID(), fieldType: r.Type()}).Debug(internal.Pad("failed to delete resource"))

		return NewRetryDestroyError(err, &r)
	}

	log.WithField("id", r.ID()).Error(internal.Pad(r.Type()))

	return nil
}
