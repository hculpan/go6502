package emulator

import (
	"github.com/ariejan/i6502"
	"github.com/hculpan/go6502/screen"
)

// Emulator encapsulates the 6502 computer
type Emulator struct {
	CPU        *i6502.Cpu
	Active     bool
	Key        rune
	KeyWaiting bool
	SingleStep bool

	keyboardInterface *KeyboardInterface

	stepWait bool
	done     bool
}

// NewEmulator create a new emulator
func NewEmulator(scr *screen.Screen) *Emulator {
	result := &Emulator{}

	ram1, err := i6502.NewRam(0x8000) // 32k
	if err != nil {
		panic(err)
	}
	scri, err := NewScreenInterface(0x8000, scr)
	if err != nil {
		panic(err)
	}
	ram2, err := i6502.NewRam(0x7C00) // 31k
	if err != nil {
		panic(err)
	}
	result.keyboardInterface = NewKeyboardInterface()

	bus, _ := i6502.NewAddressBus()
	bus.Attach(ram1, 0x0000)
	bus.Attach(scri, 0x8000)
	bus.Attach(result.keyboardInterface, 0x8001)
	bus.Attach(ram2, 0x8400)
	result.CPU, _ = i6502.NewCpu(bus)
	result.Active = false

	result.SingleStep = false
	result.stepWait = false
	result.KeyWaiting = false

	return result
}

// GetCPU returns a reference to the CPU
func (e Emulator) GetCPU() *i6502.Cpu {
	return e.CPU
}

// ReadMemory allows the caller to read data from memory
func (e Emulator) ReadMemory(address uint16) uint8 {
	return e.CPU.Bus.ReadByte(address)
}

// WriteMemory allows the caller to write data to a specific
// location in memory
func (e *Emulator) WriteMemory(address uint16, data uint8) {
	e.CPU.Bus.WriteByte(address, data)
}

// SetKeyWaiting sets the next key that is waiting to be
// read by the emulator
func (e *Emulator) SetKeyWaiting(k rune) {
	e.keyboardInterface.Key = k
	e.keyboardInterface.KeyWaiting = true
	if e.CPU.P&0b00000100 == 0 {
		e.CPU.Interrupt()
	}
}

// Terminate terminates the emulator
func (e *Emulator) Terminate() {
	e.done = true
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

// Step allows the CPU to process the next
// clock tick
func (e *Emulator) Step() {
	if !e.stepWait {
		e.CPU.Step()
		if e.SingleStep {
			e.stepWait = true
		}
	}
}

// StartEmulator starts the emulator in a separate goroutine
func (e *Emulator) StartEmulator() {
	if e.Active {
		return
	}
	e.Active = true
	e.done = false

	e.CPU.Reset()

	/*	go func() {
			defer func() {
				e.Active = false
			}()
			for {
				if e.done {
					return
				}

				if !e.stepWait {
					e.CPU.Step()
					if e.SingleStep {
						e.stepWait = true
					}
					time.Sleep(1 * time.Microsecond)
				}
			}
		}()
	*/
}
