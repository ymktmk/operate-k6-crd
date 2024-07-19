package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TestRunSpec struct {
	Script      K6Script               `json:"script"`
	Parallelism int32                  `json:"parallelism"`
	Separate    bool                   `json:"separate,omitempty"`
	Arguments   string                 `json:"arguments,omitempty"`
	Ports       []corev1.ContainerPort `json:"ports,omitempty"`
	Initializer *Pod                   `json:"initializer,omitempty"`
	Starter     Pod                    `json:"starter,omitempty"`
	Runner      Pod                    `json:"runner,omitempty"`
	Quiet       string                 `json:"quiet,omitempty"`
	Paused      string                 `json:"paused,omitempty"`
	Scuttle     K6Scuttle              `json:"scuttle,omitempty"`
	Cleanup     Cleanup                `json:"cleanup,omitempty"`

	TestRunID string `json:"testRunId,omitempty"`
	Token     string `json:"token,omitempty"`
}

type TestRunStatus struct {
	Stage           Stage  `json:"stage,omitempty"`
	TestRunID       string `json:"testRunId,omitempty"`
	AggregationVars string `json:"aggregationVars,omitempty"`

	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

type TestRun struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TestRunSpec   `json:"spec,omitempty"`
	Status TestRunStatus `json:"status,omitempty"`
}

type TestRunList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TestRun `json:"items"`
}
