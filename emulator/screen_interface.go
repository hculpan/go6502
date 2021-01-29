package emulator

import (
	"github.com/hculpan/go6502/screen"
)

const size = 1

// ScreenInterface is the Memory (see i6502) component to write
// to the screen
type ScreenInterface struct {
	StartAddress uint16
	Scr          *screen.Screen
}

// NewScreenInterface returns a new screen interface
func NewScreenInterface(startAddress uint16, scr *screen.Screen) (*ScreenInterface, error) {
	return &ScreenInterface{StartAddress: startAddress, Scr: scr}, nil
}

// Size returns the constant 1
func (s *ScreenInterface) Size() uint16 {
	return 1
}

// ReadByte reads the memory from the screen interface
func (s *ScreenInterface) ReadByte(address uint16) byte {
	if s.Scr.Busy {
		return 1
	} else {
		return 0
	}
}

// WriteByte writes to the screen interface memory location
func (s *ScreenInterface) WriteByte(address uint16, data byte) {
	s.Scr.ProcessRune(rune(data))
}
