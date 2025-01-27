package storage

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	pidscalerv1 "github.com/timson/pidhpa-operator/api/v1"
)

func TestScalerUpdateFrom(t *testing.T) {
	tests := []struct {
		name     string
		initial  PIDScalerState
		updated  PIDScalerState
		expected int
	}{
		{
			name: "No changes",
			initial: PIDScalerState{
				TargetSettings: pidscalerv1.TargetSettings{
					Deployment:  "app",
					Namespace:   "default",
					MinReplicas: 1,
					MaxReplicas: 5,
				},
				PidSettings: pidscalerv1.PIDSettings{
					Ki: "1.0", Kp: "0.5", Kd: "0.1", ReferenceSignal: 100,
				},
				KafkaSettings: pidscalerv1.KafkaSettings{
					Brokers: []string{"broker1"}, Topic: "topic", Group: "group",
				},
				Interval:        10,
				CooldownTimeout: 30,
			},
			updated: PIDScalerState{
				TargetSettings: pidscalerv1.TargetSettings{
					Deployment:  "app",
					Namespace:   "default",
					MinReplicas: 1,
					MaxReplicas: 5,
				},
				PidSettings: pidscalerv1.PIDSettings{
					Ki: "1.0", Kp: "0.5", Kd: "0.1", ReferenceSignal: 100,
				},
				KafkaSettings: pidscalerv1.KafkaSettings{
					Brokers: []string{"broker1"}, Topic: "topic", Group: "group",
				},
				Interval:        10,
				CooldownTimeout: 30,
			},
			expected: 0,
		},
		{
			name: "Change TargetSettings",
			initial: PIDScalerState{
				TargetSettings: pidscalerv1.TargetSettings{
					Deployment:  "app",
					Namespace:   "default",
					MinReplicas: 1,
					MaxReplicas: 5,
				},
			},
			updated: PIDScalerState{
				TargetSettings: pidscalerv1.TargetSettings{
					Deployment:  "app2",
					Namespace:   "default",
					MinReplicas: 1,
					MaxReplicas: 5,
				},
			},
			expected: TargetSettingsMask,
		},
		{
			name: "Change PIDSettings",
			initial: PIDScalerState{
				PidSettings: pidscalerv1.PIDSettings{
					Ki: "1.0", Kp: "0.5", Kd: "0.1", ReferenceSignal: 100,
				},
			},
			updated: PIDScalerState{
				PidSettings: pidscalerv1.PIDSettings{
					Ki: "2.0", Kp: "0.5", Kd: "0.1", ReferenceSignal: 100,
				},
			},
			expected: PidSettingsMask,
		},
		{
			name: "Change KafkaSettings",
			initial: PIDScalerState{
				KafkaSettings: pidscalerv1.KafkaSettings{
					Brokers: []string{"broker1"}, Topic: "topic", Group: "group",
				},
			},
			updated: PIDScalerState{
				KafkaSettings: pidscalerv1.KafkaSettings{
					Brokers: []string{"broker2"}, Topic: "topic", Group: "group",
				},
			},
			expected: KafkaSettingsMask,
		},
		{
			name: "Change Interval",
			initial: PIDScalerState{
				Interval: 10,
			},
			updated: PIDScalerState{
				Interval: 20,
			},
			expected: IntervalMask,
		},
		{
			name: "Change CooldownTimeout",
			initial: PIDScalerState{
				CooldownTimeout: 30,
			},
			updated: PIDScalerState{
				CooldownTimeout: 60,
			},
			expected: CooldownTimeoutMask,
		},
		{
			name: "Multiple changes",
			initial: PIDScalerState{
				TargetSettings: pidscalerv1.TargetSettings{
					Deployment: "app1",
				},
				Interval:        10,
				CooldownTimeout: 30,
			},
			updated: PIDScalerState{
				TargetSettings: pidscalerv1.TargetSettings{
					Deployment: "app2",
				},
				Interval:        20,
				CooldownTimeout: 60,
			},
			expected: TargetSettingsMask | IntervalMask | CooldownTimeoutMask,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			differenceMask := tt.initial.GetDifferenceMask(&tt.updated)
			if differenceMask != tt.expected {
				t.Errorf("UpdateFrom() differenceMask = %b, expected = %b", differenceMask, tt.expected)
			}

			if tt.expected&TargetSettingsMask != 0 && !cmp.Equal(tt.initial.TargetSettings, tt.updated.TargetSettings) {
				t.Errorf("TargetSettings not updated correctly")
			}

			if tt.expected&PidSettingsMask != 0 && !cmp.Equal(tt.initial.PidSettings, tt.updated.PidSettings) {
				t.Errorf("PidSettings not updated correctly")
			}

			if tt.expected&KafkaSettingsMask != 0 && !cmp.Equal(tt.initial.KafkaSettings, tt.updated.KafkaSettings) {
				t.Errorf("KafkaSettings not updated correctly")
			}

			if tt.expected&IntervalMask != 0 && tt.initial.Interval != tt.updated.Interval {
				t.Errorf("Interval not updated correctly")
			}

			if tt.expected&CooldownTimeoutMask != 0 && tt.initial.CooldownTimeout != tt.updated.CooldownTimeout {
				t.Errorf("CooldownTimeout not updated correctly")
			}
		})
	}
}
