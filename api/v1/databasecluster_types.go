package v1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DatabaseClusterSpec defines the desired state of DatabaseCluster
type DatabaseClusterSpec struct {
	// Engine specifies the database engine type and version
	// +kubebuilder:validation:Required
	Engine EngineSpec `json:"engine"`

	// ClusterSize specifies the number of instances in the cluster
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=3
	ClusterSize int32 `json:"clusterSize,omitempty"`

	// Resources specifies the compute resources for database instances
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Storage specifies the storage configuration
	// +kubebuilder:validation:Required
	Storage StorageSpec `json:"storage"`

	// Backup specifies the backup configuration
	// +optional
	Backup *BackupSpec `json:"backup,omitempty"`

	// Monitoring specifies the monitoring configuration
	// +optional
	Monitoring *MonitoringSpec `json:"monitoring,omitempty"`

	// Proxy specifies the proxy/load balancer configuration
	// +optional
	Proxy *ProxySpec `json:"proxy,omitempty"`

	// Expose specifies how to expose the database service
	// +optional
	Expose *ExposeSpec `json:"expose,omitempty"`

	// PodSchedulingPolicy defines pod scheduling constraints
	// +optional
	PodSchedulingPolicy *PodSchedulingPolicySpec `json:"podSchedulingPolicy,omitempty"`

	// Config contains engine-specific configuration as key-value pairs
	// +optional
	Config []ConfigParameter `json:"config,omitempty"`

	// DataSource specifies the data source for initialization
	// +optional
	DataSource *DataSourceSpec `json:"dataSource,omitempty"`
}

// EngineSpec defines the database engine configuration
type EngineSpec struct {
	// Type is the database engine type (postgresql, mongodb, mysql, kafka, etc.)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=postgresql;mongodb;mysql;kafka
	Type string `json:"type"`

	// Version is the database version
	// +kubebuilder:validation:Required
	Version string `json:"version"`

	// EngineRef references a DatabaseEngine CR for operator-specific settings
	// +optional
	EngineRef *corev1.LocalObjectReference `json:"engineRef,omitempty"`
}

// StorageSpec defines storage configuration
type StorageSpec struct {
	// Size is the storage size
	// +kubebuilder:validation:Required
	Size resource.Quantity `json:"size"`

	// StorageClassName is the storage class name
	// +optional
	StorageClassName *string `json:"storageClassName,omitempty"`

	// VolumeMode specifies the volume mode (Filesystem or Block)
	// +optional
	VolumeMode *corev1.PersistentVolumeMode `json:"volumeMode,omitempty"`
}

// BackupSpec defines backup configuration
type BackupSpec struct {
	// Enabled enables or disables backup
	// +kubebuilder:default=true
	Enabled bool `json:"enabled,omitempty"`

	// Schedule is the cron schedule for automatic backups
	// +optional
	Schedule string `json:"schedule,omitempty"`

	// BackupStorageRef references a BackupStorage CR
	// +optional
	BackupStorageRef *corev1.LocalObjectReference `json:"backupStorageRef,omitempty"`

	// RetentionPolicy specifies how long to keep backups
	// +optional
	RetentionPolicy string `json:"retentionPolicy,omitempty"`
}

// MonitoringSpec defines monitoring configuration
type MonitoringSpec struct {
	// Enabled enables or disables monitoring
	// +kubebuilder:default=true
	Enabled bool `json:"enabled,omitempty"`

	// MonitoringConfigRef references a MonitoringConfig CR
	// +optional
	MonitoringConfigRef *corev1.LocalObjectReference `json:"monitoringConfigRef,omitempty"`

	// PMM specifies Percona Monitoring and Management configuration
	// +optional
	PMM *PMMSpec `json:"pmm,omitempty"`
}

// PMMSpec defines PMM configuration
type PMMSpec struct {
	// ServerHost is the PMM server host
	ServerHost string `json:"serverHost"`

	// ServerUser is the PMM server username
	// +optional
	ServerUser string `json:"serverUser,omitempty"`

	// ServerPasswordSecretRef references a secret containing PMM password
	// +optional
	ServerPasswordSecretRef *corev1.SecretKeySelector `json:"serverPasswordSecretRef,omitempty"`
}

