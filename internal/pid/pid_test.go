package pid

import (
	"testing"
	"time"
)

func TestPIDInitialization(t *testing.T) {
	pid := NewPID(1.0, 0.5, 0.1, 0, 100, false)

	if pid.Kp != 1.0 || pid.Ki != 0.5 || pid.Kd != 0.1 {
		t.Errorf("Unexpected PID gains. Got: Kp=%f, Ki=%f, Kd=%f", pid.Kp, pid.Ki, pid.Kd)
	}
	if pid.minOutput != 0 || pid.maxOutput != 100 {
		t.Errorf("Unexpected output limits. Got: min=%f, max=%f", pid.minOutput, pid.maxOutput)
	}
	if pid.Reverse {
		t.Errorf("Reverse mode should be false by default")
	}
}

func TestPIDUpdateConfig(t *testing.T) {
	pid := NewPID(1.0, 0.5, 0.1, 0, 100, false)
	pid.UpdateConfig(2.0, 1.0, 0.2, 10, 200, true)

	if pid.Kp != 2.0 || pid.Ki != 1.0 || pid.Kd != 0.2 {
		t.Errorf("Unexpected PID gains after update. Got: Kp=%f, Ki=%f, Kd=%f", pid.Kp, pid.Ki, pid.Kd)
	}
	if pid.minOutput != 10 || pid.maxOutput != 200 {
		t.Errorf("Unexpected output limits after update. Got: min=%f, max=%f", pid.minOutput, pid.maxOutput)
	}
	if !pid.Reverse {
		t.Errorf("Reverse mode should be true after update")
	}
}

func TestPIDUpdate(t *testing.T) {
	pid := NewPID(1.0, 0.5, 0.1, 0, 100, false)

	// First update (no previous time or error)
	output := pid.Update(50, 25, time.Now())
	if output <= 0 {
		t.Errorf("Expected positive output. Got: %f", output)
	}

	// Second update (with previous time and error)
	for i := 0; i < 3; i++ {
		output = pid.Update(50, 25, time.Now())
		time.Sleep(time.Second * 1)
	}
	if output <= 0 {
		t.Errorf("Expected positive output. Got: %f", output)
	}
}

func TestPIDClamping(t *testing.T) {
	pid := NewPID(1.0, 0.5, 0.1, 10, 50, false)

	// Simulate large error to test clamping
	output := pid.Update(100, 0, time.Now())
	if output != 50 {
		t.Errorf("Output should be clamped to maxOutput. Got: %f", output)
	}

	// Simulate large error in the negative direction
	output = pid.Update(0, 100, time.Now())
	if output != 10 {
		t.Errorf("Output should be clamped to minOutput. Got: %f", output)
	}
}

func TestPIDAntiWindup(t *testing.T) {
	pid := NewPID(1.0, 0.5, 0.1, 0, 100, false)

	// Simulate saturation and check that integral doesn't grow excessively
	output := pid.Update(100, 0, time.Now())
	if output != 100 {
		t.Errorf("Expected output to be clamped to maxOutput. Got: %f", output)
	}

	// Check if integral is conditionally updated
	pid.Update(0, 100, time.Now())
	if pid.integral != 0 {
		t.Errorf("Integral term should not grow excessively during saturation")
	}
}
