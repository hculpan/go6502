package main

import (
	"fmt"

	"github.com/hculpan/go6502/keyboard"
	"github.com/hculpan/go6502/screen"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	textCols = 80
	textRows = 26
)

const (
	down = sdl.KEYDOWN
	up   = sdl.KEYUP
)

func main() {
	scr := screen.NewScreen(textCols, textRows)
	if err := scr.Show(); err != nil {
		fmt.Println(err)
		return
	}
	defer scr.CleanUp()
	k := keyboard.NewKeyboard()

	sendDownKey(keyboard.LeftShift, scr, k)
	sendDownUpKey('h', scr, k)
	sendUpKey(keyboard.LeftShift, scr, k)
	sendDownUpKey('e', scr, k)
	sendDownUpKey('l', scr, k)
	sendDownUpKey('l', scr, k)
	sendDownUpKey('o', scr, k)
	sendDownUpKey(keyboard.Enter, scr, k)
	sendDownKey(keyboard.RightShift, scr, k)
	sendDownUpKey('h', scr, k)
	sendUpKey(keyboard.RightShift, scr, k)
	sendDownUpKey('e', scr, k)
	sendDownUpKey('l', scr, k)
	sendDownUpKey('l', scr, k)
	sendDownUpKey('o', scr, k)
	for i := 0; i < 50; i++ {
		sendDownUpKey(keyboard.Backspace, scr, k)
	}
	sendDownKey(keyboard.LeftShift, scr, k)
	sendDownUpKey('g', scr, k)
	sendUpKey(keyboard.LeftShift, scr, k)
	sendDownUpKey('o', scr, k)
	sendDownUpKey('o', scr, k)
	sendDownUpKey('d', scr, k)
	sendDownUpKey('b', scr, k)
	sendDownUpKey('y', scr, k)
	sendDownUpKey('e', scr, k)
	sendDownUpKey(keyboard.Enter, scr, k)
	sendDownKey(keyboard.RightShift, scr, k)
	sendDownUpKey('t', scr, k)
	sendUpKey(keyboard.RightShift, scr, k)
	sendDownUpKey('h', scr, k)
	sendDownUpKey('i', scr, k)
	sendDownUpKey('s', scr, k)
	sendDownUpKey(' ', scr, k)
	sendDownUpKey('i', scr, k)
	sendDownUpKey('s', scr, k)
	sendDownUpKey(' ', scr, k)
	sendDownUpKey('a', scr, k)
	sendDownUpKey(' ', scr, k)
	sendDownUpKey('t', scr, k)
	sendDownUpKey('e', scr, k)
	sendDownUpKey('s', scr, k)
	sendDownUpKey('t', scr, k)
	sendDownKey(keyboard.RightShift, scr, k)
	sendDownUpKey('1', scr, k)
	sendUpKey(keyboard.RightShift, scr, k)
	sendDownUpKey(keyboard.Enter, scr, k)

	for {
		event := sdl.WaitEvent()
		switch event.(type) {
		case *sdl.QuitEvent:
			return
		case *sdl.KeyboardEvent:
			keycode := sdl.GetKeyFromScancode(event.(*sdl.KeyboardEvent).Keysym.Scancode)
			if event.(*sdl.KeyboardEvent).Type == sdl.KEYDOWN {
				switch keycode {
				case sdl.K_HOME:
					sendEscSequence("[H", scr)
				case sdl.K_INSERT:
					sendEscSequence("[2J", scr)
				case sdl.K_UP:
					sendEscSequence("[1A", scr)
				case sdl.K_DOWN:
					sendEscSequence("[1B", scr)
				case sdl.K_RIGHT:
					sendEscSequence("[1C", scr)
				case sdl.K_LEFT:
					sendEscSequence("[1D", scr)
				case sdl.K_END:
					sendEscSequence("[40;13H", scr)
				case sdl.K_ESCAPE:
					return
				default:
					r := k.ProcessKeyInput(keyboard.NewKeyInputFromEvent(event.(*sdl.KeyboardEvent)))
					if r != 0 {
						scr.ProcessRune(r)
					}
				}
			} else {
				r := k.ProcessKeyInput(keyboard.NewKeyInputFromEvent(event.(*sdl.KeyboardEvent)))
				if r != 0 {
					scr.ProcessRune(r)
				}
			}
		}

	}
}

func buildKey(r rune, keyAction uint32) *keyboard.KeyInput {
	return &keyboard.KeyInput{Key: r, Type: keyAction}
}

func sendDownUpKey(r rune, scr *screen.Screen, k *keyboard.Keyboard) {
	newrune := k.ProcessKeyInput(buildKey(r, down))
	scr.ProcessRune(newrune)
	newrune = k.ProcessKeyInput(buildKey(r, up))
	scr.ProcessRune(newrune)
}

func sendDownKey(r rune, scr *screen.Screen, k *keyboard.Keyboard) {
	newrune := k.ProcessKeyInput(buildKey(r, down))
	scr.ProcessRune(newrune)
}

func sendUpKey(r rune, scr *screen.Screen, k *keyboard.Keyboard) {
	newrune := k.ProcessKeyInput(buildKey(r, up))
	scr.ProcessRune(newrune)
}

func sendEscSequence(s string, scr *screen.Screen) {
	scr.ProcessRune(keyboard.Escape)
	for _, v := range s {
		scr.ProcessRune(v)
	}
}
