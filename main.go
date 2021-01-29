package main

import (
	"fmt"
	"io/ioutil"

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

func loadRAM(em *emulator.Emulator) {
	// Read rom file
	rom, err := ioutil.ReadFile("rom.bin")
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(rom) > 65024 {
		fmt.Println("Rom file too large, must be 65024 with origin at 0x0200")
		return
	} else if len(rom) < 65024 {
		fmt.Println("Rom file too small, must be 65024 with origin at 0x0200")
		return
	}

	// Write to memory except
	for i := 0; i < 65024; i++ {
		idx := i + 512
		if idx < 0x8000 || idx > 0x83FF {
			em.WriteMemory(uint16(idx), rom[i])
		}
	}

}

func main() {
	scr := screen.NewScreen(textCols, textRows)
	if err := scr.Show(); err != nil {
		fmt.Println(err)
		return
	}
	defer scr.CleanUp()
	k := keyboard.NewKeyboard()

	em := emulator.NewEmulator(scr)

	loadRAM(em)

	defer func() {
		em.Terminate()
	}()

	//	var lastAddress uint16 = 0
	for {
		/*		if em.Active {
				address := em.CPU.PC
				if address != lastAddress {
					data := em.ReadMemory(address)
					s := fmt.Sprintf("%04X:%02X   A:%02X    X:%02X    Y:%02X    SP:%02X\r", address, data, em.CPU.A, em.CPU.X, em.CPU.Y, em.CPU.SP)
					for _, c := range s {
						scr.ProcessRune(c)
					}
					lastAddress = address
				}
			}*/
		event := sdl.PollEvent()
		if event != nil {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			case *sdl.KeyboardEvent:
				keycode := sdl.GetKeyFromScancode(event.(*sdl.KeyboardEvent).Keysym.Scancode)
				if event.(*sdl.KeyboardEvent).Type == sdl.KEYDOWN {
					switch keycode {
					case sdl.K_F2:
						em.StartEmulator()
					case sdl.K_F3:
						em.Terminate()
						em.CPU.Reset()
						scr.Reset()
					case sdl.K_F5:
						em.EnableSingleStep()
					case sdl.K_F6:
						em.NextStep()
					case sdl.K_F7:
						em.DisableSingleStep()
					case sdl.K_F9:
						em.Terminate()
						em.CPU.Reset()
						loadRAM(em)
						scr.Reset()
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