// ProxySpec defines proxy configuration
type ProxySpec struct {
	// Enabled enables or disables proxy
	// +kubebuilder:default=false
	Enabled bool `json:"enabled,omitempty"`

	// Type is the proxy type (pgbouncer, haproxy, proxysql, etc.)
	// +optional
	Type string `json:"type,omitempty"`

	// Replicas is the number of proxy instances
	// +kubebuilder:default=2
	Replicas int32 `json:"replicas,omitempty"`

	// Resources specifies the compute resources for proxy instances
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

// ExposeSpec defines how to expose the database
type ExposeSpec struct {
	// Type is the service type (ClusterIP, NodePort, LoadBalancer)
	// +kubebuilder:default=ClusterIP
	// +kubebuilder:validation:Enum=ClusterIP;NodePort;LoadBalancer
	Type corev1.ServiceType `json:"type,omitempty"`

	// Annotations for the service
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// LoadBalancerSourceRanges for LoadBalancer type
	// +optional
	LoadBalancerSourceRanges []string `json:"loadBalancerSourceRanges,omitempty"`
}

// PodSchedulingPolicySpec defines pod scheduling constraints
type PodSchedulingPolicySpec struct {
	// NodeSelector specifies node selection constraints
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Affinity specifies pod affinity/anti-affinity
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// Tolerations specifies pod tolerations
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// PriorityClassName specifies the priority class
	// +optional
	PriorityClassName string `json:"priorityClassName,omitempty"`
}

// ConfigParameter defines a configuration key-value pair
type ConfigParameter struct {
	// Name is the configuration parameter name
	Name string `json:"name"`

	// Value is the configuration parameter value
	Value string `json:"value"`
}

// DataSourceSpec defines the data source for initialization
type DataSourceSpec struct {
	// BackupSource specifies restoration from a backup
	// +optional
	BackupSource *BackupSourceSpec `json:"backupSource,omitempty"`

	// CloneSource specifies cloning from another cluster
	// +optional
	CloneSource *CloneSourceSpec `json:"cloneSource,omitempty"`
}

// BackupSourceSpec defines backup restoration source
type BackupSourceSpec struct {
	// BackupName is the name of the backup to restore from
	BackupName string `json:"backupName"`
}

// CloneSourceSpec defines cluster cloning source
type CloneSourceSpec struct {
	// ClusterName is the name of the cluster to clone from
	ClusterName string `json:"clusterName"`

	// Timestamp is the point-in-time to clone from
	// +optional
	Timestamp *metav1.Time `json:"timestamp,omitempty"`
}

// DatabaseClusterStatus defines the observed state of DatabaseCluster
type DatabaseClusterStatus struct {
	// Conditions represent the latest available observations of the cluster's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Phase is the current phase of the cluster (Initializing, Ready, Failed, etc.)
	// +optional
	Phase ClusterPhase `json:"phase,omitempty"`

	// Database contains database-specific status
	// +optional
	Database *DatabaseStatus `json:"database,omitempty"`

	// Proxy contains proxy status
	// +optional
	Proxy *ProxyStatus `json:"proxy,omitempty"`

	// Backup contains backup status
	// +optional
	Backup *BackupStatus `json:"backup,omitempty"`

	// Monitoring contains monitoring status
	// +optional
	Monitoring *MonitoringStatus `json:"monitoring,omitempty"`

	// ObservedGeneration is the generation observed by the controller
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Message provides additional information about the current state
	// +optional
	Message string `json:"message,omitempty"`
}

// ClusterPhase represents the current phase of the cluster
// +kubebuilder:validation:Enum=Initializing;Ready;Updating;Failed;Deleting
type ClusterPhase string

const (
	ClusterPhaseInitializing ClusterPhase = "Initializing"
	ClusterPhaseReady        ClusterPhase = "Ready"
	ClusterPhaseUpdating     ClusterPhase = "Updating"
	ClusterPhaseFailed       ClusterPhase = "Failed"
	ClusterPhaseDeleting     ClusterPhase = "Deleting"
)

// DatabaseStatus contains database-specific status information
type DatabaseStatus struct {
	// Ready indicates if the database is ready to accept connections
	Ready bool `json:"ready"`

	// Instances is the current number of database instances
	Instances int32 `json:"instances"`

	// ReadyInstances is the number of ready instances
	ReadyInstances int32 `json:"readyInstances"`

	// PrimaryInstance is the name of the primary/master instance
	// +optional
	PrimaryInstance string `json:"primaryInstance,omitempty"`

	// Roles contains instance role information
	// +optional
	Roles map[string]string `json:"roles,omitempty"`

	// Endpoints contains database endpoints
	// +optional
	Endpoints *DatabaseEndpoints `json:"endpoints,omitempty"`
}

// DatabaseEndpoints contains database connection endpoints
type DatabaseEndpoints struct {
	// Primary is the endpoint for write operations
	// +optional
	Primary string `json:"primary,omitempty"`

	// Replica is the endpoint for read operations
	// +optional
	Replica string `json:"replica,omitempty"`

	// External is the external endpoint (if exposed)
	// +optional
	External string `json:"external,omitempty"`
}

// ProxyStatus contains proxy status information
type ProxyStatus struct {
	// Ready indicates if the proxy is ready
	Ready bool `json:"ready"`

	// Replicas is the current number of proxy replicas
	Replicas int32 `json:"replicas"`

	// ReadyReplicas is the number of ready replicas
	ReadyReplicas int32 `json:"readyReplicas"`
}

// BackupStatus contains backup status information
type BackupStatus struct {
	// LastBackupTime is the timestamp of the last successful backup
	// +optional
	LastBackupTime *metav1.Time `json:"lastBackupTime,omitempty"`

	// NextBackupTime is the scheduled time for the next backup
	// +optional
	NextBackupTime *metav1.Time `json:"nextBackupTime,omitempty"`

	// LastBackupName is the name of the last backup
	// +optional
	LastBackupName string `json:"lastBackupName,omitempty"`
}

// MonitoringStatus contains monitoring status information
type MonitoringStatus struct {
	// Enabled indicates if monitoring is enabled
	Enabled bool `json:"enabled"`

	// Ready indicates if monitoring is ready
	Ready bool `json:"ready"`

	// Endpoint is the monitoring endpoint
	// +optional
	Endpoint string `json:"endpoint,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=dbc
// +kubebuilder:printcolumn:name="Engine",type=string,JSONPath=`.spec.engine.type`
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.engine.version`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.database.ready`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// DatabaseCluster is the Schema for the databaseclusters API
type DatabaseCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatabaseClusterSpec   `json:"spec,omitempty"`
	Status DatabaseClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DatabaseClusterList contains a list of DatabaseCluster
type DatabaseClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DatabaseCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DatabaseCluster{}, &DatabaseClusterList{})
}
