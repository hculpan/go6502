package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/hculpan/go6502/emulator"
	"github.com/hculpan/go6502/keyboard"
	"github.com/hculpan/go6502/resources"
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

// EventResult is the event coming back from the event handler
type EventResult int

const (
	eventResultNone = iota
	eventResultQuit
)

var status *screen.ComputerStatus

func handleEvent(event sdl.Event, em *emulator.Emulator, scr *screen.Screen, k *keyboard.Keyboard) EventResult {
	if event != nil {
		switch event.(type) {
		case *sdl.QuitEvent:
			ok := dialog.Message("Do you wish to exit?").Title("Exit go6502").YesNo()
			if ok {
				em.Terminate()
				return eventResultQuit
			}
		case *sdl.MouseButtonEvent:
			me := event.(*sdl.MouseButtonEvent)
			if me.Type == sdl.MOUSEBUTTONDOWN && me.Button == sdl.BUTTON_LEFT {
				if scr.IsEmulatorOnOffClicked(me.X, me.Y) {
					toggleEmulatorOnOff(em, scr)
				}
			}
		case *sdl.WindowEvent:
			we := event.(*sdl.WindowEvent)
			if we.Data1 != 0 && we.Data2 != 0 {
				fmt.Println("Window moved!")
				fmt.Printf("Window position: %d, %d\n", we.Data1, we.Data2)
				mode, err := sdl.GetDesktopDisplayMode(0)
				if err != nil {
					panic(err)
				}
				fmt.Printf("Desktop size: %d, %d\n", mode.W, mode.H)
			}
		case *sdl.KeyboardEvent:
			keycode := sdl.GetKeyFromScancode(event.(*sdl.KeyboardEvent).Keysym.Scancode)
			if event.(*sdl.KeyboardEvent).Type == sdl.KEYDOWN {
				switch keycode {
				case sdl.K_F2:
					toggleEmulatorOnOff(em, scr)
					scr.UpdateScreen()
				case sdl.K_F3:
					if !status.Running {
						emulatorOnWithStep(em)
						scr.EnableDebug(em)
					}
					scr.UpdateScreen()
				case sdl.K_F5:
					if status.Running && !status.SingleStep {
						em.EnableSingleStep()
						scr.EnableDebug(em)
					}
					scr.UpdateScreen()
				case sdl.K_F6:
					if status.Running && status.SingleStep {
						em.NextStep()
						time.Sleep(100 * time.Millisecond) // Give emulator a little time to advance to next instruction
					}
					scr.UpdateScreen()
				case sdl.K_F7:
					if status.Running && status.SingleStep {
						em.DisableSingleStep()
						scr.DisableDebug()
					}
					scr.UpdateScreen()
				case sdl.K_F9:
					filename, err := dialog.File().Filter("TXT files", "txt").Filter("SBIN files", "sbin").Filter("BIN files", "bin").Title("Load ROM File").Load()
					if err != nil {
						if strings.Contains(err.Error(), "Cancelled") {
							break
						}
						dialog.Message(fmt.Sprintf("Unable to find %s: %s", filename, err)).Error()
						return eventResultNone
					}
					em.Terminate()
					em.CPU.Reset()
					err = loadRAM(filename, em, scr)
					if err != nil {
						dialog.Message(fmt.Sprintf("Unable to load %s: %s", filename, err)).Error()
					}
					scr.Reset()
					status.Running = false
					status.SingleStep = false
					em.DisableSingleStep()
					scr.UpdateScreen()
				case sdl.K_ESCAPE:
					ok := dialog.Message("Do you wish to exit?").Title("Exit go6502").YesNo()
					if ok {
						em.Terminate()
						return eventResultQuit
					}
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

	return eventResultNone
}

func main() {
	status = screen.NewComputerStatus()
	scr := screen.NewScreen(textCols, textRows, status)
	if err := scr.Show(); err != nil {
		fmt.Println(err)
		return
	}
	defer scr.CleanUp()
	k := keyboard.NewKeyboard()

	em := emulator.NewEmulator(scr)

	if err := loadROMBin(em, scr); err != nil {
		fmt.Println("Failed to load rom:", err)
		return
	}

	defer func() {
		em.Terminate()
	}()

	for {
		eventResult := handleEvent(sdl.PollEvent(), em, scr, k)
		switch eventResult {
		case eventResultQuit:
			return
		default:
			if status.Running {
				em.Step()
			}
			scr.DrawScreen()
		}
	}
}

func toggleEmulatorOnOff(em *emulator.Emulator, scr *screen.Screen) {
	if status.Running {
		emulatorOff(em, scr)
	} else {
		emulatorOn(em)
	}
}

func emulatorOn(em *emulator.Emulator) {
	if !status.Running {
		em.CPU.Reset()
		em.DisableSingleStep()
		status.Running = true
		status.SingleStep = false
		em.StartEmulator()
	}
}

func emulatorOnWithStep(em *emulator.Emulator) {
	if !status.Running {
		em.CPU.Reset()
		em.EnableSingleStep()
		status.Running = true
		status.SingleStep = true
		em.StartEmulator()
	}
}

func emulatorOff(em *emulator.Emulator, scr *screen.Screen) {
	if status.Running {
		em.Terminate()
		em.CPU.Reset()
		em.DisableSingleStep()
		scr.Reset()
		status.Running = false
		status.SingleStep = false
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

func loadROMBin(em *emulator.Emulator, scr *screen.Screen) error {
	// Read rom file
	rom, err := resources.Asset("resources/rom.bin")
	if err != nil {
		return err
	}
	if len(rom) > 65024 {
		return fmt.Errorf("Rom file too large, must be 65024 with origin at 0x0200")
	} else if len(rom) < 65024 {
		return fmt.Errorf("Rom file too small, must be 65024 with origin at 0x0200")
	}

	resetRAM(em)
	// Write to memory
	for i := 0; i < 65024; i++ {
		idx := i + 512
		writeToEmulatorMemory(em, uint16(idx), rom[i])
	}

	status.RomFilename = "rom.bin"
	return nil
}

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

	resetRAM(em)
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

	resetRAM(em)
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

	resetRAM(em)
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

func resetRAM(em *emulator.Emulator) {
	for x := 0; x < 65536; x++ {
		writeToEmulatorMemory(em, uint16(x), 0)
	}
}

func loadRAM(f string, em *emulator.Emulator, scr *screen.Screen) error {
	status.RomFilename = ""
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

	status.RomFilename = f
	return nil
}

func writeToEmulatorMemory(em *emulator.Emulator, address uint16, data byte) {
	if address < 0x8000 || address > 0x83FF {
		em.WriteMemory(address, data)
	}
}
