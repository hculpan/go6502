package main

import (
	"fmt"

	"github.com/hculpan/go6502/emulator"
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

	em := emulator.NewEmulator()
	var i uint
	for i = 0; i < 64*1024; i++ {
		em.WriteMemory(uint16(i), 0xEA)
	}
	em.WriteMemory(0xFFFD, 0x02)
	em.WriteMemory(0xFFFC, 0x00)

	em.WriteMemory(0x0200, 0xAD)
	em.WriteMemory(0x0201, 0x0C)
	em.WriteMemory(0x0202, 0x02)
	em.WriteMemory(0x0203, 0x20)
	em.WriteMemory(0x0204, 0x0E)
	em.WriteMemory(0x0205, 0x02)
	em.WriteMemory(0x0206, 0x8D)
	em.WriteMemory(0x0207, 0x0C)
	em.WriteMemory(0x0208, 0x02)
	em.WriteMemory(0x0209, 0x4C)
	em.WriteMemory(0x020A, 0x00)
	em.WriteMemory(0x020B, 0x02)
	em.WriteMemory(0x020C, 0x01)
	em.WriteMemory(0x020D, 0x11)
	em.WriteMemory(0x020E, 0x6D)
	em.WriteMemory(0x020F, 0x0D)
	em.WriteMemory(0x0210, 0x02)
	em.WriteMemory(0x0211, 0x60)
	defer func() {
		em.Terminate()
	}()

	var lastAddress uint16 = 0
	for {
		address := em.CPU.PC
		if address != lastAddress {
			data := em.ReadMemory(address)
			s := fmt.Sprintf("%04X:%02X   A:%02X    X:%02X    Y:%02X    SP:%02X\r", address, data, em.CPU.A, em.CPU.X, em.CPU.Y, em.CPU.SP)
			for _, c := range s {
				scr.ProcessRune(c)
			}
			lastAddress = address
		}
		event := sdl.PollEvent()
		if event != nil {
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
					case sdl.K_F2:
						em.StartEmulator(10) // 1 millisecond/step
					case sdl.K_F3:
						em.CPU.Reset()
					case sdl.K_F4:
						em.EnableSingleStep()
					case sdl.K_F5:
						em.DisableSingleStep()
					case sdl.K_F6:
						em.NextStep()
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
