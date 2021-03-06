package dcpu

import (
	"./core"
	"errors"
	"time"
)

type Machine struct {
	State   core.State
	stopper chan<- struct{}
	stopped <-chan error
}

// Start boots up the machine, with a clock rate of 1 / period
// 10MHz would be expressed as (Microsecond / 10)
func (m *Machine) Start(period time.Duration) error {
	if err := m.State.Start(); err != nil {
		return err
	}
	stopper := make(chan struct{})
	m.stopper = stopper
	stopped := make(chan error)
	m.stopped = stopped
	go func() {
		ticker := time.NewTicker(period)
		for {
			select {
			case _ = <-ticker.C:
				if err := m.State.StepCycle(); err != nil {
					stopped <- err
					break
				}
			case _ = <-stopper:
				if err := m.State.Stop(); err != nil {
					stopped <- err
				}
				break
			}
		}
		ticker.Stop()
		close(stopped)
	}()
	return nil
}

// Stop stops the machine. Returns an error if it's already stopped.
// If the machine has halted due to an error, that error is returned.
func (m *Machine) Stop() error {
	if err := m.State.Stop(); err != nil {
		return err
	}
	if err := <-m.stopped; err != nil {
		return err
	}
	close(m.stopper)
	m.stopper = nil
	m.stopped = nil
	return nil
}

// If the machine has already halted due to an error, that error is returned.
// Otherwise, nil is returned.
// If the machine has not started, an error is returned.
func (m *Machine) HasError() error {
	if m.stopped == nil {
		return errors.New("Machine has not started")
	}
	select {
	case err := <-m.stopped:
		m.State.Stop()
		close(m.stopper)
		m.stopper = nil
		m.stopped = nil
		return err
	default:
	}
	return nil
}
