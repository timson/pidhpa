package pid

import (
	"sync"
	"time"
)

const defaultDt = 1.0

// PID holds the controller parameters and state.
type PID struct {
	Kp, Ki, Kd float64 // Gains
	minOutput  float64 // Minimum output (1 pod)
	maxOutput  float64 // Maximum output (100 pods)

	integral  float64 // Integral accumulator
	prevError float64 // Previous error, for derivative
	prevTime  time.Time
	Reverse   bool // Reverse control direction
	mu        sync.Mutex
}

// NewPID returns a PID controller with given gains and output limits.
func NewPID(kp, ki, kd, minOut, maxOut float64, reverse bool) *PID {
	return &PID{
		Kp:        kp,
		Ki:        ki,
		Kd:        kd,
		minOutput: minOut,
		maxOutput: maxOut,
		Reverse:   reverse,
	}
}

func (pid *PID) UpdateConfig(kp, ki, kd, minOut, maxOut float64, reverse bool) {
	pid.mu.Lock()
	defer pid.mu.Unlock()
	pid.Kp = kp
	pid.Ki = ki
	pid.Kd = kd
	pid.minOutput = minOut
	pid.maxOutput = maxOut
	pid.Reverse = reverse
}

// Update computes the new controller output given setpoint (sp) and measured value (pv).
// 'now' is the current time; pass time.Now() in real usage.
func (pid *PID) Update(sp, pv float64, now time.Time) float64 {
	pid.mu.Lock()
	defer pid.mu.Unlock()

	var dt float64
	if !pid.prevTime.IsZero() {
		dt = now.Sub(pid.prevTime).Seconds()
	}
	pid.prevTime = now
	if dt <= 0 {
		dt = defaultDt
	}

	// Calculate error
	// If Reverse is true, the controller tries to keep the measured value below the setpoint
	var err float64
	if pid.Reverse {
		err = pv - sp
	} else {
		err = sp - pv
	}

	p := pid.Kp * err

	newIntegral := pid.integral + err*dt
	d := pid.Kd * ((err - pid.prevError) / dt)
	unclampedOutput := p + pid.Ki*newIntegral + d

	// Check for saturation (anti-windup)
	var output float64
	if unclampedOutput > pid.maxOutput {
		output = pid.maxOutput
		// If err is driving output beyond max, do not integrate further
		// (conditional integration)
		if err < 0 {
			// If error is negative, it might help drive output back within limits
			pid.integral = newIntegral
		}
		// else do not update integral, to avoid wind-up
	} else if unclampedOutput < pid.minOutput {
		output = pid.minOutput
		// If err is positive, it might help drive output back within limits
		if err > 0 {
			pid.integral = newIntegral
		}
	} else {
		// Within limits, accept
		output = unclampedOutput
		pid.integral = newIntegral
	}

	// Save state
	pid.prevError = err

	return output
}
