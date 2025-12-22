package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BackupStorageSpec defines the desired state of BackupStorage
type BackupStorageSpec struct {
	// Type is the storage type (s3, gcs, azure, nfs, etc.)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=s3;gcs;azure;nfs;local
	Type string `json:"type"`

	// S3 contains S3-compatible storage configuration
	// +optional
	S3 *S3StorageSpec `json:"s3,omitempty"`

	// GCS contains Google Cloud Storage configuration
	// +optional
	GCS *GCSStorageSpec `json:"gcs,omitempty"`

	// Azure contains Azure Blob Storage configuration
	// +optional
	Azure *AzureStorageSpec `json:"azure,omitempty"`

	// NFS contains NFS storage configuration
	// +optional
	NFS *NFSStorageSpec `json:"nfs,omitempty"`

	// CredentialsSecretRef references a secret containing storage credentials
	// +optional
	CredentialsSecretRef *corev1.LocalObjectReference `json:"credentialsSecretRef,omitempty"`

	// VerifyTLS indicates whether to verify TLS certificates
	// +kubebuilder:default=true
	VerifyTLS bool `json:"verifyTLS,omitempty"`
}

// S3StorageSpec defines S3-compatible storage configuration
type S3StorageSpec struct {
	// Endpoint is the S3 endpoint URL
	// +kubebuilder:validation:Required
	Endpoint string `json:"endpoint"`

	// Bucket is the S3 bucket name
	// +kubebuilder:validation:Required
	Bucket string `json:"bucket"`

	// Region is the S3 region
	// +optional
	Region string `json:"region,omitempty"`

	// Prefix is the path prefix within the bucket
	// +optional
	Prefix string `json:"prefix,omitempty"`
}

// GCSStorageSpec defines Google Cloud Storage configuration
type GCSStorageSpec struct {
	// Bucket is the GCS bucket name
	// +kubebuilder:validation:Required
	Bucket string `json:"bucket"`

	// Prefix is the path prefix within the bucket
	// +optional
	Prefix string `json:"prefix,omitempty"`
}

// AzureStorageSpec defines Azure Blob Storage configuration
type AzureStorageSpec struct {
	// StorageAccount is the Azure storage account name
	// +kubebuilder:validation:Required
	StorageAccount string `json:"storageAccount"`

	// Container is the Azure blob container name
	// +kubebuilder:validation:Required
	Container string `json:"container"`

	// Prefix is the path prefix within the container
	// +optional
	Prefix string `json:"prefix,omitempty"`
}

// NFSStorageSpec defines NFS storage configuration
type NFSStorageSpec struct {
	// Server is the NFS server address
	// +kubebuilder:validation:Required
	Server string `json:"server"`

	// Path is the NFS export path
	// +kubebuilder:validation:Required
	Path string `json:"path"`
}

// BackupStorageStatus defines the observed state of BackupStorage
type BackupStorageStatus struct {
	// Conditions represent the latest available observations of the storage's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Phase is the current phase of the storage (Ready, NotReady)
	// +optional
	Phase string `json:"phase,omitempty"`

	// Message provides additional information about the current state
	// +optional
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=bs
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// BackupStorage is the Schema for the backupstorages API
type BackupStorage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackupStorageSpec   `json:"spec,omitempty"`
	Status BackupStorageStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BackupStorageList contains a list of BackupStorage
type BackupStorageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackupStorage `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BackupStorage{}, &BackupStorageList{})
}
