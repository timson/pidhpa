package storage

import (
	"sync"

	"github.com/google/go-cmp/cmp"
	pidscalerv1 "github.com/timson/pidhpa-operator/api/v1"
)

// PIDScalerState is a struct that holds the parameters for the PID controller
type PIDScalerState struct {
	TargetSettings  pidscalerv1.TargetSettings
	PidSettings     pidscalerv1.PIDSettings
	KafkaSettings   pidscalerv1.KafkaSettings
	CooldownTimeout int32
	Interval        int32
	ControlCh       chan int
}

const (
	TargetSettingsMask = 1 << iota
	PidSettingsMask
	KafkaSettingsMask
	IntervalMask
	CooldownTimeoutMask
)

func NewPIDScalerState(pidScaler *pidscalerv1.PIDScaler) *PIDScalerState {
	scaler := &PIDScalerState{
		TargetSettings: pidscalerv1.TargetSettings{
			Deployment:  pidScaler.Spec.Target.Deployment,
			Namespace:   pidScaler.Spec.Target.Namespace,
			MinReplicas: pidScaler.Spec.Target.MinReplicas,
			MaxReplicas: pidScaler.Spec.Target.MaxReplicas,
		},
		PidSettings: pidscalerv1.PIDSettings{
			Kp:              pidScaler.Spec.PID.Kp,
			Ki:              pidScaler.Spec.PID.Ki,
			Kd:              pidScaler.Spec.PID.Kd,
			ReferenceSignal: pidScaler.Spec.PID.ReferenceSignal,
		},
		KafkaSettings: pidscalerv1.KafkaSettings{
			Topic:   pidScaler.Spec.Kafka.Topic,
			Group:   pidScaler.Spec.Kafka.Group,
			Brokers: pidScaler.Spec.Kafka.Brokers,
		},
		CooldownTimeout: pidScaler.Spec.CooldownTimeout,
		Interval:        pidScaler.Spec.Interval,
		ControlCh:       make(chan int),
	}
	return scaler
}

// GetDifferenceMask compare to PIDScalerState and calculate mask of changes
func (d *PIDScalerState) GetDifferenceMask(s *PIDScalerState) int {
	mask := 0
	if d.TargetSettings != s.TargetSettings {
		d.TargetSettings = s.TargetSettings
		mask |= TargetSettingsMask
	}

	if d.PidSettings != s.PidSettings {
		d.PidSettings = s.PidSettings
		mask |= PidSettingsMask
	}

	if !cmp.Equal(d.KafkaSettings, s.KafkaSettings) {
		d.KafkaSettings = s.KafkaSettings
		mask |= KafkaSettingsMask
	}

	if d.Interval != s.Interval {
		d.Interval = s.Interval
		mask |= IntervalMask
	}

	if d.CooldownTimeout != s.CooldownTimeout {
		d.CooldownTimeout = s.CooldownTimeout
		mask |= CooldownTimeoutMask
	}

	return mask
}

// PIDScalerStateStorage is a storage for PIDScaler objects
type PIDScalerStateStorage struct {
	storage map[string]*PIDScalerState
	lock    sync.RWMutex
}

// NewPIDScalerStorage creates a new PIDScalerStorage object
// and returns a pointer to it
func NewPIDScalerStorage() *PIDScalerStateStorage {
	return &PIDScalerStateStorage{
		storage: make(map[string]*PIDScalerState, 0),
	}
}

// AddOrUpdate adds a PIDScaler object to the storage
func (s *PIDScalerStateStorage) AddOrUpdate(key string, value *PIDScalerState) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.storage[key] = value
}

// Get returns a PIDScaler object from the storage
func (s *PIDScalerStateStorage) Get(key string) (*PIDScalerState, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	value, ok := s.storage[key]
	if !ok {
		return nil, false
	}
	return value, true
}

// Delete deletes a PIDScaler object from the storage
func (s *PIDScalerStateStorage) Delete(key string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	scaler, ok := s.storage[key]
	if ok && scaler.ControlCh != nil {
		close(scaler.ControlCh)
	}
	delete(s.storage, key)
}
