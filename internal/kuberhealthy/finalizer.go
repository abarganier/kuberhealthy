package kuberhealthy

import (
	"context"
	"fmt"

	khapi "github.com/kuberhealthy/kuberhealthy/v3/pkg/api"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

const khCheckFinalizer = "kuberhealthy.io/healthcheck"

// hasFinalizer returns true when the check contains the kuberhealthy healthcheck finalizer.
func (k *Kuberhealthy) hasFinalizer(check *khapi.HealthCheck) bool {
	// iterate over the finalizers and return when the kuberhealthy finalizer is present
	for _, f := range check.Finalizers {
		if f == khCheckFinalizer {
			return true
		}
	}
	return false
}

// addFinalizer appends the kuberhealthy finalizer if it is missing.
func (k *Kuberhealthy) addFinalizer(ctx context.Context, check *khapi.HealthCheck) error {
	if k.hasFinalizer(check) {
		return nil
	}
	// Use a merge patch so only metadata.finalizers is sent to the API server.
	// A full-object Update would reserialize the typed struct and rewrite
	// fields the operator does not intend to touch (e.g. *metav1.Duration
	// values like "5m" become "5m0s", and bool false with omitempty is dropped).
	base := client.MergeFrom(check.DeepCopy())
	check.Finalizers = append(check.Finalizers, khCheckFinalizer)
	if err := k.CheckClient.Patch(ctx, check, base); err != nil {
		return fmt.Errorf("failed to add finalizer: %w", err)
	}
	return nil
}

// deleteFinalizer removes the kuberhealthy finalizer when present.
func (k *Kuberhealthy) deleteFinalizer(ctx context.Context, check *khapi.HealthCheck) error {
	if !k.hasFinalizer(check) {
		return nil
	}
	// build a new finalizer slice that omits the kuberhealthy finalizer
	var finalizers []string
	for _, f := range check.Finalizers {
		if f != khCheckFinalizer {
			finalizers = append(finalizers, f)
		}
	}
	// Use a merge patch so only metadata.finalizers is sent to the API server.
	// See addFinalizer for why a full-object Update is unsafe here.
	base := client.MergeFrom(check.DeepCopy())
	check.Finalizers = finalizers
	if err := k.CheckClient.Patch(ctx, check, base); err != nil {
		return fmt.Errorf("failed to remove finalizer: %w", err)
	}
	return nil
}
