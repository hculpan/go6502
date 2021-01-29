package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hculpan/go6502/emulator"
	"github.com/hculpan/go6502/keyboard"
	"github.com/hculpan/go6502/screen"
	"github.com/sqweek/dialog"
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

func loadBIN(f string, em *emulator.Emulator) error {
	// Read rom file
	rom, err := ioutil.ReadFile(f)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if len(rom) > 65024 {
		return fmt.Errorf("Rom file too large, must be 65024 with origin at 0x0200")
	} else if len(rom) < 65024 {
		return fmt.Errorf("Rom file too small, must be 65024 with origin at 0x0200")
	}

	// Write to memory
	for i := 0; i < 65024; i++ {
		idx := i + 512
		writeToEmulatorMemory(em, uint16(idx), rom[i])
	}

	return nil
}

func loadSBIN(f string, em *emulator.Emulator) error {
	// Read rom file
	rom, err := ioutil.ReadFile(f)
	if err != nil {
		fmt.Println(err)
		return err
	}

	org := int((uint16(rom[1]) << 8) + uint16(rom[0]))
	fmt.Printf("org=%04X", org)

	// Write to memory
	for i := 4; i < len(rom); i++ {
		addr := i + org - 4
		writeToEmulatorMemory(em, uint16(addr), rom[i])
	}

	return nil
}

func loadTXT(f string, em *emulator.Emulator) error {
	file, err := os.Open(f)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.Trim(scanner.Text(), " \n\r\t")
		if len(line) > 0 {
			addrStr := line[:4]
			addr, err := strconv.ParseInt(addrStr, 16, 32)
			if err != nil {
				return fmt.Errorf("Parsing address '%s': %s", addrStr, err)
			}
			data := strings.Split(line[5:], " ")
			for i, v := range data {
				d, err := strconv.ParseInt(v, 16, 16)
				if err != nil {
					return fmt.Errorf("Parsing data '%s': %s", v, err)
				}
				writeToEmulatorMemory(em, uint16(int(addr)+i), byte(d))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func loadRAM(f string, em *emulator.Emulator, scr *screen.Screen) error {
	scr.SetRomName("")
	switch filepath.Ext(f) {
	case ".bin":
		if err := loadBIN(f, em); err != nil {
			return err
		}
	case ".sbin":
		if err := loadSBIN(f, em); err != nil {
			return err
		}
	case ".txt":
		if err := loadTXT(f, em); err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unrecognized file type: %s", filepath.Ext(f))
	}

	scr.SetRomName(f)
	return nil
}

func writeToEmulatorMemory(em *emulator.Emulator, address uint16, data byte) {
	if address < 0x8000 || address > 0x83FF {
		em.WriteMemory(address, data)
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

	if err := loadRAM("rom.bin", em, scr); err != nil {
		fmt.Println("Failed to load rom:", err)
		return
	}

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
						filename, err := dialog.File().Filter("BIN files", "bin").Filter("SBIN files", "sbin").Filter("TXT files", "txt").Title("Load ROM").Load()
						if err != nil {
							dialog.Message(fmt.Sprintf("Unable to find %s: %s", filename, err)).Error()
							return
						}
						em.Terminate()
						em.CPU.Reset()
						err = loadRAM(filename, em, scr)
						if err != nil {
							dialog.Message(fmt.Sprintf("Unable to load %s: %s", filename, err)).Error()
						}
						scr.Reset()
					case sdl.K_ESCAPE:
						return
					default:
						r := k.ProcessKeyInput(keyboard.NewKeyInputFromEvent(event.(*sdl.KeyboardEvent)))
						if r != 0 {
							em.SetKeyWaiting(r)
						}
					}
				} else {
					r := k.ProcessKeyInput(keyboard.NewKeyInputFromEvent(event.(*sdl.KeyboardEvent)))
					if r != 0 {
						em.SetKeyWaiting(r)
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
