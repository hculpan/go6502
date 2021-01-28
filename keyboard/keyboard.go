package keyboard

import (
	"strings"

	"github.com/veandco/go-sdl2/sdl"
)

// Defines keyboard runes
const (
	LeftShift  = 1073742049
	RightShift = 1073742053
	CapsLock   = 1073741881
	Enter      = 13
	Backspace  = 8
	Escape     = 27
)

// Keyboard represents a physical keyboard
type Keyboard struct {
	capsLock bool
	shiftOn  bool
}

// NewKeyboard creates a new keyboard
func NewKeyboard() *Keyboard {
	return &Keyboard{capsLock: false, shiftOn: false}
}

// ProcessKeyInput processes a keyboard event
func (k *Keyboard) ProcessKeyInput(i *KeyInput) rune {
	r := i.Key
	if i.Type == sdl.KEYDOWN {
		switch {
		case r == Backspace:
			return r
		case r == Escape:
			return r
		case r == Enter:
			return r
		case r == CapsLock:
			k.capsLock = !k.capsLock
		case r == LeftShift || r == RightShift:
			k.shiftOn = true
		case r < 32 || r > 126:
			// Nothing
		default:
			return k.runeCaps(r)
		}
	} else if i.Type == sdl.KEYUP && (r == LeftShift || r == RightShift) { // left/right shift
		k.shiftOn = false
	}

	return 0
}

func (k *Keyboard) runeCaps(r rune) rune {
	if (r > 96 && r < 123) && (k.capsLock || k.shiftOn) && !(k.capsLock && k.shiftOn) {
		return rune(strings.ToUpper(string(r))[0])
	} else if k.shiftOn {
		switch r {
		case '0':
			return ')'
		case '1':
			return '!'
		case '2':
			return '@'
		case '3':
			return '#'
		case '4':
			return '$'
		case '5':
			return '%'
		case '6':
			return '^'
		case '7':
			return '&'
		case '8':
			return '*'
		case '9':
			return '('
		case '-':
			return '_'
		case '=':
			return '+'
		case '\\':
			return '|'
		case '[':
			return '{'
		case ']':
			return '}'
		case ';':
			return ':'
		case '\'':
			return '"'
		case ',':
			return '<'
		case '.':
			return '>'
		case '/':
			return '?'
		case '`':
			return '~'
		}
	}

	return r
}
