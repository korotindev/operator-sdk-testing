package v1alpha1

import (
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ApplicationContainerPort struct {
	// +kubebuilder:validation:MaxLength=15
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:validation:Minimum=1
	HostPort int32 `json:"hostPort"`
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:validation:Minimum=1
	ContainerPort int32 `json:"containerPort"`
}

// ApplicationContainer defines specific container in application
type ApplicationContainer struct {
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
	// +kubebuilder:validation:MinLength=1
	Image string `json:"image"`
	// +kubebuilder:validation:MinItems=1
	Ports []ApplicationContainerPort `json:"ports"`
	// +kubebuilder:validation:MaxLength=15
	// +kubebuilder:validation:MinLength=1
	CPULimit string `json:"cpuLimit"`
	// +kubebuilder:validation:MaxLength=15
	// +kubebuilder:validation:MinLength=1
	MemoryLimit string `json:"memoryLimit"`
}

// ApplicationSpec defines the desired state of Application
type ApplicationSpec struct {
	// +kubebuilder:validation:MinItems=1
	Containers []ApplicationContainer `json:"containers"`
	// +kubebuilder:validation:Minimum=0
	Replicas *int32 `json:"replicas"`
}

// ApplicationStatus defines the observed state of Application
type ApplicationStatus struct {
	Replicas int32    `json:"replicas"`
	Pods     []string `json:"pods"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Application is the Schema for the applications API
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=applications,scope=Namespaced
type Application struct {
	metaV1.TypeMeta   `json:",inline"`
	metaV1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSpec   `json:"spec,omitempty"`
	Status ApplicationStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ApplicationList contains a list of Application
type ApplicationList struct {
	metaV1.TypeMeta `json:",inline"`
	metaV1.ListMeta `json:"metadata,omitempty"`
	Items           []Application `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Application{}, &ApplicationList{})
}
