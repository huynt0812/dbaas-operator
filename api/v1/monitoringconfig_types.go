package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MonitoringConfigSpec defines the desired state of MonitoringConfig
type MonitoringConfigSpec struct {
	// Type is the monitoring type (pmm, prometheus, datadog, etc.)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=pmm;prometheus;datadog;newrelic
	Type string `json:"type"`

	// PMM contains PMM-specific configuration
	// +optional
	PMM *PMMConfigSpec `json:"pmm,omitempty"`

	// Prometheus contains Prometheus-specific configuration
	// +optional
	Prometheus *PrometheusConfigSpec `json:"prometheus,omitempty"`

	// CredentialsSecretRef references a secret containing monitoring credentials
	// +optional
	CredentialsSecretRef *corev1.LocalObjectReference `json:"credentialsSecretRef,omitempty"`
}

// PMMConfigSpec defines PMM configuration
type PMMConfigSpec struct {
	// ServerHost is the PMM server host
	// +kubebuilder:validation:Required
	ServerHost string `json:"serverHost"`

	// ServerPort is the PMM server port
	// +kubebuilder:default=443
	ServerPort int32 `json:"serverPort,omitempty"`

	// ServerUser is the PMM server username
	// +optional
	ServerUser string `json:"serverUser,omitempty"`

	// Image is the PMM client image
	// +optional
	Image string `json:"image,omitempty"`

	// Resources specifies the compute resources for PMM client
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

// PrometheusConfigSpec defines Prometheus configuration
type PrometheusConfigSpec struct {
	// ServiceMonitorEnabled enables ServiceMonitor creation
	// +kubebuilder:default=true
	ServiceMonitorEnabled bool `json:"serviceMonitorEnabled,omitempty"`

	// Interval is the scrape interval
	// +kubebuilder:default="30s"
	Interval string `json:"interval,omitempty"`

	// ScrapeTimeout is the scrape timeout
	// +optional
	ScrapeTimeout string `json:"scrapeTimeout,omitempty"`

	// AdditionalLabels are additional labels for ServiceMonitor
	// +optional
	AdditionalLabels map[string]string `json:"additionalLabels,omitempty"`
}

// MonitoringConfigStatus defines the observed state of MonitoringConfig
type MonitoringConfigStatus struct {
	// Conditions represent the latest available observations of the config's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Phase is the current phase of the config (Ready, NotReady)
	// +optional
	Phase string `json:"phase,omitempty"`

	// Message provides additional information about the current state
	// +optional
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=mc
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// MonitoringConfig is the Schema for the monitoringconfigs API
type MonitoringConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MonitoringConfigSpec   `json:"spec,omitempty"`
	Status MonitoringConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MonitoringConfigList contains a list of MonitoringConfig
type MonitoringConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MonitoringConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MonitoringConfig{}, &MonitoringConfigList{})
}
