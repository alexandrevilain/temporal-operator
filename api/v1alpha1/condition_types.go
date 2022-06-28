package v1alpha1

import (
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ReconcileErrorCondition indicates a transient or persistent reconciliation error.
	ReconcileErrorCondition string = "ReconcileError"
	// ReconcileSuccessCondition indicates a successful reconciliation.
	ReconcileSuccessCondition string = "ReconcileSuccess"
	// ReadyCondition indicates the cluster is ready to receive traffic.
	ReadyCondition string = "Ready"
)

const (
	// ProgressingReason signals a reconciliation has started.
	ProgressingReason string = "Progressing"
	// ReconcileErrorReason signals a unknown reconciliation error.
	ReconcileErrorReason string = "LastReconcileCycleFailed"
	// ReconcileSuccessReason signals a successful reconciliation.
	ReconcileSuccessReason string = "LastReconcileCycleSucceded"
	// ServicesReadyReason signals all temporal services for the cluster are in ready state.
	ServicesReadyReason string = "ServicesReady"
	// ServicesNotReadyReason signals that not all temporal services for the cluster are in ready state.
	ServicesNotReadyReason string = "ServicesNotReady"
	// PersistenceReconciliationFailedReason signals an error while reconciling persistence.
	PersistenceReconciliationFailedReason string = "PersistenceReconciliationFailed"
	// ResourcesReconciliationFailedReason signals an error while reconciling cluster resources.
	ResourcesReconciliationFailedReason string = "ResoucesReconciliationFailed"
	// TemporalClusterValidationFailedReason signals an error while validation desired cluster version.
	TemporalClusterValidationFailedReason string = "TemporalClusterValidationFailed"
)

// SetTemporalClusterReconcileSuccess sets the ReconcileSuccessCondition status for a temporal cluster.
func SetTemporalClusterReconcileSuccess(c *TemporalCluster, status metav1.ConditionStatus, reason, message string) {
	condition := metav1.Condition{
		Type:               ReconcileSuccessCondition,
		LastTransitionTime: metav1.Now(),
		ObservedGeneration: c.GetGeneration(),
		Reason:             reason,
		Status:             status,
		Message:            message,
	}
	apimeta.SetStatusCondition(&c.Status.Conditions, condition)
}

// SetTemporalClusterReconcileError sets the ReconcileErrorCondition status for a temporal cluster.
func SetTemporalClusterReconcileError(c *TemporalCluster, status metav1.ConditionStatus, reason, message string) {
	condition := metav1.Condition{
		Type:               ReconcileErrorCondition,
		LastTransitionTime: metav1.Now(),
		ObservedGeneration: c.GetGeneration(),
		Reason:             reason,
		Status:             status,
		Message:            message,
	}
	apimeta.SetStatusCondition(&c.Status.Conditions, condition)
}

// SetTemporalClusterReady sets the ReadyCondition status for a temporal cluster.
func SetTemporalClusterReady(c *TemporalCluster, status metav1.ConditionStatus, reason, message string) {
	condition := metav1.Condition{
		Type:               ReadyCondition,
		LastTransitionTime: metav1.Now(),
		ObservedGeneration: c.GetGeneration(),
		Reason:             reason,
		Status:             status,
		Message:            message,
	}
	apimeta.SetStatusCondition(&c.Status.Conditions, condition)
}
