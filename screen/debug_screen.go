package screen

import (
	"fmt"

	"github.com/ariejan/i6502"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const (
	textCols = 80
	textRows = 26
)

// EmulatorInterface provides a de-coupled interface
// to the emulator to avoid circular package dependencies
type EmulatorInterface interface {
	GetCPU() *i6502.Cpu
	ReadMemory(address uint16) uint8
}

// DebugScreen is the output for the debug info
type DebugScreen struct {
	parent   *Screen
	window   *sdl.Window
	renderer *sdl.Renderer
	font     *ttf.Font

	fontmetrics *ttf.GlyphMetrics
	charWidth   int32
	charHeight  int32

	screenWidth  int32
	screenHeight int32

	em EmulatorInterface

	lastPC           uint16
	lastDebugTexture *sdl.Texture

	Active bool
}

// NewDebugScreen creates a new debug window
func NewDebugScreen(parent *Screen) *DebugScreen {
	result := &DebugScreen{parent: parent, Active: false}
	result.createDebugWindow()
	result.lastPC = 0
	result.lastDebugTexture = nil
	return result
}

func (s *DebugScreen) initializeFonts() error {
	font, fontmetrics, err := loadFont("UbuntuMono-B.ttf")
	if err != nil {
		return fmt.Errorf("Error loading font UbuntuMono-B: %v", err)
	}
	s.font = font
	s.fontmetrics = fontmetrics

	s.charWidth, s.charHeight = getCharacterMetrics(s.fontmetrics)

	return nil
}

func (s *DebugScreen) createDebugWindow() error {
	if err := s.initializeFonts(); err != nil {
		panic(err)
	}

	s.screenHeight = int32((s.fontmetrics.MaxY + s.fontmetrics.Advance) * textRows)
	s.screenWidth = int32((s.fontmetrics.MaxX) * textCols)

	x, y := s.parent.GetPosition()

	window, err := sdl.CreateWindow(
		"Debug Monitor",
		x-s.screenWidth,
		y,
		s.screenWidth,
		s.screenHeight,
		sdl.WINDOW_ALLOW_HIGHDPI|sdl.WINDOW_HIDDEN,
	)
	if err != nil {
		return err
	}
	s.window = window

	renderer, err := sdl.CreateRenderer(s.window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		return err
	}
	s.renderer = renderer

	// This is a kludge.  I force it to load all the printable ASCII
	// characters in the font, and this seems to cause it to print
	// on the debug screen correctly.  Without this, I get all sorts
	// weird issues
	s.initializeSymbols()

	return nil
}

func (s *DebugScreen) displayDebugInfo(renderer *sdl.Renderer) error {
	if !s.Active {
		return nil
	}

	c := s.em.GetCPU()

	if s.lastDebugTexture == nil || s.lastPC != s.em.GetCPU().PC {
		instr := s.em.ReadMemory(c.PC)
		msg := fmt.Sprintf("Addr:%04X    A:%02X    X:%02X    Y:%02X    SP:%02X    FLAGS:%02X    Instr:%02X", c.PC, c.A, c.X, c.Y, c.SP, c.P, instr)
		texture, err := createTexture(msg, s.parent.foreground, s.font, renderer)
		if err != nil {
			return fmt.Errorf("Creating debug texture: %v", err)
		} else if texture == nil {
			return nil
		}
		s.lastDebugTexture = texture
		s.lastPC = s.em.GetCPU().PC
	}

	_, _, w, h, err := s.lastDebugTexture.Query()
	if err != nil {
		return fmt.Errorf("Unable to query texture: %v", err)
	}

	renderer.Copy(
		s.lastDebugTexture,
		&sdl.Rect{X: 0, Y: 0, W: w, H: h},
		&sdl.Rect{X: 20, Y: 20 + (h / 2), W: w, H: h},
	)

	return nil
}

// DrawScreen draws the entire screen
func (s *DebugScreen) DrawScreen() {
	renderer, err := s.window.GetRenderer()
	if err != nil {
		panic(err)
	}

	renderer.SetDrawColor(s.parent.background.R, s.parent.background.G, s.parent.background.B, s.parent.background.A)
	renderer.Clear()

	if err := s.displayDebugInfo(renderer); err != nil {
		panic(err)
	}

	renderer.Present()
}

// CleanUp cleans all resources created by the debug window
func (s *DebugScreen) CleanUp() {
	if s.window == nil {
		return
	}

	if s.lastDebugTexture != nil {
		s.lastDebugTexture.Destroy()
	}

	s.font.Close()
	if err := s.renderer.Destroy(); err != nil {
		panic(err)
	}
	if err := s.window.Destroy(); err != nil {
		panic(err)
	}
	s.window = nil
}

// Show shows the debug window
func (s *DebugScreen) Show(em EmulatorInterface) {
	if !s.Active {
		s.Active = true
		s.em = em
		s.window.Show()
	}
}

// Hide hides the debug window
func (s *DebugScreen) Hide() {
	if s.Active {
		s.Active = false
		s.window.Hide()
	}
}

func (s *DebugScreen) initializeSymbols() error {
	for x := 32; x < 127; x++ {
		t, err := createTexture(string(rune(x)), s.parent.foreground, s.font, s.renderer)
		if err != nil {
			return err
		}
		t.Destroy()
	}

	return nil
}
