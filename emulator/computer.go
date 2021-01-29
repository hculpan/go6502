package emulator

import (
	"time"

	"github.com/ariejan/i6502"
)

// Emulator encapsulates the 6502 computer
type Emulator struct {
	CPU    *i6502.Cpu
	Active bool

	SingleStep bool

	stepWait bool

	done chan bool
}

// NewEmulator create a new emulator
func NewEmulator() *Emulator {
	result := &Emulator{}

	ram1, _ := i6502.NewRam(0x10000) // 64k
	//	ram2, _ := i6502.NewRam(0x7C00) // 31k
	bus, _ := i6502.NewAddressBus()
	bus.Attach(ram1, 0x0000)
	//	bus.Attach(ram2, 0x8400)
	result.CPU, _ = i6502.NewCpu(bus)
	result.Active = false

	result.SingleStep = false
	result.stepWait = false

	return result
}

// WriteMemory allows the caller to write data to a specific
// location in memory
func (e *Emulator) WriteMemory(address uint16, data uint8) {
	e.CPU.Bus.WriteByte(address, data)
}

// ReadMemory allows the caller to read data from memory
func (e *Emulator) ReadMemory(address uint16) uint8 {
	return e.CPU.Bus.ReadByte(address)
}

// Terminate terminates the emulator
func (e *Emulator) Terminate() {
	e.done <- true
}

// EnableSingleStep turns on single stepping through instructions
func (e *Emulator) EnableSingleStep() {
	e.SingleStep = true
	e.stepWait = true
}

// DisableSingleStep turns on single stepping through instructions
// and resumes normal processing
func (e *Emulator) DisableSingleStep() {
	e.SingleStep = false
	e.stepWait = false
}

// NextStep processes the next cpu step
// Useful only when SingleStep is enabled
func (e *Emulator) NextStep() {
	e.stepWait = false
}

// StartEmulator starts the emulator in a separate goroutine
// t indicates the pause in microseconds
func (e *Emulator) StartEmulator(t int64) {
	if e.Active {
		return
	}
	e.Active = true

	e.CPU.Reset()

	ticker := time.NewTicker(time.Duration(t) * time.Microsecond)
	e.done = make(chan bool)

	go func() {
		defer func() {
			e.Active = false
		}()
		for {
			select {
			case <-e.done:
				return
			case <-ticker.C:
				if !e.stepWait {
					e.CPU.Step()
					if e.SingleStep {
						e.stepWait = true
					}
				}
			}
		}
	}()

}
