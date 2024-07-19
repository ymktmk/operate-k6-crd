package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TestRunSpec struct {
	Script      K6Script               `yaml:"script"`
	Parallelism int32                  `yaml:"parallelism"`
	Separate    bool                   `yaml:"separate,omitempty"`
	Arguments   string                 `yaml:"arguments,omitempty"`
	Ports       []corev1.ContainerPort `yaml:"ports,omitempty"`
	Initializer *Pod                   `yaml:"initializer,omitempty"`
	Starter     Pod                    `yaml:"starter,omitempty"`
	Runner      Pod                    `yaml:"runner,omitempty"`
	Quiet       string                 `yaml:"quiet,omitempty"`
	Paused      string                 `yaml:"paused,omitempty"`
	Scuttle     K6Scuttle              `yaml:"scuttle,omitempty"`
	Cleanup     Cleanup                `yaml:"cleanup,omitempty"`

	TestRunID string `yaml:"testRunId,omitempty"`
	Token     string `yaml:"token,omitempty"`
}

type TestRunStatus struct {
	Stage           Stage  `yaml:"stage,omitempty"`
	TestRunID       string `yaml:"testRunId,omitempty"`
	AggregationVars string `yaml:"aggregationVars,omitempty"`

	Conditions []metav1.Condition `yaml:"conditions,omitempty"`
}

type TestRun struct {
	metav1.TypeMeta   `yaml:",inline"`
	metav1.ObjectMeta `yaml:"metadata,omitempty"`

	Spec   TestRunSpec   `yaml:"spec,omitempty"`
	Status TestRunStatus `yaml:"status,omitempty"`
}

type TestRunList struct {
	metav1.TypeMeta `yaml:",inline"`
	metav1.ListMeta `yaml:"metadata,omitempty"`
	Items           []TestRun `yaml:"items"`
}
