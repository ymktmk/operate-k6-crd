// To add a yaml tag to the structure(K6) defined here
package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type EnvVar struct {
	Name      string       `yaml:"name" protobuf:"bytes,1,opt,name=name"`
	Value     string       `yaml:"value,omitempty" protobuf:"bytes,2,opt,name=value"`
	ValueFrom EnvVarSource `yaml:"valueFrom,omitempty" protobuf:"bytes,3,opt,name=valueFrom"`
}

type EnvVarSource struct {
	SecretKeyRef SecretKeySelector `yaml:"secretKeyRef,omitempty" protobuf:"bytes,4,opt,name=secretKeyRef"`
}

type SecretKeySelector struct {
	LocalObjectReference `yaml:",inline" protobuf:"bytes,1,opt,name=localObjectReference"`
	Key                  string `yaml:"key" protobuf:"bytes,2,opt,name=key"`
}

type LocalObjectReference struct {
	Name string `yaml:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
}

type PodMetadata struct {
	Annotations map[string]string `yaml:"annotations,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
}

// Env []corev1.EnvVar
type Pod struct {
	Affinity                     *corev1.Affinity              `yaml:"affinity,omitempty"`
	AutomountServiceAccountToken string                        `yaml:"automountServiceAccountToken,omitempty"`
	Env                          []EnvVar                      `yaml:"env,omitempty"`
	Image                        string                        `yaml:"image,omitempty"`
	ImagePullSecrets             []corev1.LocalObjectReference `yaml:"imagePullSecrets,omitempty"`
	ImagePullPolicy              corev1.PullPolicy             `yaml:"imagePullPolicy,omitempty"`
	Metadata                     PodMetadata                   `yaml:"metadata,omitempty"`
	NodeSelector                 map[string]string             `yaml:"nodeselector,omitempty"`
	Tolerations                  []corev1.Toleration           `yaml:"tolerations,omitempty"`
	Resources                    corev1.ResourceRequirements   `yaml:"resources,omitempty"`
	ServiceAccountName           string                        `yaml:"serviceAccountName,omitempty"`
	SecurityContext              corev1.PodSecurityContext     `yaml:"securityContext,omitempty"`
	EnvFrom                      []corev1.EnvFromSource        `yaml:"envFrom,omitempty"`
}

type K6Scuttle struct {
	Enabled                 string `yaml:"enabled,omitempty"`
	EnvoyAdminApi           string `yaml:"envoyAdminApi,omitempty"`
	NeverKillIstio          bool   `yaml:"neverKillIstio,omitempty"`
	NeverKillIstioOnFailure bool   `yaml:"neverKillIstioOnFailure,omitempty"`
	ScuttleLogging          bool   `yaml:"scuttleLogging,omitempty"`
	StartWithoutEnvoy       bool   `yaml:"startWithoutEnvoy,omitempty"`
	WaitForEnvoyTimeout     string `yaml:"waitForEnvoyTimeout,omitempty"`
	IstioQuitApi            string `yaml:"istioQuitApi,omitempty"`
	GenericQuitEndpoint     string `yaml:"genericQuitEndpoint,omitempty"`
	QuitWithoutEnvoyTimeout string `yaml:"quitWithoutEnvoyTimeout,omitempty"`
}

type K6Spec struct {
	Script      K6Script               `yaml:"script"`
	Parallelism int32                  `yaml:"parallelism"`
	Separate    bool                   `yaml:"separate,omitempty"`
	Arguments   string                 `yaml:"arguments,omitempty"`
	Ports       []corev1.ContainerPort `yaml:"ports,omitempty"`
	Starter     Pod                    `yaml:"starter,omitempty"`
	Runner      Pod                    `yaml:"runner,omitempty"`
	Quiet       string                 `yaml:"quiet,omitempty"`
	Paused      string                 `yaml:"paused,omitempty"`
	Scuttle     K6Scuttle              `yaml:"scuttle,omitempty"`
	Cleanup     Cleanup                `yaml:"cleanup,omitempty"`
}

type K6Script struct {
	VolumeClaim K6VolumeClaim `yaml:"volumeClaim,omitempty"`
	ConfigMap   K6Configmap   `yaml:"configMap,omitempty"`
	LocalFile   string        `yaml:"localFile,omitempty"`
}

type K6VolumeClaim struct {
	Name string `yaml:"name"`
	File string `yaml:"file,omitempty"`
}

type K6Configmap struct {
	Name string `yaml:"name"`
	File string `yaml:"file,omitempty"`
}

type Cleanup string

type Stage string

type K6Status struct {
	Stage Stage `yaml:"stage,omitempty"`
}

type K6 struct {
	metav1.TypeMeta   `yaml:",inline"`
	metav1.ObjectMeta `yaml:"metadata,omitempty"`

	Spec   K6Spec   `yaml:"spec,omitempty"`
	Status K6Status `yaml:"status,omitempty"`
}

type K6List struct {
	metav1.TypeMeta `yaml:",inline"`
	metav1.ListMeta `yaml:"metadata,omitempty"`
	Items           []K6 `yaml:"items"`
}
