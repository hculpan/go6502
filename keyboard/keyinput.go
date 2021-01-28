package keyboard

import "github.com/veandco/go-sdl2/sdl"

// KeyInput is a single key that is sent
// to the screen object
type KeyInput struct {
	Key  rune
	Type uint32
}

// NewKeyInputFromEvent creates a new key input from a keyboard event
func NewKeyInputFromEvent(e *sdl.KeyboardEvent) *KeyInput {
	return &KeyInput{Key: rune(sdl.GetKeyFromScancode(e.Keysym.Scancode)), Type: e.Type}
}
