package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DatabaseEngineSpec defines the desired state of DatabaseEngine
type DatabaseEngineSpec struct {
	// Type is the database engine type (postgresql, mongodb, mysql, kafka, etc.)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=postgresql;mongodb;mysql;kafka
	Type string `json:"type"`

	// Provider is the operator provider (cnpg, percona, strimzi, etc.)
	// +kubebuilder:validation:Required
	Provider string `json:"provider"`

	// SupportedVersions lists the database versions supported by this engine
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	SupportedVersions []string `json:"supportedVersions"`

	// OperatorVersion is the version of the underlying operator
	// +kubebuilder:validation:Required
	OperatorVersion string `json:"operatorVersion"`

	// DefaultConfig contains default configuration parameters
	// +optional
	DefaultConfig []ConfigParameter `json:"defaultConfig,omitempty"`

	// Features lists the features supported by this engine
	// +optional
	Features EngineFeatures `json:"features,omitempty"`
}

// EngineFeatures defines the features supported by a database engine
type EngineFeatures struct {
	// Backup indicates if backup is supported
	// +kubebuilder:default=true
	Backup bool `json:"backup,omitempty"`

	// PITR indicates if Point-in-Time Recovery is supported
	// +kubebuilder:default=false
	PITR bool `json:"pitr,omitempty"`

	// Monitoring indicates if monitoring is supported
	// +kubebuilder:default=true
	Monitoring bool `json:"monitoring,omitempty"`

	// Proxy indicates if proxy is supported
	// +kubebuilder:default=false
	Proxy bool `json:"proxy,omitempty"`

	// HorizontalScaling indicates if horizontal scaling is supported
	// +kubebuilder:default=true
	HorizontalScaling bool `json:"horizontalScaling,omitempty"`

	// VerticalScaling indicates if vertical scaling is supported
	// +kubebuilder:default=true
	VerticalScaling bool `json:"verticalScaling,omitempty"`

	// VolumeExpansion indicates if volume expansion is supported
	// +kubebuilder:default=true
	VolumeExpansion bool `json:"volumeExpansion,omitempty"`
}

// DatabaseEngineStatus defines the observed state of DatabaseEngine
type DatabaseEngineStatus struct {
	// Conditions represent the latest available observations of the engine's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Phase is the current phase of the engine (Ready, NotReady)
	// +optional
	Phase string `json:"phase,omitempty"`

	// Message provides additional information about the current state
	// +optional
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=dbe
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`
// +kubebuilder:printcolumn:name="Provider",type=string,JSONPath=`.spec.provider`
// +kubebuilder:printcolumn:name="Operator Version",type=string,JSONPath=`.spec.operatorVersion`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// DatabaseEngine is the Schema for the databaseengines API
type DatabaseEngine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatabaseEngineSpec   `json:"spec,omitempty"`
	Status DatabaseEngineStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DatabaseEngineList contains a list of DatabaseEngine
type DatabaseEngineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DatabaseEngine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DatabaseEngine{}, &DatabaseEngineList{})
}
