package emulator

// KeyboardInterface is the Memory (see i6502) component to write
// to the screen
type KeyboardInterface struct {
	KeyWaiting bool
	Key        rune
}

// NewKeyboardInterface returns a new screen interface
func NewKeyboardInterface() *KeyboardInterface {
	return &KeyboardInterface{}
}

// Size returns the constant 1
func (s *KeyboardInterface) Size() uint16 {
	return 1
}

// ReadByte reads the memory from the screen interface
func (s *KeyboardInterface) ReadByte(address uint16) byte {
	var result byte = 0
	if s.KeyWaiting {
		result = byte(s.Key)
		s.KeyWaiting = false
	}
	return result
}

// WriteByte writes to the screen interface memory location
func (s *KeyboardInterface) WriteByte(address uint16, data byte) {
	// nothing
}
