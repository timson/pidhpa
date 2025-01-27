/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	StatusDeployed    = "Deployed"
	StatusUnknown     = "Unknown"
	StatusFailed      = "Failed"
	StatusInProgress  = "InProgress"
	StatusTerminating = "Terminating"
)

type OperatorStatus struct {
	Status     string      `json:"status"`
	Message    string      `json:"message,omitempty"`
	UpdateTime metav1.Time `json:"update_time,omitempty"`
}

type KafkaSettings struct {
	Brokers       []string `json:"brokers"`
	Topic         string   `json:"topic"`
	Group         string   `json:"group"`
	UseSASL       bool     `json:"use_sasl,omitempty"`
	SASLMechanism string   `json:"sasl_mechanism,omitempty"`
	Username      string   `json:"username,omitempty"`
	Password      string   `json:"password,omitempty"`
}

type TargetSettings struct {
	Deployment  string `json:"deployment"`
	Namespace   string `json:"namespace"`
	MinReplicas int32  `json:"min_replicas"`
	MaxReplicas int32  `json:"max_replicas"`
	// PID Regulator output will set desired replicas, and the controller will scale the deployment to this number in the next reconciliation
	DesiredReplicas *int32      `json:"desired_replicas,omitempty"`
	UpdateTime      metav1.Time `json:"update_time,omitempty"`
}

type PIDSettings struct {
	Ki              string `json:"ki"`
	Kp              string `json:"kp"`
	Kd              string `json:"kd"`
	ReferenceSignal int64  `json:"reference_signal"`
}

func (s *PIDSettings) getFloat(v string) float64 {
	floatValue, _ := strconv.ParseFloat(v, 64)
	return floatValue
}

func (s *PIDSettings) GetKp() float64 {
	return s.getFloat(s.Kp)
}

func (s *PIDSettings) GetKi() float64 {
	return s.getFloat(s.Ki)
}

func (s *PIDSettings) GetKd() float64 {
	return s.getFloat(s.Kd)
}

// PIDScalerSpec defines the desired state of PIDScaler
type PIDScalerSpec struct {
	Kafka           KafkaSettings  `json:"kafka"`
	PID             PIDSettings    `json:"pid"`
	Target          TargetSettings `json:"target"`
	Interval        int32          `json:"interval"`
	CooldownTimeout int32          `json:"cooldown_timeout"`
}

// PIDScalerStatus defines the observed state of PIDScaler
type PIDScalerStatus struct {
	Status     string      `json:"status"`
	Message    string      `json:"message,omitempty"`
	UpdateTime metav1.Time `json:"update_time,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PIDScaler is the Schema for the pidscalers API
type PIDScaler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PIDScalerSpec   `json:"spec,omitempty"`
	Status PIDScalerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PIDScalerList contains a list of PIDScaler
type PIDScalerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PIDScaler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PIDScaler{}, &PIDScalerList{})
}
