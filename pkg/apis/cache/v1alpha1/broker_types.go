package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PodPolicy defines the policy for pods owned by RockerMQ operator.
type PodPolicy struct {
	// Resources is the resource requirements for the containers.
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// PersistentVolumeClaim is the claim to describe PVC for the RockerMQ container
	PersistentVolumeClaim *corev1.PersistentVolumeClaim `json:"persistentVolumeClaim,omitempty"`
}

// BrokerSpec defines the desired state of Broker
// +k8s:openapi-gen=true
type BrokerSpec struct {
	// Size is the number of the RocketMQ broker cluster. Default: 1
	Size int32 `json:"size"`
	// BrokerImage is the RocketMQ broker container image to use for the Pods.
	BrokerImage string `json:"brokerImage"`
	// ImagePullPolicy defines how the image is pulled.
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy"`
	// NameServers defines the name service list e.g. 192.168.1.1:9876;192.168.1.2:9876
	NameServers string `json:"nameServers,omitempty"`
	// SlavePerGroup defines how many slave brokers each broker group has
	SlavePerGroup int `json:"slavePerGroup,omitempty"`
	// ReplicationMode defines the replication is sync or async e.g. SYNC
	ReplicationMode string `json:"replicationMode,omitempty"`
	// Pod defines the policy for pods owned by RocketMQ operator.
	// This field cannot be updated once the CR is created.
	Pod *PodPolicy `json:"pod,omitempty"`
}

// BrokerStatus defines the observed state of Broker
// +k8s:openapi-gen=true
type BrokerStatus struct {
	Nodes []string `json:"nodes"`
	// PersistentVolumeClaimName is the name of the PVC backing the pods in the cluster.
	PersistentVolumeClaimName string `json:"persistentVolumeClaimName"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Broker is the Schema for the brokers API
// +k8s:openapi-gen=true
type Broker struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BrokerSpec   `json:"spec,omitempty"`
	Status BrokerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BrokerList contains a list of Broker
type BrokerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Broker `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Broker{}, &BrokerList{})
}
