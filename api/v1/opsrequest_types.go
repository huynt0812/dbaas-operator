package v1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// OpsRequestSpec defines the desired state of OpsRequest
type OpsRequestSpec struct {
	// ClusterRef references the target DatabaseCluster
	// +kubebuilder:validation:Required
	ClusterRef corev1.LocalObjectReference `json:"clusterRef"`

	// Type is the operation type
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Start;Stop;Restart;Switchover;VerticalScaling;HorizontalScaling;VolumeExpansion;Reconfiguring;Upgrade;Backup;Restore;Expose;RebuildInstance;Custom
	Type OpsRequestType `json:"type"`

	// HorizontalScaling contains horizontal scaling parameters
	// +optional
	HorizontalScaling *HorizontalScalingSpec `json:"horizontalScaling,omitempty"`

	// VerticalScaling contains vertical scaling parameters
	// +optional
	VerticalScaling *VerticalScalingSpec `json:"verticalScaling,omitempty"`

	// VolumeExpansion contains volume expansion parameters
	// +optional
	VolumeExpansion *VolumeExpansionSpec `json:"volumeExpansion,omitempty"`

	// Reconfiguring contains reconfiguration parameters
	// +optional
	Reconfiguring *ReconfiguringSpec `json:"reconfiguring,omitempty"`

	// Upgrade contains upgrade parameters
	// +optional
	Upgrade *UpgradeSpec `json:"upgrade,omitempty"`

	// Backup contains backup parameters
	// +optional
	Backup *BackupRequestSpec `json:"backup,omitempty"`

	// Restore contains restore parameters
	// +optional
	Restore *RestoreRequestSpec `json:"restore,omitempty"`

	// Switchover contains switchover parameters
	// +optional
	Switchover *SwitchoverSpec `json:"switchover,omitempty"`

	// RebuildInstance contains rebuild instance parameters
	// +optional
	RebuildInstance *RebuildInstanceSpec `json:"rebuildInstance,omitempty"`

	// Custom contains custom operation parameters
	// +optional
	Custom *CustomOperationSpec `json:"custom,omitempty"`

	// TTLSecondsAfterFinished is the time to live after the operation finished
	// +optional
	TTLSecondsAfterFinished *int32 `json:"ttlSecondsAfterFinished,omitempty"`
}

// OpsRequestType represents the type of operation
type OpsRequestType string

const (
	OpsRequestTypeStart           OpsRequestType = "Start"
	OpsRequestTypeStop            OpsRequestType = "Stop"
	OpsRequestTypeRestart         OpsRequestType = "Restart"
	OpsRequestTypeSwitchover      OpsRequestType = "Switchover"
	OpsRequestTypeVerticalScaling OpsRequestType = "VerticalScaling"
	OpsRequestTypeHorizontalScaling OpsRequestType = "HorizontalScaling"
	OpsRequestTypeVolumeExpansion OpsRequestType = "VolumeExpansion"
	OpsRequestTypeReconfiguring   OpsRequestType = "Reconfiguring"
	OpsRequestTypeUpgrade         OpsRequestType = "Upgrade"
	OpsRequestTypeBackup          OpsRequestType = "Backup"
	OpsRequestTypeRestore         OpsRequestType = "Restore"
	OpsRequestTypeExpose          OpsRequestType = "Expose"
	OpsRequestTypeRebuildInstance OpsRequestType = "RebuildInstance"
	OpsRequestTypeCustom          OpsRequestType = "Custom"
)

// HorizontalScalingSpec defines horizontal scaling parameters
type HorizontalScalingSpec struct {
	// Replicas is the desired number of replicas
	// +kubebuilder:validation:Minimum=1
	Replicas int32 `json:"replicas"`
}

// VerticalScalingSpec defines vertical scaling parameters
type VerticalScalingSpec struct {
	// Resources specifies the new resource requirements
	Resources corev1.ResourceRequirements `json:"resources"`
}

// VolumeExpansionSpec defines volume expansion parameters
type VolumeExpansionSpec struct {
	// Size is the new storage size
	Size resource.Quantity `json:"size"`
}

// ReconfiguringSpec defines reconfiguration parameters
type ReconfiguringSpec struct {
	// Config contains the new configuration parameters
	Config []ConfigParameter `json:"config"`
}

// UpgradeSpec defines upgrade parameters
type UpgradeSpec struct {
	// TargetVersion is the target database version
	TargetVersion string `json:"targetVersion"`
}

// BackupRequestSpec defines backup parameters
type BackupRequestSpec struct {
	// BackupName is the name for the backup
	// +optional
	BackupName string `json:"backupName,omitempty"`
}

// RestoreRequestSpec defines restore parameters
type RestoreRequestSpec struct {
	// BackupName is the name of the backup to restore from
	BackupName string `json:"backupName"`

	// PointInTime is the timestamp to restore to (for PITR)
	// +optional
	PointInTime *metav1.Time `json:"pointInTime,omitempty"`
}

// SwitchoverSpec defines switchover parameters
type SwitchoverSpec struct {
	// TargetInstance is the instance to promote as primary
	// +optional
	TargetInstance string `json:"targetInstance,omitempty"`
}

// RebuildInstanceSpec defines rebuild instance parameters
type RebuildInstanceSpec struct {
	// InstanceName is the name of the instance to rebuild
	InstanceName string `json:"instanceName"`

	// SourceInstance is the instance to rebuild from
	// +optional
	SourceInstance string `json:"sourceInstance,omitempty"`
}

// CustomOperationSpec defines custom operation parameters
type CustomOperationSpec struct {
	// Operation is the custom operation name
	Operation string `json:"operation"`

	// Parameters contains operation-specific parameters
	// +optional
	Parameters map[string]string `json:"parameters,omitempty"`
}

// OpsRequestStatus defines the observed state of OpsRequest
type OpsRequestStatus struct {
	// Conditions represent the latest available observations of the operation's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Phase is the current phase of the operation
	// +optional
	Phase OpsRequestPhase `json:"phase,omitempty"`

	// StartTime is when the operation started
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime is when the operation completed
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// Message provides additional information about the current state
	// +optional
	Message string `json:"message,omitempty"`

	// ActionLog contains logs from the operation execution
	// +optional
	ActionLog []ActionLogEntry `json:"actionLog,omitempty"`
}

// OpsRequestPhase represents the current phase of the operation
// +kubebuilder:validation:Enum=Pending;Running;Succeeded;Failed
type OpsRequestPhase string

const (
	OpsRequestPhasePending   OpsRequestPhase = "Pending"
	OpsRequestPhaseRunning   OpsRequestPhase = "Running"
	OpsRequestPhaseSucceeded OpsRequestPhase = "Succeeded"
	OpsRequestPhaseFailed    OpsRequestPhase = "Failed"
)

// ActionLogEntry represents a log entry for an action
type ActionLogEntry struct {
	// Timestamp is when the action occurred
	Timestamp metav1.Time `json:"timestamp"`

	// Action is the action that was performed
	Action string `json:"action"`

	// Status is the status of the action (Success, Failed, InProgress)
	Status string `json:"status"`

	// Message provides additional details about the action
	// +optional
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=ops
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`
// +kubebuilder:printcolumn:name="Cluster",type=string,JSONPath=`.spec.clusterRef.name`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// OpsRequest is the Schema for the opsrequests API
type OpsRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OpsRequestSpec   `json:"spec,omitempty"`
	Status OpsRequestStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// OpsRequestList contains a list of OpsRequest
type OpsRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OpsRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OpsRequest{}, &OpsRequestList{})
}
