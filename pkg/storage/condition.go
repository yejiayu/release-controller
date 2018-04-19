package storage

import (
	releaseapi "github.com/caicloud/clientset/pkg/apis/release/v1alpha1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Reasons for releases
const (
	ReasonAvailable   = "Available"
	ReasonFailure     = "Failure"
	ReasonCreating    = "Creating"
	ReasonUpdating    = "Updating"
	ReasonRollbacking = "Rollbacking"
)

// ConditionAvailable returns an available condition.
func ConditionAvailable() releaseapi.ReleaseCondition {
	return releaseapi.ReleaseCondition{
		Type:               releaseapi.ReleaseAvailable,
		Status:             core.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             ReasonAvailable,
		Message:            "",
	}
}

// ConditionFailure returns a failure condition.
func ConditionFailure(message string) releaseapi.ReleaseCondition {
	return releaseapi.ReleaseCondition{
		Type:               releaseapi.ReleaseFailure,
		Status:             core.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             ReasonFailure,
		Message:            message,
	}
}

// ConditionCreating returns a creating condition.
func ConditionCreating() releaseapi.ReleaseCondition {
	return releaseapi.ReleaseCondition{
		Type:               releaseapi.ReleaseProgressing,
		Status:             core.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             ReasonCreating,
		Message:            "",
	}
}

// ConditionUpdating returns a updating condition.
func ConditionUpdating() releaseapi.ReleaseCondition {
	return releaseapi.ReleaseCondition{
		Type:               releaseapi.ReleaseProgressing,
		Status:             core.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             ReasonUpdating,
		Message:            "",
	}
}

// ConditionRollbacking returns a rollbacking condition.
func ConditionRollbacking() releaseapi.ReleaseCondition {
	return releaseapi.ReleaseCondition{
		Type:               releaseapi.ReleaseProgressing,
		Status:             core.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             ReasonRollbacking,
		Message:            "",
	}
}
